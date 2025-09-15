package utils

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// InitLogger initializes the global logger
func InitLogger(environment string) {
	// Set log level
	level := strings.ToLower(os.Getenv("LOG_LEVEL"))
	switch level {
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn", "warning":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		if environment == "development" {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		} else {
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
		}
	}

	// Configure output format
	if environment == "development" {
		// Pretty console output for development
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.RFC3339,
			NoColor:    false,
		})
	} else {
		// JSON output for production
		log.Logger = zerolog.New(os.Stderr).
			With().
			Timestamp().
			Str("service", "instagram-user-processor").
			Logger()
	}

	log.Info().
		Str("level", zerolog.GlobalLevel().String()).
		Str("environment", environment).
		Msg("logger initialized")
}