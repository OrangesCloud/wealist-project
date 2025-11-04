package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Init initializes and returns a zap logger based on the provided level and environment
func Init(level string, env string) (*zap.Logger, error) {
	// Parse log level
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
		zapLevel = zapcore.InfoLevel // default to info
	}

	// Configure based on environment
	var config zap.Config
	if env == "prod" {
		// Production: JSON format
		config = zap.NewProductionConfig()
	} else {
		// Development: Console format
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// Set log level
	config.Level = zap.NewAtomicLevelAt(zapLevel)

	// Build logger
	logger, err := config.Build(
		zap.AddCallerSkip(1), // Skip one caller to show actual caller
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	return logger, nil
}

// NewNop returns a no-op logger for testing
func NewNop() *zap.Logger {
	return zap.NewNop()
}
