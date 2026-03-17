package services

import (
	"context"
	"errors"
	"fmt"

	"sun-booking-tours/internal/constants"
	appErrors "sun-booking-tours/internal/errors"
	"sun-booking-tours/internal/models"
	"sun-booking-tours/internal/repository"

	"gorm.io/gorm"
)

type BookingService struct {
	db           *gorm.DB
	bookingRepo  repository.BookingRepo
	scheduleRepo repository.ScheduleRepo
}

func NewBookingService(db *gorm.DB, bookingRepo repository.BookingRepo, scheduleRepo repository.ScheduleRepo) *BookingService {
	return &BookingService{db: db, bookingRepo: bookingRepo, scheduleRepo: scheduleRepo}
}

func (s *BookingService) CreateBooking(ctx context.Context, userID, tourID, scheduleID uint, numParticipants int, tourPrice float64, note string) (*models.Booking, error) {
	var booking *models.Booking

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var schedule models.TourSchedule
		if err := tx.First(&schedule, scheduleID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return appErrors.ErrScheduleNotFound
			}
			return fmt.Errorf("%s: %w", appErrors.ErrCtxBookingServiceCreate, err)
		}

		if schedule.Status != constants.ScheduleStatusOpen {
			return appErrors.ErrScheduleNotOpen
		}
		if schedule.AvailableSlots < numParticipants {
			return appErrors.ErrNotEnoughSlots
		}

		unitPrice := tourPrice
		if schedule.PriceOverride != nil {
			unitPrice = *schedule.PriceOverride
		}
		totalPrice := unitPrice * float64(numParticipants)

		result := tx.Model(&models.TourSchedule{}).
			Where("id = ? AND status = ? AND available_slots >= ?", scheduleID, constants.ScheduleStatusOpen, numParticipants).
			Update("available_slots", gorm.Expr("available_slots - ?", numParticipants))
		if result.Error != nil {
			return fmt.Errorf("%s: %w", appErrors.ErrCtxScheduleUpdateSlots, result.Error)
		}
		if result.RowsAffected == 0 {
			return appErrors.ErrNotEnoughSlots
		}

		var updatedSchedule models.TourSchedule
		if err := tx.First(&updatedSchedule, scheduleID).Error; err != nil {
			return fmt.Errorf("%s: %w", appErrors.ErrCtxBookingServiceCreate, err)
		}
		if updatedSchedule.AvailableSlots == 0 {
			tx.Model(&models.TourSchedule{}).Where("id = ?", scheduleID).Update("status", constants.ScheduleStatusFull)
		}

		booking = &models.Booking{
			UserID:          userID,
			TourID:          tourID,
			ScheduleID:      scheduleID,
			NumParticipants: numParticipants,
			TotalPrice:      totalPrice,
			Status:          constants.BookingStatusPending,
			Note:            note,
		}
		if err := tx.Create(booking).Error; err != nil {
			return fmt.Errorf("%s: %w", appErrors.ErrCtxBookingCreate, err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return booking, nil
}

func (s *BookingService) GetBooking(ctx context.Context, id, userID uint) (*models.Booking, error) {
	booking, err := s.bookingRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrBookingNotFound
		}
		return nil, fmt.Errorf("%s: %w", appErrors.ErrCtxBookingServiceGet, err)
	}
	if booking.UserID != userID {
		return nil, appErrors.ErrBookingNotFound
	}
	return booking, nil
}

func (s *BookingService) ListMyBookings(ctx context.Context, userID uint, page, limit int) ([]models.Booking, int64, error) {
	bookings, total, err := s.bookingRepo.FindByUserID(ctx, userID, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", appErrors.ErrCtxBookingServiceList, err)
	}
	return bookings, total, nil
}

