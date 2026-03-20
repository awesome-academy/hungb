package repository

import (
	"context"
	"fmt"

	appErrors "sun-booking-tours/internal/errors"
	"sun-booking-tours/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RatingRepo interface {
	Upsert(ctx context.Context, rating *models.Rating) error
	FindByUserAndTour(ctx context.Context, userID, tourID uint) (*models.Rating, error)
	FindByTourID(ctx context.Context, tourID uint, page, limit int) ([]models.Rating, int64, error)
	CalcAvgByTourID(ctx context.Context, tourID uint) (float64, error)
}

type ratingRepository struct {
	db *gorm.DB
}

func NewRatingRepository(db *gorm.DB) RatingRepo {
	return &ratingRepository{db: db}
}

func (r *ratingRepository) Upsert(ctx context.Context, rating *models.Rating) error {
	if err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "tour_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"score", "comment", "updated_at"}),
		}).
		Create(rating).Error; err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxRatingUpsert, err)
	}
	return nil
}

func (r *ratingRepository) FindByUserAndTour(ctx context.Context, userID, tourID uint) (*models.Rating, error) {
	var rating models.Rating
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND tour_id = ?", userID, tourID).
		First(&rating).Error; err != nil {
		return nil, fmt.Errorf("%s: %w", appErrors.ErrCtxRatingFindByUserTour, err)
	}
	return &rating, nil
}

func (r *ratingRepository) FindByTourID(ctx context.Context, tourID uint, page, limit int) ([]models.Rating, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&models.Rating{}).
		Where("tour_id = ?", tourID).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("%s: %w", appErrors.ErrCtxRatingCountByTour, err)
	}

	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	var ratings []models.Rating
	if err := r.db.WithContext(ctx).
		Preload("User").
		Where("tour_id = ?", tourID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&ratings).Error; err != nil {
		return nil, 0, fmt.Errorf("%s: %w", appErrors.ErrCtxRatingFindByTour, err)
	}
	return ratings, total, nil
}

func (r *ratingRepository) CalcAvgByTourID(ctx context.Context, tourID uint) (float64, error) {
	var avg *float64
	if err := r.db.WithContext(ctx).Model(&models.Rating{}).
		Where("tour_id = ?", tourID).
		Select("AVG(score)").
		Scan(&avg).Error; err != nil {
		return 0, fmt.Errorf("%s: %w", appErrors.ErrCtxRatingCalcAvg, err)
	}
	if avg == nil {
		return 0, nil
	}
	return *avg, nil
}
