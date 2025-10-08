package instagram

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"instagram-user-processor/pkg/database"
	"instagram-user-processor/pkg/external"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func fetchUser(ctx context.Context, username string) (UserResponse, error) {

	// Get user data from database first
	user, err := database.GetUserByUsername(ctx, username)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Error().Err(err).Str("username", username).Msg("database error")
		return UserResponse{}, fmt.Errorf("database error: %w", err)
	}

	source := "database"
	// If user not found in database, scrape from RocketAPI
	if errors.Is(err, sql.ErrNoRows) {
		log.Info().Str("username", username).Msg("user not found in database, scraping from RocketAPI")

		scrapedUser, err := external.ScrapeInstagramUser(ctx, username)
		if err != nil {
			log.Error().Err(err).Str("username", username).Msg("failed to scrape user")
			return UserResponse{}, fmt.Errorf("user not found: %w", err)
		}

		// Store in database
		if err := database.UpsertUser(ctx, scrapedUser); err != nil {
			log.Error().Err(err).Str("username", username).Msg("failed to store user")
			return UserResponse{}, fmt.Errorf("failed to store user: %w", err)
		}

		user = scrapedUser
		source = "rocketapi"
	}

	// Get detailed stats
	stats, err := database.GetUserStats(ctx, user.ID)
	if err != nil {
		log.Error().Err(err).Str("user_id", user.ID).Msg("failed to get user stats")
		// Continue without stats
	}

	userResponse := UserResponse{
		User:  *user,
		Stats: stats,
		Meta: ResponseMeta{
			ProcessedAt: time.Now(),
			Source:      source,
		},
	}
	return userResponse, nil
}

// GetUserHandler handles single user requests - WORKING IMPLEMENTATION
// GET /api/v1/instagram/user/:username
func GetUserHandler(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "username is required",
		})
		return
	}

	response, err := fetchUser(c.Request.Context(), username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetUserStatsHandler gets detailed user statistics
// GET /api/v1/instagram/users/:id/stats
func GetUserStatsHandler(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user ID is required",
		})
		return
	}

	stats, err := database.GetUserStats(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "user not found",
			})
			return
		}
		log.Error().Err(err).Str("user_id", userID).Msg("failed to get user stats")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get user stats",
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func fetchDataForUsers(ctx context.Context, request BatchRequest) ([]UserResponse, error) {
	userResponses := make([]UserResponse, len(request.Usernames))
	wg := sync.WaitGroup{}
	startIndex := 0
	endIndex := request.MaxConcurrency

	for endIndex <= len(request.Usernames) {
		for i, username := range request.Usernames[startIndex:endIndex] {
			wg.Add(1)
			go func(i int, username string) {
				defer wg.Done()
				userResponse, err := fetchUser(ctx, username)
				if err != nil {
					log.Error().Err(err).Str("username", username).Msg("failed to fetch user")
					return
				}
				userResponses[i] = userResponse
			}(i, username)
		}
		startIndex = endIndex
		endIndex = endIndex + request.MaxConcurrency
		wg.Wait()
	}

	return userResponses, nil
}

// BatchProcessUsersHandler handles batch user processing requests
// POST /api/v1/instagram/users/batch
//
// THIS IS THE MAIN CHALLENGE FOR CANDIDATES TO IMPLEMENT
func BatchProcessUsersHandler(c *gin.Context) {
	var req BatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Validate request
	if len(req.Usernames) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "usernames array cannot be empty",
		})
		return
	}

	if len(req.Usernames) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "maximum 100 users per batch",
		})
		return
	}

	// Set defaults
	if req.MaxConcurrency <= 0 {
		req.MaxConcurrency = 5 // Default concurrency
	}
	if req.MaxConcurrency > 20 {
		req.MaxConcurrency = 20 // Max allowed concurrency
	}

	if req.TimeoutSeconds <= 0 {
		req.TimeoutSeconds = 300 // Default 5 minutes
	}

	log.Info().
		Int("user_count", len(req.Usernames)).
		Int("max_concurrency", req.MaxConcurrency).
		Int("timeout", req.TimeoutSeconds).
		Msg("starting batch user processing")

	userResponses, err := fetchDataForUsers(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, userResponses)
}
