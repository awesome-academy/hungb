package database

import (
	"log/slog"

	"sun-booking-tours/internal/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Seed inserts initial data into the database.
// Uses FirstOrCreate to avoid duplicates on repeated runs.
func Seed(db *gorm.DB) error {
	slog.Info("seeding database...")

	if err := seedAdminUser(db); err != nil {
		return err
	}

	if err := seedCategories(db); err != nil {
		return err
	}

	slog.Info("database seeding completed successfully")
	return nil
}

// seedAdminUser creates the default admin user if not exists.
func seedAdminUser(db *gorm.DB) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := models.User{
		Email:    "admin@sunbooking.vn",
		Password: string(hashedPassword),
		FullName: "System Admin",
		Role:     "admin",
		Status:   "active",
	}

	result := db.Where("email = ?", admin.Email).FirstOrCreate(&admin)
	if result.Error != nil {
		slog.Error("failed to seed admin user", "error", result.Error)
		return result.Error
	}

	if result.RowsAffected > 0 {
		slog.Info("admin user created", "email", admin.Email)
	} else {
		slog.Info("admin user already exists", "email", admin.Email)
	}

	return nil
}

// seedCategories creates sample tour categories if not exist.
func seedCategories(db *gorm.DB) error {
	categories := []models.Category{
		{Name: "Du lịch biển", Slug: "du-lich-bien", Description: "Các tour du lịch biển đảo"},
		{Name: "Du lịch núi", Slug: "du-lich-nui", Description: "Các tour leo núi, trekking"},
		{Name: "Du lịch văn hóa", Slug: "du-lich-van-hoa", Description: "Tham quan di tích lịch sử, văn hóa"},
		{Name: "Du lịch ẩm thực", Slug: "du-lich-am-thuc", Description: "Khám phá ẩm thực địa phương"},
		{Name: "Du lịch sinh thái", Slug: "du-lich-sinh-thai", Description: "Trải nghiệm thiên nhiên, sinh thái"},
	}

	for _, cat := range categories {
		result := db.Where("slug = ?", cat.Slug).FirstOrCreate(&cat)
		if result.Error != nil {
			slog.Error("failed to seed category", "slug", cat.Slug, "error", result.Error)
			return result.Error
		}
		if result.RowsAffected > 0 {
			slog.Info("category created", "name", cat.Name)
		}
	}

	return nil
}
