package utils

import (
	"os"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

// Config holds application configuration
type Config struct {
	Environment    string
	ServerPort     string
	DatabaseURL    string
	RocketAPIKey   string
	RateLimit      int  // requests per second
	MaxConcurrency int  // max concurrent workers
	LogLevel       string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	config := &Config{
		Environment:    getEnvWithDefault("ENVIRONMENT", "development"),
		ServerPort:     getEnvWithDefault("PORT", "8080"),
		DatabaseURL:    getEnvWithDefault("DATABASE_URL", "postgres://user:password@localhost:5432/instagram_processor?sslmode=disable"),
		RocketAPIKey:   getEnvWithDefault("ROCKETAPI_KEY", "demo_key_123"),
		RateLimit:      getEnvIntWithDefault("RATE_LIMIT", 10),
		MaxConcurrency: getEnvIntWithDefault("MAX_CONCURRENCY", 5),
		LogLevel:       getEnvWithDefault("LOG_LEVEL", "info"),
	}

	// Validate configuration
	if config.RateLimit <= 0 {
		config.RateLimit = 10
		log.Warn().Msg("invalid RATE_LIMIT, using default: 10")
	}

	if config.MaxConcurrency <= 0 {
		config.MaxConcurrency = 5
		log.Warn().Msg("invalid MAX_CONCURRENCY, using default: 5")
	}

	if config.MaxConcurrency > 50 {
		config.MaxConcurrency = 50
		log.Warn().Msg("MAX_CONCURRENCY too high, limiting to: 50")
	}

	// Log configuration (without sensitive data)
	log.Info().
		Str("environment", config.Environment).
		Str("port", config.ServerPort).
		Int("rate_limit", config.RateLimit).
		Int("max_concurrency", config.MaxConcurrency).
		Str("log_level", config.LogLevel).
		Msg("configuration loaded")

	return config
}

// getEnvWithDefault gets an environment variable with a default value
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return strings.TrimSpace(value)
	}
	return defaultValue
}

// getEnvIntWithDefault gets an integer environment variable with a default value
func getEnvIntWithDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(strings.TrimSpace(value)); err == nil {
			return intVal
		}
		log.Warn().Str("key", key).Str("value", value).Msg("invalid integer environment variable, using default")
	}
	return defaultValue
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return strings.ToLower(c.Environment) == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return strings.ToLower(c.Environment) == "production"
}