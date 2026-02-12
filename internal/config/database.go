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
// Returns the *gorm.DB instance or exits on failure.
func ConnectDB(cfg *Config) *gorm.DB {
	dsn := cfg.DSN()

	// Configure GORM logger level based on gin mode
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
		slog.Error("failed to connect to database", "error", err)
		panic(fmt.Sprintf("failed to connect to database: %v", err))
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		slog.Error("failed to get underlying sql.DB", "error", err)
		panic(fmt.Sprintf("failed to get sql.DB: %v", err))
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	sqlDB.SetConnMaxIdleTime(1 * time.Minute)

	// Verify connection
	if err := sqlDB.Ping(); err != nil {
		slog.Error("failed to ping database", "error", err)
		panic(fmt.Sprintf("failed to ping database: %v", err))
	}

	slog.Info("database connected successfully",
		"host", cfg.DBHost,
		"port", cfg.DBPort,
		"db", cfg.DBName,
	)

	return db
}
