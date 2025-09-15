package main

import (
	"fmt"
	"net/http"
	"time"

	"instagram-user-processor/pkg/api"
	"instagram-user-processor/pkg/database"
	"instagram-user-processor/pkg/external"
	"instagram-user-processor/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// @title Instagram User Processor API
// @version 1.0
// @description Interview assignment for Instagram user batch processing
// @host localhost:8080
// @BasePath /api/v1

func main() {
	// Load configuration
	config := utils.LoadConfig()

	// Initialize logging
	utils.InitLogger(config.Environment)

	// Initialize database
	if err := database.Initialize(config.DatabaseURL); err != nil {
		log.Fatal().Err(err).Msg("failed to initialize database")
	}

	// Initialize RocketAPI client
	external.InitRocketAPI()

	// Set Gin mode
	if config.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize router
	router := api.InitRouter(config)

	// Configure HTTP server
	server := &http.Server{
		Addr:           fmt.Sprintf(":%s", config.ServerPort),
		Handler:        router,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	// Enable keep-alive connections
	server.SetKeepAlivesEnabled(true)

	log.Info().Msgf("Starting Instagram User Processor on port %s", config.ServerPort)
	log.Info().Msgf("Environment: %s", config.Environment)
	log.Info().Msgf("Rate limit: %d requests/second", config.RateLimit)

	if err := server.ListenAndServe(); err != nil {
		log.Fatal().Err(err).Msg("failed to start server")
	}
}