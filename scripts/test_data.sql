-- Test data for Instagram User Processor

-- Insert sample users
INSERT INTO instagram_users (id, username, full_name, biography, is_verified, is_business_account, followers, following, posts) VALUES
('1001', 'musiclover2024', 'Alex Johnson', 'Music producer and DJ 🎵 Based in LA', true, true, 125000, 850, 342),
('1002', 'techguru_sarah', 'Sarah Chen', 'Software engineer at @techcorp Building cool stuff!', false, false, 8500, 1200, 156),
('1003', 'foodie_adventures', 'Mike Rodriguez', 'Food blogger | Trying every restaurant in NYC 🍕', true, true, 45000, 2300, 1240),
('1004', 'fitness_jenny', 'Jenny Williams', 'Certified personal trainer 💪 DM for workout plans', false, true, 18000, 450, 890),
('1005', 'artist_david', 'David Kim', 'Digital artist and illustrator ✨ Commissions open', false, false, 3400, 890, 234),
('1006', 'travel_emma', 'Emma Thompson', 'Travel photographer 📸 40 countries and counting', true, false, 78000, 1100, 567),
('1007', 'comedian_joe', 'Joe Martinez', 'Stand-up comedian | Shows every weekend 😂', false, true, 12000, 340, 445),
('1008', 'fashion_lisa', 'Lisa Anderson', 'Fashion stylist | Personal shopping available 👗', true, true, 156000, 2100, 678),
('1009', 'gamer_alex', 'Alex Turner', 'Pro gamer | Streaming daily on Twitch 🎮', false, false, 22000, 560, 123),
('1010', 'chef_maria', 'Maria Gonzalez', 'Head chef @finedining | Cookbook coming soon 👩‍🍳', true, true, 34000, 890, 445);

-- Insert sample posts for these users
INSERT INTO instagram_posts (id, user_id, username, caption, like_count, comment_count, play_count, is_ad, posted_at) VALUES
-- Posts for musiclover2024
('post_1001_1', '1001', 'musiclover2024', 'New track dropping this Friday! Can\'t wait to share it with you all 🎵', 4500, 234, 12000, false, NOW() - INTERVAL '1 day'),
('post_1001_2', '1001', 'musiclover2024', 'Studio session vibes ✨ #NewMusic #Producer', 3200, 156, 8900, false, NOW() - INTERVAL '3 days'),
('post_1001_3', '1001', 'musiclover2024', 'Thanks to @soundgear for the amazing equipment! #ad #sponsored', 2800, 89, 5600, true, NOW() - INTERVAL '5 days'),

-- Posts for techguru_sarah
('post_1002_1', '1002', 'techguru_sarah', 'Just deployed my first ML model to production! 🚀', 890, 67, null, false, NOW() - INTERVAL '2 days'),
('post_1002_2', '1002', 'techguru_sarah', 'Code review best practices - thread 🧵', 654, 45, null, false, NOW() - INTERVAL '1 week'),

-- Posts for foodie_adventures
('post_1003_1', '1003', 'foodie_adventures', 'This pizza at @nycpizzaplace is INCREDIBLE! 🍕', 2300, 145, null, false, NOW() - INTERVAL '1 day'),
('post_1003_2', '1003', 'foodie_adventures', 'Top 10 brunch spots in Manhattan - swipe to see! ➡️', 3400, 234, null, false, NOW() - INTERVAL '4 days'),
('post_1003_3', '1003', 'foodie_adventures', 'Thanks to @restaurantapp for sponsoring this review! #ad', 1800, 78, null, true, NOW() - INTERVAL '1 week'),

-- Posts for fitness_jenny
('post_1004_1', '1004', 'fitness_jenny', '30-minute HIIT workout - no equipment needed! 💪', 1200, 89, 45000, false, NOW() - INTERVAL '1 day'),
('post_1004_2', '1004', 'fitness_jenny', 'Morning motivation: Your only competition is who you were yesterday', 890, 67, null, false, NOW() - INTERVAL '2 days'),

-- Posts for artist_david
('post_1005_1', '1005', 'artist_david', 'New digital artwork finished! What do you think? 🎨', 567, 34, null, false, NOW() - INTERVAL '3 days'),
('post_1005_2', '1005', 'artist_david', 'Time-lapse of my drawing process ✨', 678, 45, 2300, false, NOW() - INTERVAL '1 week'),

-- Posts for travel_emma
('post_1006_1', '1006', 'travel_emma', 'Sunrise over the Himalayas 🏔️ #Nepal #Photography', 5600, 345, null, false, NOW() - INTERVAL '2 days'),
('post_1006_2', '1006', 'travel_emma', 'Travel gear essentials - link in bio 📸', 3400, 167, null, false, NOW() - INTERVAL '5 days'),

