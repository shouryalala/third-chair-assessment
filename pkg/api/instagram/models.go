package instagram

import (
	"instagram-user-processor/pkg/database"
	"time"
)

// BatchRequest represents a batch processing request
type BatchRequest struct {
	Usernames      []string `json:"usernames" binding:"required"`
	MaxConcurrency int      `json:"max_concurrency,omitempty"`
	TimeoutSeconds int      `json:"timeout_seconds,omitempty"`
}

// BatchResponse represents a batch processing response
type BatchResponse struct {
	Results []UserResult `json:"results"`
	Summary Summary      `json:"summary"`
}

// UserResult represents the result for a single user
type UserResult struct {
	Username    string               `json:"username"`
	Status      string               `json:"status"` // "success", "error"
	User        *database.User       `json:"user,omitempty"`
	Error       string               `json:"error,omitempty"`
	ProcessedAt time.Time           `json:"processed_at"`
}

// Summary represents batch processing summary statistics
type Summary struct {
	Total           int     `json:"total"`
	Successful      int     `json:"successful"`
	Failed          int     `json:"failed"`
	DurationSeconds float64 `json:"duration_seconds"`
	StartedAt       time.Time `json:"started_at"`
	CompletedAt     time.Time `json:"completed_at"`
}

// UserResponse represents a single user response
type UserResponse struct {
	User  database.User    `json:"user"`
	Stats *database.UserStats `json:"stats,omitempty"`
	Meta  ResponseMeta     `json:"meta"`
}

// ResponseMeta provides metadata about the response
type ResponseMeta struct {
	ProcessedAt time.Time `json:"processed_at"`
	Source      string    `json:"source"` // "database", "api"
}

// ProgressUpdate represents real-time progress updates
type ProgressUpdate struct {
	Completed int     `json:"completed"`
	Total     int     `json:"total"`
	Progress  float64 `json:"progress"` // 0-100
	Status    string  `json:"status"`
}