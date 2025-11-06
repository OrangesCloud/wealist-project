package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server struct {
		Port           string
		Env            string // dev, prod
		UseAutoMigrate bool   // Enable GORM AutoMigrate in dev mode
	}
	Database struct {
		URL string // postgresql://user:pass@host:5432/dbname
	}
	Redis struct {
		URL string // redis://:password@host:6379/1
	}
	JWT struct {
		Secret string // HS512 secret (shared with User Service)
	}
	UserService struct {
		URL string // http://user-service:8080
	}
	CORS struct {
		Origins []string
	}
	Log struct {
		Level string // debug, info, warn, error
	}
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	v := viper.New()

	// Set defaults
	v.SetDefault("SERVER_PORT", "8000")
	v.SetDefault("ENV", "dev")
	v.SetDefault("USE_AUTO_MIGRATE", true)
	v.SetDefault("LOG_LEVEL", "info")
	v.SetDefault("CORS_ORIGINS", "http://localhost:3000")

	// Bind environment variables only (no .env file loading)
	v.AutomaticEnv()

	// Read environment variables
	cfg := &Config{}

	// Server
	cfg.Server.Port = v.GetString("SERVER_PORT")
	cfg.Server.Env = v.GetString("ENV")
	cfg.Server.UseAutoMigrate = v.GetBool("USE_AUTO_MIGRATE")

	// Database
	cfg.Database.URL = v.GetString("DATABASE_URL")
	if cfg.Database.URL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	// Redis
	cfg.Redis.URL = v.GetString("REDIS_URL")
	if cfg.Redis.URL == "" {
		return nil, fmt.Errorf("REDIS_URL is required")
	}

	// JWT
	cfg.JWT.Secret = v.GetString("SECRET_KEY")
	if cfg.JWT.Secret == "" {
		return nil, fmt.Errorf("SECRET_KEY is required")
	}

	// User Service
	cfg.UserService.URL = v.GetString("USER_SERVICE_URL")
	if cfg.UserService.URL == "" {
		return nil, fmt.Errorf("USER_SERVICE_URL is required")
	}

	// CORS
	originsStr := v.GetString("CORS_ORIGINS")
	cfg.CORS.Origins = strings.Split(originsStr, ",")
	for i := range cfg.CORS.Origins {
		cfg.CORS.Origins[i] = strings.TrimSpace(cfg.CORS.Origins[i])
	}

	// Logging
	cfg.Log.Level = v.GetString("LOG_LEVEL")

	return cfg, nil
}
