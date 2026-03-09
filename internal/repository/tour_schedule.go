package repository

import (
	"context"
	"fmt"

	"sun-booking-tours/internal/constants"
	appErrors "sun-booking-tours/internal/errors"
	"sun-booking-tours/internal/models"

	"gorm.io/gorm"
)

type ScheduleRepo interface {
	FindByTourID(ctx context.Context, tourID uint) ([]models.TourSchedule, error)
	FindByID(ctx context.Context, id uint) (*models.TourSchedule, error)
	Create(ctx context.Context, schedule *models.TourSchedule) error
	Update(ctx context.Context, schedule *models.TourSchedule) error
	Delete(ctx context.Context, id uint) error
	HasBookings(ctx context.Context, scheduleID uint) (bool, error)
}

type scheduleRepository struct {
	db *gorm.DB
}

func NewScheduleRepository(db *gorm.DB) ScheduleRepo {
	return &scheduleRepository{db: db}
}

func (r *scheduleRepository) FindByTourID(ctx context.Context, tourID uint) ([]models.TourSchedule, error) {
	var schedules []models.TourSchedule
	if err := r.db.WithContext(ctx).
		Where("tour_id = ?", tourID).
		Order("departure_date ASC").
		Find(&schedules).Error; err != nil {
		return nil, fmt.Errorf("%s: %w", appErrors.ErrCtxScheduleFindByTour, err)
	}
	return schedules, nil
}

func (r *scheduleRepository) FindByID(ctx context.Context, id uint) (*models.TourSchedule, error) {
	var schedule models.TourSchedule
	if err := r.db.WithContext(ctx).
		Preload("Tour").
		First(&schedule, id).Error; err != nil {
		return nil, fmt.Errorf("%s: %w", appErrors.ErrCtxScheduleFindByID, err)
	}
	return &schedule, nil
}

func (r *scheduleRepository) Create(ctx context.Context, schedule *models.TourSchedule) error {
	if err := r.db.WithContext(ctx).Create(schedule).Error; err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxScheduleCreate, err)
	}
	return nil
}

func (r *scheduleRepository) Update(ctx context.Context, schedule *models.TourSchedule) error {
	if err := r.db.WithContext(ctx).Save(schedule).Error; err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxScheduleUpdate, err)
	}
	return nil
}

func (r *scheduleRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&models.TourSchedule{}, id).Error; err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxScheduleDelete, err)
	}
	return nil
}

func (r *scheduleRepository) HasBookings(ctx context.Context, scheduleID uint) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Booking{}).
		Where("schedule_id = ? AND status IN ?", scheduleID, []string{constants.BookingStatusPending, constants.BookingStatusConfirmed}).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("%s: %w", appErrors.ErrCtxScheduleHasBookings, err)
	}
	return count > 0, nil
}