func (s *BookingService) CancelBooking(ctx context.Context, id, userID uint) error {
	booking, err := s.bookingRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return appErrors.ErrBookingNotFound
		}
		return fmt.Errorf("%s: %w", appErrors.ErrCtxBookingServiceCancel, err)
	}
	if booking.UserID != userID {
		return appErrors.ErrBookingNotFound
	}
	if booking.Status != constants.BookingStatusPending && booking.Status != constants.BookingStatusConfirmed {
		return appErrors.ErrBookingCannotCancel
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Booking{}).Where("id = ?", id).Update("status", constants.BookingStatusCancelled).Error; err != nil {
			return fmt.Errorf("%s: %w", appErrors.ErrCtxBookingUpdateStatus, err)
		}

		result := tx.Model(&models.TourSchedule{}).
			Where("id = ?", booking.ScheduleID).
			Update("available_slots", gorm.Expr("available_slots + ?", booking.NumParticipants))
		if result.Error != nil {
			return fmt.Errorf("%s: %w", appErrors.ErrCtxScheduleUpdateSlots, result.Error)
		}

		tx.Model(&models.TourSchedule{}).
			Where("id = ? AND status = ?", booking.ScheduleID, constants.ScheduleStatusFull).
			Update("status", constants.ScheduleStatusOpen)

		if len(booking.Payments) > 0 {
			tx.Model(&models.Payment{}).
				Where("booking_id = ? AND status = ?", id, constants.PaymentStatusSuccess).
				Update("status", constants.PaymentStatusRefunded)
		}

		return nil
	})
}

func (s *BookingService) ListAllBookings(ctx context.Context, filter repository.BookingFilter) ([]models.Booking, int64, error) {
	bookings, total, err := s.bookingRepo.FindAll(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", appErrors.ErrCtxBookingServiceList, err)
	}
	return bookings, total, nil
}

func (s *BookingService) GetBookingByID(ctx context.Context, id uint) (*models.Booking, error) {
	booking, err := s.bookingRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrBookingNotFound
		}
		return nil, fmt.Errorf("%s: %w", appErrors.ErrCtxBookingServiceGet, err)
	}
	return booking, nil
}

func (s *BookingService) AdminConfirmBooking(ctx context.Context, id uint) error {
	booking, err := s.bookingRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return appErrors.ErrBookingNotFound
		}
		return fmt.Errorf("%s: %w", appErrors.ErrCtxBookingServiceConfirm, err)
	}
	if booking.Status != constants.BookingStatusPending {
		return appErrors.ErrBookingCannotConfirm
	}
	return s.bookingRepo.UpdateStatus(ctx, id, constants.BookingStatusConfirmed)
}

func (s *BookingService) AdminCancelBooking(ctx context.Context, id uint) error {
	booking, err := s.bookingRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return appErrors.ErrBookingNotFound
		}
		return fmt.Errorf("%s: %w", appErrors.ErrCtxBookingServiceCancel, err)
	}
	if booking.Status != constants.BookingStatusPending && booking.Status != constants.BookingStatusConfirmed {
		return appErrors.ErrBookingCannotCancel
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Booking{}).Where("id = ?", id).Update("status", constants.BookingStatusCancelled).Error; err != nil {
			return fmt.Errorf("%s: %w", appErrors.ErrCtxBookingUpdateStatus, err)
		}

		tx.Model(&models.TourSchedule{}).
			Where("id = ?", booking.ScheduleID).
			Update("available_slots", gorm.Expr("available_slots + ?", booking.NumParticipants))

		tx.Model(&models.TourSchedule{}).
			Where("id = ? AND status = ?", booking.ScheduleID, constants.ScheduleStatusFull).
			Update("status", constants.ScheduleStatusOpen)

		if len(booking.Payments) > 0 {
			tx.Model(&models.Payment{}).
				Where("booking_id = ? AND status = ?", id, constants.PaymentStatusSuccess).
				Update("status", constants.PaymentStatusRefunded)
		}

		return nil
	})
}

func (s *BookingService) AdminCompleteBooking(ctx context.Context, id uint) error {
	booking, err := s.bookingRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return appErrors.ErrBookingNotFound
		}
		return fmt.Errorf("%s: %w", appErrors.ErrCtxBookingServiceComplete, err)
	}
	if booking.Status != constants.BookingStatusConfirmed {
		return appErrors.ErrBookingCannotComplete
	}
	return s.bookingRepo.UpdateStatus(ctx, id, constants.BookingStatusCompleted)
}
