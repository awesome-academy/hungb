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

type TourFilter struct {
	Status       string
	CategoryID   uint
	CategorySlug string
	Search       string
	MinPrice     float64
	MaxPrice     float64
	Location     string
	MinDuration  int
	MaxDuration  int
	SortBy       string
	SortOrder    string
	Page         int
	Limit        int
}

type TourRepo interface {
	FindAll(ctx context.Context, filter TourFilter) ([]models.Tour, int64, error)
	FindByID(ctx context.Context, id uint) (*models.Tour, error)
	FindBySlug(ctx context.Context, slug string) (*models.Tour, error)
	FindBySlugPublic(ctx context.Context, slug string) (*models.Tour, error)
	ExistsBySlug(ctx context.Context, slug string) (bool, error)
	ExistsBySlugExcluding(ctx context.Context, slug string, excludeID uint) (bool, error)
	Create(ctx context.Context, tour *models.Tour) error
	Update(ctx context.Context, tour *models.Tour) error
	Delete(ctx context.Context, id uint) error
	HasActiveBookings(ctx context.Context, tourID uint) (bool, error)
	ReplaceCategories(ctx context.Context, tour *models.Tour, categories []models.Category) error
	CountRatingsByTourID(ctx context.Context, tourID uint) (int64, error)
}

type tourRepository struct {
	db *gorm.DB
}

func NewTourRepository(db *gorm.DB) TourRepo {
	return &tourRepository{db: db}
}

func (r *tourRepository) FindAll(ctx context.Context, filter TourFilter) ([]models.Tour, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.Tour{})

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.CategorySlug != "" {
		query = query.Where("id IN (?)",
			r.db.Table("tour_categories").Select("tour_id").
				Where("category_id IN (?)",
					r.db.Table("categories").Select("id").Where("slug = ?", filter.CategorySlug),
				),
		)
	}
	if filter.CategoryID > 0 {
		query = query.Where("id IN (?)",
			r.db.Table("tour_categories").Select("tour_id").Where("category_id = ?", filter.CategoryID),
		)
	}
	if filter.Search != "" {
		like := "%" + filter.Search + "%"
		query = query.Where("title ILIKE ? OR description ILIKE ? OR location ILIKE ?", like, like, like)
	}
	if filter.MinPrice > 0 {
		query = query.Where("price >= ?", filter.MinPrice)
	}
	if filter.MaxPrice > 0 {
		query = query.Where("price <= ?", filter.MaxPrice)
	}
	if filter.Location != "" {
		query = query.Where("location ILIKE ?", "%"+filter.Location+"%")
	}
	if filter.MinDuration > 0 {
		query = query.Where("duration_days >= ?", filter.MinDuration)
	}
	if filter.MaxDuration > 0 {
		query = query.Where("duration_days <= ?", filter.MaxDuration)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("%s: %w", appErrors.ErrCtxTourCount, err)
	}

	if filter.Limit <= 0 {
		filter.Limit = constants.DefaultPageLimit
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	offset := (filter.Page - 1) * filter.Limit

	var tours []models.Tour

	sortCol := "created_at"
	switch filter.SortBy {
	case "price", "avg_rating", "duration_days":
		sortCol = filter.SortBy
	}
	sortDir := "DESC"
	if filter.SortOrder == "asc" {
		sortDir = "ASC"
	}

	if err := query.
		Preload("Categories").
		Preload("Schedules", func(db *gorm.DB) *gorm.DB {
			return db.
				Where("status = ? AND departure_date >= ?", constants.ScheduleStatusOpen, time.Now()).
				Order("departure_date ASC")
		}).
		Order(sortCol + " " + sortDir).
		Limit(filter.Limit).
		Offset(offset).
		Find(&tours).Error; err != nil {
		return nil, 0, fmt.Errorf("%s: %w", appErrors.ErrCtxTourFindAll, err)
	}
	return tours, total, nil
}

func (r *tourRepository) FindByID(ctx context.Context, id uint) (*models.Tour, error) {
	var tour models.Tour
	if err := r.db.WithContext(ctx).
		Preload("Categories").
		Preload("Schedules", func(db *gorm.DB) *gorm.DB {
			return db.Order("departure_date ASC")
		}).
		First(&tour, id).Error; err != nil {
		return nil, fmt.Errorf("%s: %w", appErrors.ErrCtxTourFindByID, err)
	}
	return &tour, nil
}

func (r *tourRepository) FindBySlug(ctx context.Context, slug string) (*models.Tour, error) {
	var tour models.Tour
	if err := r.db.WithContext(ctx).
		Preload("Categories").
		Where("slug = ?", slug).
		First(&tour).Error; err != nil {
		return nil, fmt.Errorf("%s: %w", appErrors.ErrCtxTourFindBySlug, err)
	}
	return &tour, nil
}

func (r *tourRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Tour{}).
		Where("slug = ?", slug).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("%s: %w", appErrors.ErrCtxTourCheckSlugExists, err)
	}
	return count > 0, nil
}

func (r *tourRepository) ExistsBySlugExcluding(ctx context.Context, slug string, excludeID uint) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Tour{}).
		Where("slug = ? AND id != ?", slug, excludeID).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("%s: %w", appErrors.ErrCtxTourCheckSlugExcluding, err)
	}
	return count > 0, nil
}

func (r *tourRepository) Create(ctx context.Context, tour *models.Tour) error {
	if err := r.db.WithContext(ctx).Create(tour).Error; err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxTourCreate, err)
	}
	return nil
}

func (r *tourRepository) Update(ctx context.Context, tour *models.Tour) error {
	if err := r.db.WithContext(ctx).Save(tour).Error; err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxTourUpdate, err)
	}
	return nil
}

func (r *tourRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&models.Tour{}, id).Error; err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxTourDelete, err)
	}
	return nil
}

func (r *tourRepository) HasActiveBookings(ctx context.Context, tourID uint) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Booking{}).
		Where("tour_id = ? AND status IN ?", tourID, []string{constants.BookingStatusPending, constants.BookingStatusConfirmed}).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("%s: %w", appErrors.ErrCtxTourHasActiveBookings, err)
	}
	return count > 0, nil
}

func (r *tourRepository) ReplaceCategories(ctx context.Context, tour *models.Tour, categories []models.Category) error {
	if err := r.db.WithContext(ctx).Model(tour).Association("Categories").Replace(categories); err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxTourReplaceCategories, err)
	}
	return nil
}

func (r *tourRepository) FindBySlugPublic(ctx context.Context, slug string) (*models.Tour, error) {
	var tour models.Tour
	if err := r.db.WithContext(ctx).
		Preload("Categories").
		Preload("Schedules", func(db *gorm.DB) *gorm.DB {
			return db.
				Where("status = ? AND departure_date >= ?", constants.ScheduleStatusOpen, time.Now()).
				Order("departure_date ASC")
		}).
		Where("slug = ? AND status = ?", slug, constants.TourStatusActive).
		First(&tour).Error; err != nil {
		return nil, fmt.Errorf("%s: %w", appErrors.ErrCtxPublicTourFindBySlug, err)
	}
	return &tour, nil
}

func (r *tourRepository) CountRatingsByTourID(ctx context.Context, tourID uint) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Rating{}).
		Where("tour_id = ?", tourID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("%s: %w", appErrors.ErrCtxPublicTourCountRatings, err)
	}
	return count, nil
}
