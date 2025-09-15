package external

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"instagram-user-processor/pkg/database"
	"io"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/time/rate"
)

var (
	client      *http.Client
	rateLimiter *rate.Limiter
	apiKey      string
)

const (
	maxRetries  = 5
	baseDelayMS = 500 // Base delay in milliseconds
)

// Initialize the RocketAPI client
func InitRocketAPI() {
	client = &http.Client{
		Timeout: 30 * time.Second,
	}

	// Rate limiter: 10 requests per second (RocketAPI limit)
	rateLimiter = rate.NewLimiter(rate.Every(100*time.Millisecond), 10)

	apiKey = os.Getenv("ROCKETAPI_KEY")
	if apiKey == "" {
		log.Warn().Msg("ROCKETAPI_KEY not set, using demo key")
		apiKey = "demo_key_123"
	}

	log.Info().Msg("RocketAPI client initialized")
}

// RocketAPIResponse represents the wrapper response from RocketAPI
type RocketAPIResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Response struct {
		StatusCode int             `json:"status_code"`
		Body       json.RawMessage `json:"body"`
	} `json:"response"`
}

// RocketAPIUser represents the user data from RocketAPI
type RocketAPIUser struct {
	ID              string `json:"id"`
	Username        string `json:"username"`
	FullName        string `json:"full_name"`
	Biography       string `json:"biography"`
	IsVerified      bool   `json:"is_verified"`
	IsBusinessAccount bool `json:"is_business_account"`
	IsProfessionalAccount bool `json:"is_professional_account"`
	IsPrivate       bool   `json:"is_private"`
	CategoryName    string `json:"category_name"`

	EdgeFollow struct {
		Count int64 `json:"count"`
	} `json:"edge_follow"`

	EdgeFollowedBy struct {
		Count int64 `json:"count"`
	} `json:"edge_followed_by"`

	EdgeOwnerToTimelineMedia struct {
		Count int64 `json:"count"`
	} `json:"edge_owner_to_timeline_media"`
}

// UserNotFoundError represents a user not found error
type UserNotFoundError struct {
	Username string
	Message  string
}

func (e UserNotFoundError) Error() string {
	return fmt.Sprintf("user %s not found: %s", e.Username, e.Message)
}

// retryWithBackoff implements exponential backoff retry logic
func retryWithBackoff(ctx context.Context, operation func() (*RocketAPIResponse, []byte, error), operationName string) (*RocketAPIResponse, []byte, error) {
	var lastErr error
	var lastBody []byte

	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, body, err := operation()

		// If no error and response status is not "error", return success
		if err == nil && resp != nil && resp.Status != "error" {
			return resp, body, nil
		}

		// Don't retry UserNotFoundError - return immediately
		var userNotFoundErr UserNotFoundError
		if errors.As(err, &userNotFoundErr) {
			return resp, body, err
		}

		// Store the error/body for potential final return
		lastErr = err
		lastBody = body

		// If this was the last attempt, don't wait
		if attempt == maxRetries-1 {
			break
		}

		// Calculate exponential backoff delay: baseDelay * 2^attempt
		delay := time.Duration(baseDelayMS*int(math.Pow(2, float64(attempt)))) * time.Millisecond

		log.Warn().
			Err(err).
			Str("operation", operationName).
			Int("attempt", attempt+1).
			Dur("retry_delay", delay).
			Msg("operation failed, retrying")

		// Wait with context cancellation support
		select {
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	// All retries exhausted
	log.Error().Err(lastErr).Msgf("%s failed after %d attempts", operationName, maxRetries)
	return nil, lastBody, lastErr
}

// ScrapeInstagramUser scrapes user data from Instagram via RocketAPI
func ScrapeInstagramUser(ctx context.Context, username string) (*database.User, error) {
	if username == "" {
		return nil, errors.New("username cannot be empty")
	}

	log.Debug().Str("username", username).Msg("scraping Instagram user")

	operation := func() (*RocketAPIResponse, []byte, error) {
		// Respect rate limit
		if rateLimiter == nil {
			return nil, nil, fmt.Errorf("rate limiter not initialized - call InitRocketAPI() first")
		}
		if err := rateLimiter.Wait(ctx); err != nil {
			return nil, nil, fmt.Errorf("rate limit wait failed: %w", err)
		}

		requestURL := "https://v1.rocketapi.io/instagram/user/get_info"
		reqBody := fmt.Sprintf(`{"username": "%s"}`, username)

		req, err := http.NewRequestWithContext(ctx, "POST", requestURL, strings.NewReader(reqBody))
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Token %s", apiKey))

		res, err := client.Do(req)
		if err != nil {
			return nil, nil, fmt.Errorf("HTTP request failed: %w", err)
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read response body: %w", err)
		}

		// Handle HTTP errors
		if res.StatusCode == http.StatusNotFound {
			return nil, body, UserNotFoundError{Username: username, Message: "HTTP 404"}
		}

		if res.StatusCode != http.StatusOK {
			return nil, body, fmt.Errorf("unexpected HTTP status %d: %s", res.StatusCode, string(body))
		}

		// Parse RocketAPI wrapper response
		var resp RocketAPIResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, body, fmt.Errorf("failed to parse RocketAPI response: %w", err)
		}

		// Handle RocketAPI-level errors
		if resp.Status == "error" || resp.Status == "fail" {
			if strings.Contains(resp.Message, "user not found") || strings.Contains(resp.Message, "User not found") {
				return &resp, body, UserNotFoundError{Username: username, Message: resp.Message}
			}
			return &resp, body, fmt.Errorf("RocketAPI error: %s", resp.Message)
		}

		return &resp, body, nil
	}

	resp, _, err := retryWithBackoff(ctx, operation, "ScrapeInstagramUser")
	if err != nil {
		return nil, err
	}

	// Parse the actual user data
	var userResp struct {
		User RocketAPIUser `json:"user"`
	}

	if err := json.Unmarshal(resp.Response.Body, &userResp); err != nil {
		return nil, fmt.Errorf("failed to parse user data: %w", err)
	}

	// Convert RocketAPI user to our database user model
	user := &database.User{
		ID:                    userResp.User.ID,
		Username:              userResp.User.Username,
		FullName:              sql.NullString{String: userResp.User.FullName, Valid: userResp.User.FullName != ""},
		Biography:             sql.NullString{String: userResp.User.Biography, Valid: userResp.User.Biography != ""},
		IsVerified:            userResp.User.IsVerified,
		IsBusinessAccount:     userResp.User.IsBusinessAccount,
		IsProfessionalAccount: userResp.User.IsProfessionalAccount,
		IsPrivate:             userResp.User.IsPrivate,
		CategoryName:          sql.NullString{String: userResp.User.CategoryName, Valid: userResp.User.CategoryName != ""},
		Followers:             userResp.User.EdgeFollowedBy.Count,
		Following:             userResp.User.EdgeFollow.Count,
		Posts:                 userResp.User.EdgeOwnerToTimelineMedia.Count,
		ScrapedAt:             time.Now(),
	}

	log.Debug().
		Str("username", username).
		Str("user_id", user.ID).
		Int64("followers", user.Followers).
		Bool("verified", user.IsVerified).
		Msg("successfully scraped Instagram user")

	return user, nil
}