package config

import (
	"errors"
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServiceName     string
	Environment     string
	HTTPPort        string
	LogLevel        string
	DatabaseURL     string
	JWTSecret       string
	AdminEmail      string
	AdminPassword   string
	ShutdownTimeout time.Duration
	MaxUploadBytes  int64
}

func Load() (Config, error) {
	cfg := Config{
		ServiceName:     getenv("SERVICE_NAME", "rag-bot-api"),
		Environment:     getenv("ENVIRONMENT", "local"),
		HTTPPort:        getenv("HTTP_PORT", "8080"),
		LogLevel:        getenv("LOG_LEVEL", "info"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		JWTSecret:       os.Getenv("JWT_SECRET"),
		AdminEmail:      os.Getenv("ADMIN_EMAIL"),
		AdminPassword:   os.Getenv("ADMIN_PASSWORD"),
		ShutdownTimeout: getDuration("SHUTDOWN_TIMEOUT", 10*time.Second),
		MaxUploadBytes:  getInt64("MAX_UPLOAD_BYTES", 20*1024*1024),
	}

	if cfg.MaxUploadBytes <= 0 {
		return Config{}, errors.New("MAX_UPLOAD_BYTES must be positive")
	}
	return cfg, nil
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getInt64(key string, fallback int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}
	return parsed
}
