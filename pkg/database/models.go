package database

import (
	"encoding/json"
	"time"
)

// User represents an Instagram user
type User struct {
	ID                    string    `json:"id" db:"id"`
	Username              string    `json:"username" db:"username"`
	FullName              string    `json:"full_name" db:"full_name"`
	Biography             string    `json:"biography" db:"biography"`
	IsVerified            bool      `json:"is_verified" db:"is_verified"`
	IsBusinessAccount     bool      `json:"is_business_account" db:"is_business_account"`
	IsProfessionalAccount bool      `json:"is_professional_account" db:"is_professional_account"`
	IsPrivate             bool      `json:"is_private" db:"is_private"`
	CategoryName          string    `json:"category_name" db:"category_name"`
	Followers             int64     `json:"followers" db:"followers"`
	Following             int64     `json:"following" db:"following"`
	Posts                 int64     `json:"posts" db:"posts"`
	ScrapedAt             time.Time `json:"scraped_at" db:"scraped_at"`
	CreatedAt             time.Time `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time `json:"updated_at" db:"updated_at"`
}

// UserStats represents detailed user statistics (from complex query)
type UserStats struct {
	User                 json.RawMessage `json:"user"`
	TaggedUsernames      json.RawMessage `json:"tagged_usernames"`
	CoauthoredUsernames  json.RawMessage `json:"coauthored_usernames"`
	TotalPostedCount     int             `json:"total_posted_count"`
	TotalTaggedInCount   int             `json:"total_tagged_in_count"`
	TotalCoauthoredCount int             `json:"total_coauthored_count"`
	EngagementRate       float64         `json:"engagement_rate"`
	AveragePostsPerWeek  float64         `json:"average_posts_per_week"`
}

// Post represents an Instagram post (simplified for demo)
type Post struct {
	ID          string    `json:"id" db:"id"`
	UserID      string    `json:"user_id" db:"user_id"`
	Username    string    `json:"username" db:"username"`
	Caption     string    `json:"caption" db:"caption"`
	LikeCount   int64     `json:"like_count" db:"like_count"`
	CommentCount int64    `json:"comment_count" db:"comment_count"`
	PlayCount   *int64    `json:"play_count,omitempty" db:"play_count"`
	IsAd        bool      `json:"is_ad" db:"is_ad"`
	PostedAt    time.Time `json:"posted_at" db:"posted_at"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Asset represents media assets associated with posts
type Asset struct {
	ID                   string    `json:"id" db:"id"`
	PostID               string    `json:"post_id" db:"post_id"`
	AssetType            string    `json:"asset_type" db:"asset_type"` // "image", "video", "carousel"
	URL                  string    `json:"url" db:"url"`
	TaggedUserUsernames  []string  `json:"tagged_user_usernames" db:"tagged_user_usernames"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time `json:"updated_at" db:"updated_at"`
}

// ProcessingJob represents a batch processing job
type ProcessingJob struct {
	ID              string            `json:"id" db:"id"`
	Status          string            `json:"status" db:"status"` // "pending", "running", "completed", "failed"
	TotalUsers      int               `json:"total_users" db:"total_users"`
	ProcessedUsers  int               `json:"processed_users" db:"processed_users"`
	SuccessfulUsers int               `json:"successful_users" db:"successful_users"`
	FailedUsers     int               `json:"failed_users" db:"failed_users"`
	MaxConcurrency  int               `json:"max_concurrency" db:"max_concurrency"`
	StartedAt       *time.Time        `json:"started_at,omitempty" db:"started_at"`
	CompletedAt     *time.Time        `json:"completed_at,omitempty" db:"completed_at"`
	Errors          map[string]string `json:"errors" db:"errors"`
	CreatedAt       time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at" db:"updated_at"`
}