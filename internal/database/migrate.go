package database

import (
	"log/slog"

	"sun-booking-tours/internal/models"

	"gorm.io/gorm"
)

// AllModels returns the list of all GORM models for AutoMigrate.
func AllModels() []interface{} {
	return []interface{}{
		&models.User{},
		&models.SocialAccount{},
		&models.BankAccount{},
		&models.Tour{},
		&models.Category{},
		&models.TourSchedule{},
		&models.Booking{},
		&models.Payment{},
		&models.Rating{},
		&models.Review{},
		&models.ReviewLike{},
		&models.Comment{},
		&models.ActivityLog{},
	}
}

// Migrate runs GORM AutoMigrate for all registered models.
func Migrate(db *gorm.DB) error {
	slog.Info("running database migration...")

	if err := db.AutoMigrate(AllModels()...); err != nil {
		slog.Error("migration failed", "error", err)
		return err
	}

	slog.Info("database migration completed successfully")
	return nil
}
