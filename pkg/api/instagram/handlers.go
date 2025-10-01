package instagram

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"instagram-user-processor/pkg/database"
	"instagram-user-processor/pkg/external"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

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

	// Get user data from database first
	user, err := database.GetUserByUsername(c.Request.Context(), username)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Error().Err(err).Str("username", username).Msg("database error")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "database error",
		})
		return
	}

	source := "database"
	// If user not found in database, scrape from RocketAPI
	if errors.Is(err, sql.ErrNoRows) {
		log.Info().Str("username", username).Msg("user not found in database, scraping from RocketAPI")

		scrapedUser, err := external.ScrapeInstagramUser(c.Request.Context(), username)
		if err != nil {
			log.Error().Err(err).Str("username", username).Msg("failed to scrape user")
			c.JSON(http.StatusNotFound, gin.H{
				"error": fmt.Sprintf("user not found: %s", username),
			})
			return
		}

		// Store in database
		if err := database.UpsertUser(c.Request.Context(), scrapedUser); err != nil {
			log.Error().Err(err).Str("username", username).Msg("failed to store user")
		}

		user = scrapedUser
		source = "rocketapi"
	}

	// Get detailed stats
	stats, err := database.GetUserStats(c.Request.Context(), user.ID)
	if err != nil {
		log.Error().Err(err).Str("user_id", user.ID).Msg("failed to get user stats")
		// Continue without stats
	}

	response := UserResponse{
		User:  *user,
		Stats: stats,
		Meta: ResponseMeta{
			ProcessedAt: time.Now(),
			Source:      source,
		},
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

	// Normalize and deduplicate usernames (lowercase to avoid duplicates)
	seen := make(map[string]struct{})
	unique := make([]string, 0, len(req.Usernames))
	for _, u := range req.Usernames {
		uname := strings.TrimSpace(strings.ToLower(u))
		if uname == "" {
			continue
		}
		if _, ok := seen[uname]; ok {
			continue
		}
		seen[uname] = struct{}{}
		unique = append(unique, uname)
	}
	if len(unique) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "no valid usernames provided",
		})
		return
	}

	// Context with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), time.Duration(req.TimeoutSeconds)*time.Second)
	defer cancel()

	startedAt := time.Now()

	// Global rate limiter for RocketAPI: 500/min ≈ 1 every 120ms
	// We also have an internal limiter in external.ScrapeInstagramUser (~10 rps),
	// but we gate here explicitly to 500/min to be safe.
	rateTicker := time.NewTicker(120 * time.Millisecond)
	defer rateTicker.Stop()

	jobs := make(chan string)
	results := make(chan UserResult)

	// Worker function
	worker := func() {
		for username := range jobs {
			select {
			case <-ctx.Done():
				results <- UserResult{Username: username, Status: "error", Error: ctx.Err().Error(), ProcessedAt: time.Now()}
				continue
			default:
			}

			// Try DB first
			user, err := database.GetUserByUsername(ctx, username)
			if err == nil && user != nil {
				results <- UserResult{Username: username, Status: "success", User: user, ProcessedAt: time.Now()}
				continue
			}

			// If not found in DB, scrape with rate limit
			select {
			case <-ctx.Done():
				results <- UserResult{Username: username, Status: "error", Error: ctx.Err().Error(), ProcessedAt: time.Now()}
				continue
			case <-rateTicker.C:
				// proceed
			}

			scraped, err := external.ScrapeInstagramUser(ctx, username)
			if err != nil {
				results <- UserResult{Username: username, Status: "error", Error: err.Error(), ProcessedAt: time.Now()}
				continue
			}

			// Upsert scraped user
			if upErr := database.UpsertUser(ctx, scraped); upErr != nil {
				log.Error().Err(upErr).Str("username", username).Msg("failed to upsert user")
				// still return success with user data, as scrape succeeded
			}

			results <- UserResult{Username: username, Status: "success", User: scraped, ProcessedAt: time.Now()}
		}
	}

	// Launch workers
	workerCount := req.MaxConcurrency
	for i := 0; i < workerCount; i++ {
		go worker()
	}

	// Feed jobs
	go func() {
		for _, u := range unique {
			select {
			case <-ctx.Done():
				close(jobs)
				return
			default:
				jobs <- u
			}
		}
		close(jobs)
	}()

	// Collect results
	collected := make([]UserResult, 0, len(unique))
	completed := 0
	for completed < len(unique) {
		select {
		case <-ctx.Done():
			// Drain remaining expected results as errors to maintain contract
			remaining := len(unique) - completed
			for i := 0; i < remaining; i++ {
				collected = append(collected, UserResult{Status: "error", Error: ctx.Err().Error(), ProcessedAt: time.Now()})
			}
			completed = len(unique)
		case res := <-results:
			collected = append(collected, res)
			completed++
		}
	}

	// Compute summary
	success := 0
	fail := 0
	for _, r := range collected {
		if r.Status == "success" {
			success++
		} else {
			fail++
		}
	}

	// Sort results by username for stable output
	sort.SliceStable(collected, func(i, j int) bool { return collected[i].Username < collected[j].Username })

	completedAt := time.Now()
	resp := BatchResponse{
		Results: collected,
		Summary: Summary{
			Total:           len(unique),
			Successful:      success,
			Failed:          fail,
			DurationSeconds: completedAt.Sub(startedAt).Seconds(),
			StartedAt:       startedAt,
			CompletedAt:     completedAt,
		},
	}

	c.JSON(http.StatusOK, resp)
}
