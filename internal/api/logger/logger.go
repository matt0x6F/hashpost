package logger

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Init initializes the global logger with zerolog
func Init() {
	InitWithLevel("info")
}

// InitWithLevel initializes the global logger with a specific log level
func InitWithLevel(level string) {
	// Configure zerolog for pretty console output in development
	zerolog.TimeFieldFormat = time.RFC3339

	// Parse log level
	logLevel := parseLogLevel(level)
	zerolog.SetGlobalLevel(logLevel)

	// Use pretty console writer for development
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}

	// Set the global logger
	log.Logger = log.Output(consoleWriter)
}

// parseLogLevel converts a string log level to zerolog level
func parseLogLevel(level string) zerolog.Level {
	switch strings.ToLower(level) {
	case "trace":
		return zerolog.TraceLevel
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	default:
		return zerolog.InfoLevel
	}
}

// GetLogger returns a logger instance for the given component
func GetLogger(component string) zerolog.Logger {
	return log.With().Str("component", component).Logger()
}

// GetRequestLogger returns a logger instance for HTTP requests
func GetRequestLogger() zerolog.Logger {
	return log.With().Str("component", "http").Logger()
}
