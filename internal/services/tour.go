package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"sun-booking-tours/internal/constants"
	appErrors "sun-booking-tours/internal/errors"
	"sun-booking-tours/internal/models"
	"sun-booking-tours/internal/repository"
	"sun-booking-tours/internal/utils"

	"gorm.io/gorm"
)

type TourForm struct {
	Title           string   `form:"title" binding:"required,max=500"`
	Description     string   `form:"description"`
	Price           float64  `form:"price" binding:"required,gt=0"`
	DurationDays    int      `form:"duration_days" binding:"required,gt=0"`
	Location        string   `form:"location" binding:"max=500"`
	MaxParticipants int      `form:"max_participants" binding:"required,gt=0"`
	Status          string   `form:"status" binding:"required"`
	CategoryIDs     []uint   `form:"category_ids"`
	ImageURLs       []string `form:"image_urls"`
}

type TourService struct {
	repo    repository.TourRepo
	catRepo repository.CategoryRepo
}

func NewTourService(repo repository.TourRepo, catRepo repository.CategoryRepo) *TourService {
	return &TourService{repo: repo, catRepo: catRepo}
}

func (s *TourService) ListTours(ctx context.Context, filter repository.TourFilter) ([]models.Tour, int64, error) {
	tours, total, err := s.repo.FindAll(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", appErrors.ErrCtxTourServiceList, err)
	}
	return tours, total, nil
}

func (s *TourService) GetTour(ctx context.Context, id uint) (*models.Tour, error) {
	tour, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrTourNotFound
		}
		return nil, fmt.Errorf("%s: %w", appErrors.ErrCtxTourServiceGet, err)
	}
	return tour, nil
}

func (s *TourService) CreateTour(ctx context.Context, form *TourForm) error {
	title := strings.TrimSpace(form.Title)
	if title == "" {
		return appErrors.NewAppError(http.StatusBadRequest, appErrors.ErrMsgTourTitleRequired)
	}

	if !isValidTourStatus(form.Status) {
		return appErrors.NewAppError(http.StatusBadRequest, appErrors.ErrMsgTourInvalidStatus)
	}

	slug := utils.Slugify(title)

	exists, err := s.repo.ExistsBySlug(ctx, slug)
	if err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxTourServiceCreateCheckSlug, err)
	}
	if exists {
		return appErrors.NewAppError(http.StatusConflict, appErrors.ErrMsgTourTitleDuplicate)
	}

	imagesJSON, _ := json.Marshal(filterNonEmpty(form.ImageURLs))

	tour := models.Tour{
		Title:           title,
		Slug:            slug,
		Description:     strings.TrimSpace(form.Description),
		Price:           form.Price,
		DurationDays:    form.DurationDays,
		Location:        strings.TrimSpace(form.Location),
		MaxParticipants: form.MaxParticipants,
		Images:          imagesJSON,
		Status:          form.Status,
	}

	if err := s.repo.Create(ctx, &tour); err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxTourServiceCreate, err)
	}

	if len(form.CategoryIDs) > 0 {
		if err := s.validateCategoryIDs(ctx, form.CategoryIDs); err != nil {
			return err
		}
		cats := make([]models.Category, len(form.CategoryIDs))
		for i, cid := range form.CategoryIDs {
			cats[i] = models.Category{ID: cid}
		}
		if err := s.repo.ReplaceCategories(ctx, &tour, cats); err != nil {
			return fmt.Errorf("%s: %w", appErrors.ErrCtxTourServiceCreate, err)
		}
	}

	return nil
}

func (s *TourService) UpdateTour(ctx context.Context, id uint, form *TourForm) error {
	tour, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return appErrors.ErrTourNotFound
		}
		return fmt.Errorf("%s: %w", appErrors.ErrCtxTourServiceUpdateFind, err)
	}

	title := strings.TrimSpace(form.Title)
	if title == "" {
		return appErrors.NewAppError(http.StatusBadRequest, appErrors.ErrMsgTourTitleRequired)
	}

	if !isValidTourStatus(form.Status) {
		return appErrors.NewAppError(http.StatusBadRequest, appErrors.ErrMsgTourInvalidStatus)
	}

	slug := utils.Slugify(title)

	slugExists, err := s.repo.ExistsBySlugExcluding(ctx, slug, id)
	if err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxTourServiceUpdateCheckSlug, err)
	}
	if slugExists {
		return appErrors.NewAppError(http.StatusConflict, appErrors.ErrMsgTourTitleDuplicate)
	}

	imagesJSON, _ := json.Marshal(filterNonEmpty(form.ImageURLs))

	tour.Title = title
	tour.Slug = slug
	tour.Description = strings.TrimSpace(form.Description)
	tour.Price = form.Price
	tour.DurationDays = form.DurationDays
	tour.Location = strings.TrimSpace(form.Location)
	tour.MaxParticipants = form.MaxParticipants
	tour.Images = imagesJSON
	tour.Status = form.Status

	if err := s.repo.Update(ctx, tour); err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxTourServiceUpdate, err)
	}

	if len(form.CategoryIDs) > 0 {
		if err := s.validateCategoryIDs(ctx, form.CategoryIDs); err != nil {
			return err
		}
	}
	cats := make([]models.Category, len(form.CategoryIDs))
	for i, cid := range form.CategoryIDs {
		cats[i] = models.Category{ID: cid}
	}
	if err := s.repo.ReplaceCategories(ctx, tour, cats); err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxTourServiceUpdate, err)
	}

	return nil
}

func (s *TourService) DeleteTour(ctx context.Context, id uint) error {
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return appErrors.ErrTourNotFound
		}
		return fmt.Errorf("%s: %w", appErrors.ErrCtxTourServiceDeleteFind, err)
	}

	hasBookings, err := s.repo.HasActiveBookings(ctx, id)
	if err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxTourServiceDeleteCheckBooks, err)
	}
	if hasBookings {
		return appErrors.NewAppError(http.StatusBadRequest, appErrors.ErrMsgTourCannotDeleteBooking)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxTourServiceDelete, err)
	}
	return nil
}

func (s *TourService) validateCategoryIDs(ctx context.Context, ids []uint) error {
	count, err := s.catRepo.CountByIDs(ctx, ids)
	if err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxTourServiceValidateCategories, err)
	}
	if int(count) != len(ids) {
		return appErrors.NewAppError(http.StatusBadRequest, appErrors.ErrMsgTourCategoryNotFound)
	}
	return nil
}

func isValidTourStatus(status string) bool {
	return status == constants.TourStatusDraft ||
		status == constants.TourStatusActive ||
		status == constants.TourStatusInactive
}

func filterNonEmpty(ss []string) []string {
	result := make([]string, 0, len(ss))
	for _, s := range ss {
		trimmed := strings.TrimSpace(s)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
