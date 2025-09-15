-- Instagram User Processor Database Schema

-- Enable necessary extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create instagram_users table
CREATE TABLE IF NOT EXISTS instagram_users (
    id VARCHAR(50) PRIMARY KEY,
    username VARCHAR(100) UNIQUE NOT NULL,
    full_name TEXT,
    biography TEXT,
    is_verified BOOLEAN DEFAULT FALSE,
    is_business_account BOOLEAN DEFAULT FALSE,
    is_professional_account BOOLEAN DEFAULT FALSE,
    is_private BOOLEAN DEFAULT FALSE,
    category_name VARCHAR(100),
    followers BIGINT DEFAULT 0,
    following BIGINT DEFAULT 0,
    posts BIGINT DEFAULT 0,
    scraped_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create instagram_posts table (for complex query demonstrations)
CREATE TABLE IF NOT EXISTS instagram_posts (
    id VARCHAR(50) PRIMARY KEY,
    user_id VARCHAR(50) NOT NULL REFERENCES instagram_users(id),
    username VARCHAR(100) NOT NULL,
    caption TEXT,
    like_count BIGINT DEFAULT 0,
    comment_count BIGINT DEFAULT 0,
    play_count BIGINT,
    is_ad BOOLEAN DEFAULT FALSE,
    posted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create instagram_assets table (for tagging features)
CREATE TABLE IF NOT EXISTS instagram_assets (
    id VARCHAR(50) PRIMARY KEY,
    post_id VARCHAR(50) NOT NULL REFERENCES instagram_posts(id),
    asset_type VARCHAR(20) NOT NULL CHECK (asset_type IN ('image', 'video', 'carousel')),
    url TEXT NOT NULL,
    tagged_user_usernames TEXT[],
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create processing_jobs table (for batch processing tracking)
CREATE TABLE IF NOT EXISTS processing_jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'running', 'completed', 'failed', 'cancelled')),
    total_users INTEGER NOT NULL DEFAULT 0,
    processed_users INTEGER NOT NULL DEFAULT 0,
    successful_users INTEGER NOT NULL DEFAULT 0,
    failed_users INTEGER NOT NULL DEFAULT 0,
    max_concurrency INTEGER NOT NULL DEFAULT 5,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    errors JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_instagram_users_username ON instagram_users(username);
CREATE INDEX IF NOT EXISTS idx_instagram_users_scraped_at ON instagram_users(scraped_at);
CREATE INDEX IF NOT EXISTS idx_instagram_users_followers ON instagram_users(followers DESC);
CREATE INDEX IF NOT EXISTS idx_instagram_users_verified ON instagram_users(is_verified) WHERE is_verified = TRUE;
CREATE INDEX IF NOT EXISTS idx_instagram_users_business ON instagram_users(is_business_account) WHERE is_business_account = TRUE;

CREATE INDEX IF NOT EXISTS idx_instagram_posts_user_id ON instagram_posts(user_id);
CREATE INDEX IF NOT EXISTS idx_instagram_posts_posted_at ON instagram_posts(posted_at DESC);
CREATE INDEX IF NOT EXISTS idx_instagram_posts_engagement ON instagram_posts(like_count DESC, comment_count DESC);

CREATE INDEX IF NOT EXISTS idx_instagram_assets_post_id ON instagram_assets(post_id);
CREATE INDEX IF NOT EXISTS idx_instagram_assets_tagged_users ON instagram_assets USING GIN(tagged_user_usernames);

CREATE INDEX IF NOT EXISTS idx_processing_jobs_status ON processing_jobs(status);
CREATE INDEX IF NOT EXISTS idx_processing_jobs_created_at ON processing_jobs(created_at DESC);

-- Create trigger to update updated_at column
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply the trigger to all tables
CREATE TRIGGER update_instagram_users_updated_at BEFORE UPDATE ON instagram_users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_instagram_posts_updated_at BEFORE UPDATE ON instagram_posts FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_instagram_assets_updated_at BEFORE UPDATE ON instagram_assets FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_processing_jobs_updated_at BEFORE UPDATE ON processing_jobs FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create a view for user statistics (used by the complex query)
CREATE OR REPLACE VIEW user_engagement_stats AS
SELECT
    u.id,
    u.username,
    u.followers,
    COUNT(p.id) as total_posts,
    AVG(COALESCE(p.like_count, 0) + COALESCE(p.comment_count, 0)) as avg_engagement,
    SUM(COALESCE(p.like_count, 0)) as total_likes,
    SUM(COALESCE(p.comment_count, 0)) as total_comments,
    MAX(p.posted_at) as last_post_date
FROM instagram_users u
LEFT JOIN instagram_posts p ON u.id = p.user_id
GROUP BY u.id, u.username, u.followers;

-- Sample function to demonstrate stored procedures
CREATE OR REPLACE FUNCTION get_top_users_by_engagement(limit_count INTEGER DEFAULT 10)
RETURNS TABLE (
    user_id VARCHAR(50),
    username VARCHAR(100),
    followers BIGINT,
    engagement_rate NUMERIC
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        u.id,
        u.username,
        u.followers,
        CASE
            WHEN u.followers > 0 THEN
                ROUND((AVG(COALESCE(p.like_count, 0) + COALESCE(p.comment_count, 0)) / u.followers * 100)::NUMERIC, 2)
            ELSE 0
        END as engagement_rate
    FROM instagram_users u
    LEFT JOIN instagram_posts p ON u.id = p.user_id
    WHERE u.followers > 1000  -- Filter out accounts with very few followers
    GROUP BY u.id, u.username, u.followers
    HAVING COUNT(p.id) > 5  -- Users with at least 5 posts
    ORDER BY engagement_rate DESC
    LIMIT limit_count;
END;
$$ LANGUAGE plpgsql;