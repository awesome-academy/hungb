package config

import (
	"fmt"
	"log/slog"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ConnectDB opens a PostgreSQL connection using GORM with connection pool settings.
// Returns the *gorm.DB instance or an error if connection fails.
func ConnectDB(cfg *Config) (*gorm.DB, error) {
	dsn := cfg.DSN()

	var logLevel logger.LogLevel
	if cfg.GinMode == "debug" {
		logLevel = logger.Info
	} else {
		logLevel = logger.Warn
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	sqlDB.SetConnMaxIdleTime(1 * time.Minute)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	slog.Info("database connected successfully",
		"host", cfg.DBHost,
		"port", cfg.DBPort,
		"db", cfg.DBName,
	)

	return db, nil
}
