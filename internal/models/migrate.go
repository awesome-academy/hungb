package models

import (
	"log/slog"

	"gorm.io/gorm"
)

// AllModels returns the list of all GORM models for AutoMigrate.
func AllModels() []interface{} {
	return []interface{}{
		&User{},
		&SocialAccount{},
		&BankAccount{},
		&Tour{},
		&Category{},
		&TourSchedule{},
		&Booking{},
		&Payment{},
		&Rating{},
		&Review{},
		&ReviewLike{},
		&Comment{},
	}
}

func Migrate(db *gorm.DB) error {
	slog.Info("running database migration...")

	if err := db.AutoMigrate(AllModels()...); err != nil {
		slog.Error("migration failed", "error", err)
		return err
	}

	slog.Info("database migration completed successfully")
	return nil
}
