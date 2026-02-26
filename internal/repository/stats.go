package repository

import (
	"context"
	"time"

	"sun-booking-tours/internal/models"

	"gorm.io/gorm"
)

type StatsRepository struct {
	db *gorm.DB
}

func NewStatsRepository(db *gorm.DB) *StatsRepository {
	return &StatsRepository{db: db}
}

// CountUsers returns the total number of users.
func (r *StatsRepository) CountUsers(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.User{}).Count(&count).Error
	return count, err
}

// CountActiveTours returns the number of tours with status "active".
func (r *StatsRepository) CountActiveTours(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Tour{}).
		Where("status = ?", "active").
		Count(&count).Error
	return count, err
}

// CountTodayBookings returns the number of bookings created today.
func (r *StatsRepository) CountTodayBookings(ctx context.Context) (int64, error) {
	var count int64
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	err := r.db.WithContext(ctx).Model(&models.Booking{}).
		Where("created_at >= ? AND created_at < ?", startOfDay, endOfDay).
		Count(&count).Error
	return count, err
}

// SumMonthRevenue returns the total successful payment amount for the current month.
func (r *StatsRepository) SumMonthRevenue(ctx context.Context) (float64, error) {
	var total float64
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	err := r.db.WithContext(ctx).Model(&models.Payment{}).
		Where("status = ? AND paid_at >= ?", "success", startOfMonth).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error
	return total, err
}

// RecentBookings returns the most recent bookings with User and Tour preloaded.
func (r *StatsRepository) RecentBookings(ctx context.Context, limit int) ([]models.Booking, error) {
	var bookings []models.Booking
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Tour").
		Order("created_at DESC").
		Limit(limit).
		Find(&bookings).Error
	return bookings, err
}

// PendingReviews returns reviews with status "pending" with User preloaded.
func (r *StatsRepository) PendingReviews(ctx context.Context, limit int) ([]models.Review, error) {
	var reviews []models.Review
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("status = ?", "pending").
		Order("created_at DESC").
		Limit(limit).
		Find(&reviews).Error
	return reviews, err
}
