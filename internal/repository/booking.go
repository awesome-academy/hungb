package repository

import (
	"context"
	"fmt"

	appErrors "sun-booking-tours/internal/errors"
	"sun-booking-tours/internal/models"

	"gorm.io/gorm"
)

type BookingRepo interface {
	Create(ctx context.Context, booking *models.Booking) error
	FindByID(ctx context.Context, id uint) (*models.Booking, error)
	FindByUserID(ctx context.Context, userID uint, page, limit int) ([]models.Booking, int64, error)
	UpdateStatus(ctx context.Context, id uint, status string) error
}

type bookingRepository struct {
	db *gorm.DB
}

func NewBookingRepository(db *gorm.DB) BookingRepo {
	return &bookingRepository{db: db}
}

func (r *bookingRepository) Create(ctx context.Context, booking *models.Booking) error {
	if err := r.db.WithContext(ctx).Create(booking).Error; err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxBookingCreate, err)
	}
	return nil
}

func (r *bookingRepository) FindByID(ctx context.Context, id uint) (*models.Booking, error) {
	var booking models.Booking
	if err := r.db.WithContext(ctx).
		Preload("Tour").
		Preload("Schedule").
		Preload("Payments").
		First(&booking, id).Error; err != nil {
		return nil, fmt.Errorf("%s: %w", appErrors.ErrCtxBookingFindByID, err)
	}
	return &booking, nil
}

func (r *bookingRepository) FindByUserID(ctx context.Context, userID uint, page, limit int) ([]models.Booking, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&models.Booking{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("%s: %w", appErrors.ErrCtxBookingCount, err)
	}

	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	var bookings []models.Booking
	if err := r.db.WithContext(ctx).
		Preload("Tour").
		Preload("Schedule").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&bookings).Error; err != nil {
		return nil, 0, fmt.Errorf("%s: %w", appErrors.ErrCtxBookingFindByUser, err)
	}
	return bookings, total, nil
}

func (r *bookingRepository) UpdateStatus(ctx context.Context, id uint, status string) error {
	if err := r.db.WithContext(ctx).
		Model(&models.Booking{}).
		Where("id = ?", id).
		Update("status", status).Error; err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxBookingUpdateStatus, err)
	}
	return nil
}
