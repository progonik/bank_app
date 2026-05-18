package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	ServerPort string

	JWTSecret          string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration

	SQBBaseURL   string
	SQBLocalAddr string

	BirdarchaBaseURL      string
	BirdarchaSyncInterval time.Duration
	BirdarchaCutoffDate   string
}

func (c *Config) DatabaseURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode,
	)
}

func Load() (*Config, error) {
	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
	}

	accessExpiry, err := time.ParseDuration(getEnv("ACCESS_TOKEN_EXPIRY", "15m"))
	if err != nil {
		return nil, fmt.Errorf("invalid ACCESS_TOKEN_EXPIRY: %w", err)
	}

	refreshExpiry, err := time.ParseDuration(getEnv("REFRESH_TOKEN_EXPIRY", "168h"))
	if err != nil {
		return nil, fmt.Errorf("invalid REFRESH_TOKEN_EXPIRY: %w", err)
	}

	birdarchaInterval, err := time.ParseDuration(getEnv("BIRDARCHA_SYNC_INTERVAL", "10m"))
	if err != nil {
		return nil, fmt.Errorf("invalid BIRDARCHA_SYNC_INTERVAL: %w", err)
	}

	cfg := &Config{
		DBHost:             getEnv("DB_HOST", "localhost"),
		DBPort:             dbPort,
		DBUser:             getEnv("DB_USER", "postgres"),
		DBPassword:         getEnv("DB_PASSWORD", "postgres"),
		DBName:             getEnv("DB_NAME", "bank_app"),
		DBSSLMode:          getEnv("DB_SSLMODE", "disable"),
		ServerPort:         getEnv("SERVER_PORT", "8080"),
		JWTSecret:          getEnv("JWT_SECRET", ""),
		AccessTokenExpiry:  accessExpiry,
		RefreshTokenExpiry: refreshExpiry,
		SQBBaseURL:         getEnv("SQB_BASE_URL", "https://ocrm.sqb.uz/backend/leads"),
		SQBLocalAddr:       getEnv("SQB_LOCAL_ADDR", "46.8.176.85"),

		BirdarchaBaseURL:      getEnv("BIRDARCHA_BASE_URL", "https://api.birdarcha.uz"),
		BirdarchaSyncInterval: birdarchaInterval,
		BirdarchaCutoffDate:   getEnv("BIRDARCHA_CUTOFF_DATE", "15.04.2026"),
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET environment variable is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
