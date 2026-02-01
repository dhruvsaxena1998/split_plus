package config

import (
	"errors"
	"os"
	"time"
)

type Config struct {
	DatabaseURL        string
	JWTSecret          string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
	CORSAllowedOrigins string
	Environment        string
}

func Load() (*Config, error) {
	cfg := &Config{
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		JWTSecret:          os.Getenv("JWT_SECRET"),
		CORSAllowedOrigins: os.Getenv("CORS_ALLOWED_ORIGINS"),
		Environment:        getEnvOrDefault("ENV", "development"),
	}

	// Validate required fields
	if cfg.DatabaseURL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}

	if cfg.JWTSecret == "" {
		return nil, errors.New("JWT_SECRET is required")
	}

	// Parse token expiry durations
	accessTokenExpiry := getEnvOrDefault("ACCESS_TOKEN_EXPIRY", "168h") // 7 days
	var err error
	cfg.AccessTokenExpiry, err = time.ParseDuration(accessTokenExpiry)
	if err != nil {
		return nil, errors.New("invalid ACCESS_TOKEN_EXPIRY format")
	}

	refreshTokenExpiry := getEnvOrDefault("REFRESH_TOKEN_EXPIRY", "720h") // 30 days
	cfg.RefreshTokenExpiry, err = time.ParseDuration(refreshTokenExpiry)
	if err != nil {
		return nil, errors.New("invalid REFRESH_TOKEN_EXPIRY format")
	}

	return cfg, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
