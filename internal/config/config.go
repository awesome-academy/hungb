package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port    string
	GinMode string

	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	SessionSecret string

	GoogleClientID     string
	GoogleClientSecret string
	FBClientID         string
	FBClientSecret     string
	TwitterAPIKey      string
	TwitterAPISecret   string
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

// LoadConfig loads configuration from .env file and environment variables.
// Environment variables take precedence over .env values.
func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		slog.Warn("no .env file found, using environment variables only")
	}

	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		dbPort = 5432
	}

	return &Config{
		Port:    getEnv("PORT", "8080"),
		GinMode: getEnv("GIN_MODE", "debug"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     dbPort,
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "sun_booking_tours"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		SessionSecret: getEnv("SESSION_SECRET", "change-me-in-production"),

		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		FBClientID:         getEnv("FACEBOOK_CLIENT_ID", ""),
		FBClientSecret:     getEnv("FACEBOOK_CLIENT_SECRET", ""),
		TwitterAPIKey:      getEnv("TWITTER_API_KEY", ""),
		TwitterAPISecret:   getEnv("TWITTER_API_SECRET", ""),
	}
}

// getEnv returns the value of an environment variable or a default value.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
