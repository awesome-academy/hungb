package repository

import (
	"context"
	"fmt"
	"time"

	"sun-booking-tours/internal/constants"
	appErrors "sun-booking-tours/internal/errors"
	"sun-booking-tours/internal/models"

	"gorm.io/gorm"
)

type BookingFilter struct {
	Status   string
	UserID   uint
	TourID   uint
	DateFrom time.Time
	DateTo   time.Time
	Page     int
	Limit    int
}

type BookingRepo interface {
	Create(ctx context.Context, booking *models.Booking) error
	FindByID(ctx context.Context, id uint) (*models.Booking, error)
	FindByUserID(ctx context.Context, userID uint, page, limit int) ([]models.Booking, int64, error)
	FindAll(ctx context.Context, filter BookingFilter) ([]models.Booking, int64, error)
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

func (r *bookingRepository) FindAll(ctx context.Context, filter BookingFilter) ([]models.Booking, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.Booking{})

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.UserID > 0 {
		query = query.Where("user_id = ?", filter.UserID)
	}
	if filter.TourID > 0 {
		query = query.Where("tour_id = ?", filter.TourID)
	}
	if !filter.DateFrom.IsZero() {
		query = query.Where("created_at >= ?", filter.DateFrom)
	}
	if !filter.DateTo.IsZero() {
		query = query.Where("created_at <= ?", filter.DateTo)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("%s: %w", appErrors.ErrCtxBookingCountAll, err)
	}

	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = constants.DefaultPageLimit
	}
	offset := (filter.Page - 1) * filter.Limit

	var bookings []models.Booking
	if err := query.
		Preload("User").
		Preload("Tour").
		Preload("Schedule").
		Preload("Payments").
		Order("created_at DESC").
		Limit(filter.Limit).
		Offset(offset).
		Find(&bookings).Error; err != nil {
		return nil, 0, fmt.Errorf("%s: %w", appErrors.ErrCtxBookingFindAll, err)
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
