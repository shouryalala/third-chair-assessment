package api

import (
	"fmt"
	"instagram-user-processor/pkg/api/instagram"
	"instagram-user-processor/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func InitRouter(config *utils.Config) *gin.Engine {
	r := gin.Default()

	// Add middleware
	r.Use(LoggingMiddleware())
	r.Use(CORSMiddleware())

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"service":   "instagram-user-processor",
			"timestamp": gin.H{"unix": gin.H{}},
		})
	})

	// API v1 group
	v1 := r.Group("/api/v1")

	// Instagram endpoints
	instagramGroup := v1.Group("/instagram")
	{
		// Existing single user endpoint (working implementation)
		instagramGroup.GET("/user/:username", instagram.GetUserHandler)

		// New batch endpoint (to be implemented by candidate)
		instagramGroup.POST("/users/batch", instagram.BatchProcessUsersHandler)

		// Helper endpoint for testing
		instagramGroup.GET("/users/:id/stats", instagram.GetUserStatsHandler)
	}

	// 404 handler
	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"error":   "route not found",
			"path":    c.Request.URL.Path,
			"method":  c.Request.Method,
		})
	})

	return r
}

// LoggingMiddleware provides request logging
func LoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("[%s] %s %s %d %s \"%s\" \"%s\" %s\n",
			param.TimeStamp.Format("2006-01-02 15:04:05"),
			param.Method,
			param.Path,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
			param.ClientIP,
		)
	})
}

// CORSMiddleware handles CORS
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}