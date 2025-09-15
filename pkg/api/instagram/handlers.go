package instagram

import (
	"database/sql"
	"errors"
	"fmt"
	"instagram-user-processor/pkg/database"
	"instagram-user-processor/pkg/external"
	"net/http"
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
			Source:      "database",
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
			"error": "invalid request format",
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

	// TODO: CANDIDATE IMPLEMENTS THIS
	//
	// Expected implementation:
	// 1. Create a worker pool with req.MaxConcurrency workers
	// 2. Process users in parallel while respecting rate limits
	// 3. Handle individual user failures gracefully
	// 4. Return detailed results for each user
	// 5. Implement proper timeout handling
	//
	// Hints:
	// - Use the existing external.ScrapeInstagramUser function
	// - Use the existing database.UpsertUser function
	// - Respect the RocketAPI rate limit (10 req/sec)
	// - Consider using channels for communication between workers
	// - Track progress and errors for each user
	//
	// Example successful response structure:
	// {
	//   "results": [
	//     {
	//       "username": "user1",
	//       "status": "success",
	//       "user": {...user data...},
	//       "processed_at": "2023-..."
	//     },
	//     {
	//       "username": "user2",
	//       "status": "error",
	//       "error": "user not found",
	//       "processed_at": "2023-..."
	//     }
	//   ],
	//   "summary": {
	//     "total": 2,
	//     "successful": 1,
	//     "failed": 1,
	//     "duration_seconds": 15.2
	//   }
	// }

	// PLACEHOLDER RESPONSE - REPLACE WITH ACTUAL IMPLEMENTATION
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "batch processing not implemented yet",
		"task": "Implement parallel processing of Instagram users",
		"requirements": []string{
			"Process users in parallel with configurable concurrency",
			"Respect RocketAPI rate limits (10 req/sec)",
			"Handle individual user failures gracefully",
			"Return detailed status for each user",
			"Implement timeout handling",
		},
		"hints": map[string]interface{}{
			"functions_to_use": []string{
				"external.ScrapeInstagramUser(ctx, username)",
				"database.UpsertUser(ctx, user)",
			},
			"patterns_to_implement": []string{
				"Worker pool pattern",
				"Rate limiting across workers",
				"Error aggregation",
				"Progress tracking",
			},
		},
	})
}