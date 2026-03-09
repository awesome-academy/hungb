package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"sun-booking-tours/internal/constants"
	appErrors "sun-booking-tours/internal/errors"
	"sun-booking-tours/internal/models"
	"sun-booking-tours/internal/repository"

	"gorm.io/gorm"
)

type ScheduleForm struct {
	TourID         uint     `form:"tour_id"`
	DepartureDate  string   `form:"departure_date" binding:"required"`
	ReturnDate     string   `form:"return_date" binding:"required"`
	AvailableSlots int      `form:"available_slots" binding:"required,gt=0"`
	PriceOverride  *float64 `form:"price_override"`
	Status         string   `form:"status" binding:"required"`
}

type ScheduleService struct {
	repo     repository.ScheduleRepo
	tourRepo repository.TourRepo
}

func NewScheduleService(repo repository.ScheduleRepo, tourRepo repository.TourRepo) *ScheduleService {
	return &ScheduleService{repo: repo, tourRepo: tourRepo}
}

func (s *ScheduleService) ListByTour(ctx context.Context, tourID uint) ([]models.TourSchedule, error) {
	schedules, err := s.repo.FindByTourID(ctx, tourID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", appErrors.ErrCtxScheduleServiceList, err)
	}
	return schedules, nil
}

func (s *ScheduleService) GetSchedule(ctx context.Context, id uint) (*models.TourSchedule, error) {
	schedule, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrScheduleNotFound
		}
		return nil, fmt.Errorf("%s: %w", appErrors.ErrCtxScheduleServiceGet, err)
	}
	return schedule, nil
}

func (s *ScheduleService) CreateSchedule(ctx context.Context, form *ScheduleForm) error {
	_, err := s.tourRepo.FindByID(ctx, form.TourID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return appErrors.NewAppError(http.StatusBadRequest, appErrors.ErrMsgScheduleTourNotFound)
		}
		return fmt.Errorf("%s: %w", appErrors.ErrCtxScheduleServiceCreate, err)
	}

	departure, err := parseDate(form.DepartureDate)
	if err != nil {
		return appErrors.NewAppError(http.StatusBadRequest, appErrors.ErrMsgScheduleDepartureDateReq)
	}
	returnDate, err := parseDate(form.ReturnDate)
	if err != nil {
		return appErrors.NewAppError(http.StatusBadRequest, appErrors.ErrMsgScheduleReturnDateReq)
	}

	if !returnDate.After(departure) {
		return appErrors.NewAppError(http.StatusBadRequest, appErrors.ErrMsgScheduleReturnAfterDepart)
	}

	if !isValidScheduleStatus(form.Status) {
		return appErrors.NewAppError(http.StatusBadRequest, appErrors.ErrMsgScheduleInvalidStatus)
	}

	schedule := models.TourSchedule{
		TourID:         form.TourID,
		DepartureDate:  departure,
		ReturnDate:     returnDate,
		AvailableSlots: form.AvailableSlots,
		PriceOverride:  form.PriceOverride,
		Status:         form.Status,
	}

	if err := s.repo.Create(ctx, &schedule); err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxScheduleServiceCreate, err)
	}
	return nil
}

func (s *ScheduleService) UpdateSchedule(ctx context.Context, id uint, form *ScheduleForm) error {
	schedule, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return appErrors.ErrScheduleNotFound
		}
		return fmt.Errorf("%s: %w", appErrors.ErrCtxScheduleServiceUpdateFind, err)
	}

	departure, err := parseDate(form.DepartureDate)
	if err != nil {
		return appErrors.NewAppError(http.StatusBadRequest, appErrors.ErrMsgScheduleDepartureDateReq)
	}
	returnDate, err := parseDate(form.ReturnDate)
	if err != nil {
		return appErrors.NewAppError(http.StatusBadRequest, appErrors.ErrMsgScheduleReturnDateReq)
	}

	if !returnDate.After(departure) {
		return appErrors.NewAppError(http.StatusBadRequest, appErrors.ErrMsgScheduleReturnAfterDepart)
	}

	if !isValidScheduleStatus(form.Status) {
		return appErrors.NewAppError(http.StatusBadRequest, appErrors.ErrMsgScheduleInvalidStatus)
	}

	schedule.DepartureDate = departure
	schedule.ReturnDate = returnDate
	schedule.AvailableSlots = form.AvailableSlots
	schedule.PriceOverride = form.PriceOverride
	schedule.Status = form.Status

	if err := s.repo.Update(ctx, schedule); err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxScheduleServiceUpdate, err)
	}
	return nil
}

func (s *ScheduleService) DeleteSchedule(ctx context.Context, id uint) error {
	schedule, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return appErrors.ErrScheduleNotFound
		}
		return fmt.Errorf("%s: %w", appErrors.ErrCtxScheduleServiceDeleteFind, err)
	}

	hasBookings, err := s.repo.HasBookings(ctx, schedule.ID)
	if err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxScheduleServiceDeleteCheck, err)
	}
	if hasBookings {
		return appErrors.NewAppError(http.StatusBadRequest, appErrors.ErrMsgScheduleCannotDeleteBooking)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxScheduleServiceDelete, err)
	}
	return nil
}

func isValidScheduleStatus(status string) bool {
	return status == constants.ScheduleStatusOpen ||
		status == constants.ScheduleStatusFull ||
		status == constants.ScheduleStatusCancelled
}

func parseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}
