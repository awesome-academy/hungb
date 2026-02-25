package database

import (
	"errors"
	"log/slog"
	"os"

	"sun-booking-tours/internal/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Seed inserts initial data into the database.
// Only runs in debug mode to prevent accidental seeding in production.
func Seed(db *gorm.DB) error {
	if gin.Mode() != gin.DebugMode {
		slog.Warn("database seeding skipped: only runs in debug mode")
		return nil
	}

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
// Requires ADMIN_EMAIL and ADMIN_PASSWORD environment variables.
func seedAdminUser(db *gorm.DB) error {
	adminEmail := os.Getenv("ADMIN_EMAIL")
	adminPassword := os.Getenv("ADMIN_PASSWORD")

	if adminEmail == "" || adminPassword == "" {
		return errors.New("ADMIN_EMAIL and ADMIN_PASSWORD must be set for seeding")
	}

	// Check if admin already exists
	var existingAdmin models.User
	if err := db.Where("email = ?", adminEmail).First(&existingAdmin).Error; err == nil {
		slog.Info("admin user already exists", "email", adminEmail)
		return nil
	}

	// Hash password only if needed
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := models.User{
		Email:    adminEmail,
		Password: string(hashedPassword),
		FullName: "System Admin",
		Role:     "admin",
		Status:   "active",
	}

	if err := db.Create(&admin).Error; err != nil {
		slog.Error("failed to seed admin user", "error", err)
		return err
	}

	slog.Info("admin user created", "email", admin.Email)
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
