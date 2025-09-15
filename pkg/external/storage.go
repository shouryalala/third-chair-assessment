package external

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// StorageClient interface for profile picture storage
type StorageClient interface {
	UploadProfilePicture(ctx context.Context, userID string, imageURL string) error
	GetUploadURL(userID string) string
}

// MockStorageClient simulates AWS S3 storage operations
type MockStorageClient struct {
	uploads map[string]string // userID -> imageURL
	mu      sync.RWMutex
}

// NewMockStorageClient creates a new mock storage client
func NewMockStorageClient() *MockStorageClient {
	return &MockStorageClient{
		uploads: make(map[string]string),
	}
}

// UploadProfilePicture simulates uploading a profile picture to S3
func (m *MockStorageClient) UploadProfilePicture(ctx context.Context, userID string, imageURL string) error {
	if userID == "" {
		return fmt.Errorf("userID cannot be empty")
	}

	if imageURL == "" {
		return fmt.Errorf("imageURL cannot be empty")
	}

	// Simulate upload delay (network latency)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(100 * time.Millisecond):
		// Continue with upload
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.uploads[userID] = imageURL

	log.Debug().
		Str("user_id", userID).
		Str("image_url", imageURL).
		Msg("mock uploaded profile picture")

	return nil
}

// GetUploadURL returns the mock S3 URL for a user's profile picture
func (m *MockStorageClient) GetUploadURL(userID string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if _, exists := m.uploads[userID]; exists {
		return fmt.Sprintf("https://mock-s3-bucket.s3.amazonaws.com/profile-pics/%s.jpg", userID)
	}

	return ""
}

// GetUploadCount returns the number of uploads (for testing)
func (m *MockStorageClient) GetUploadCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.uploads)
}

// GetUploads returns all uploads (for testing)
func (m *MockStorageClient) GetUploads() map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	uploads := make(map[string]string)
	for k, v := range m.uploads {
		uploads[k] = v
	}
	return uploads
}

// Global mock storage client instance
var mockStorage *MockStorageClient

// InitMockStorage initializes the global mock storage client
func InitMockStorage() {
	mockStorage = NewMockStorageClient()
	log.Info().Msg("Mock storage client initialized")
}

// GetStorageClient returns the global storage client
func GetStorageClient() StorageClient {
	if mockStorage == nil {
		InitMockStorage()
	}
	return mockStorage
}