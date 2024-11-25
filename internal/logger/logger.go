// internal/logger/logger.go

package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Init() {
	// Pretty print logs in development
	if os.Getenv("ENV") != "production" {
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		})
	}

	// Set global log level based on environment
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if os.Getenv("ENV") == "development" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// Add basic configuration
	log.Logger = log.With().
		Timestamp().
		Caller().
		Logger()
}

// GetLogger returns a logger instance with a given context
func GetLogger(context string) zerolog.Logger {
	return log.With().Str("context", context).Logger()
}