-- Posts for comedian_joe
('post_1007_1', '1007', 'comedian_joe', 'When you realize it\'s Monday tomorrow 😂', 890, 123, 12000, false, NOW() - INTERVAL '1 day'),
('post_1007_2', '1007', 'comedian_joe', 'Show tonight at 8 PM! Come laugh with us 🎭', 654, 45, null, false, NOW() - INTERVAL '3 days'),

-- Posts for fashion_lisa
('post_1008_1', '1008', 'fashion_lisa', 'Fall fashion trends you need to know! 👗', 4500, 234, null, false, NOW() - INTERVAL '1 day'),
('post_1008_2', '1008', 'fashion_lisa', 'Styling tips for petite women ✨', 3200, 178, null, false, NOW() - INTERVAL '4 days'),
('post_1008_3', '1008', 'fashion_lisa', 'Thanks to @fashionbrand for this gorgeous dress! #gifted #ad', 2800, 145, null, true, NOW() - INTERVAL '6 days'),

-- Posts for gamer_alex
('post_1009_1', '1009', 'gamer_alex', 'Just hit Grandmaster! 🎮 Road to pro continues', 1200, 89, 34000, false, NOW() - INTERVAL '1 day'),
('post_1009_2', '1009', 'gamer_alex', 'New gaming setup reveal! Specs in comments 💻', 890, 67, 23000, false, NOW() - INTERVAL '5 days'),

-- Posts for chef_maria
('post_1010_1', '1010', 'chef_maria', 'Tonight\'s special: Pan-seared salmon with truffle risotto 🍽️', 2300, 134, null, false, NOW() - INTERVAL '1 day'),
('post_1010_2', '1010', 'chef_maria', 'Behind the scenes at @finedining kitchen 👩‍🍳', 1800, 98, 15000, false, NOW() - INTERVAL '3 days');

-- Insert sample assets (for complex query testing)
INSERT INTO instagram_assets (id, post_id, asset_type, url, tagged_user_usernames) VALUES
('asset_1', 'post_1001_1', 'video', 'https://example.com/video1.mp4', ARRAY['techguru_sarah', 'artist_david']),
('asset_2', 'post_1003_1', 'image', 'https://example.com/image1.jpg', ARRAY['chef_maria']),
('asset_3', 'post_1004_1', 'video', 'https://example.com/video2.mp4', ARRAY['travel_emma', 'fashion_lisa']),
('asset_4', 'post_1006_1', 'image', 'https://example.com/image2.jpg', ARRAY[]),
('asset_5', 'post_1008_1', 'carousel', 'https://example.com/carousel1/', ARRAY['foodie_adventures', 'fitness_jenny']);

-- Insert a sample processing job
INSERT INTO processing_jobs (id, status, total_users, processed_users, successful_users, failed_users, max_concurrency, started_at, completed_at, errors) VALUES
(uuid_generate_v4(), 'completed', 5, 5, 4, 1, 3, NOW() - INTERVAL '1 hour', NOW() - INTERVAL '45 minutes', '{"user_not_found": "User @nonexistent not found"}');

-- Create some additional users for testing batch processing
INSERT INTO instagram_users (id, username, full_name, biography, is_verified, followers, following, posts) VALUES
('2001', 'test_user_1', 'Test User One', 'This is a test user for batch processing', false, 1000, 500, 50),
('2002', 'test_user_2', 'Test User Two', 'Another test user', false, 2000, 750, 75),
('2003', 'test_user_3', 'Test User Three', 'Third test user', true, 5000, 1000, 100),
('2004', 'test_user_4', 'Test User Four', 'Fourth test user', false, 800, 300, 25),
('2005', 'test_user_5', 'Test User Five', 'Fifth test user', false, 1500, 600, 80);

-- Update the scraped_at timestamps to show some variation
UPDATE instagram_users SET scraped_at = NOW() - INTERVAL '2 hours' WHERE id IN ('1001', '1002');
UPDATE instagram_users SET scraped_at = NOW() - INTERVAL '1 day' WHERE id IN ('1003', '1004');
UPDATE instagram_users SET scraped_at = NOW() - INTERVAL '3 days' WHERE id IN ('1005', '1006');
UPDATE instagram_users SET scraped_at = NOW() - INTERVAL '1 week' WHERE id IN ('1007', '1008');

-- Add some comments to demonstrate the database is working
SELECT 'Sample data inserted successfully!' as message;
SELECT COUNT(*) as total_users FROM instagram_users;
SELECT COUNT(*) as total_posts FROM instagram_posts;
SELECT COUNT(*) as total_assets FROM instagram_assets;