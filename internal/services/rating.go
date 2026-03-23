package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strings"

	"sun-booking-tours/internal/constants"
	appErrors "sun-booking-tours/internal/errors"
	"sun-booking-tours/internal/models"
	"sun-booking-tours/internal/repository"

	"gorm.io/gorm"
)

type RatingService struct {
	ratingRepo repository.RatingRepo
	tourRepo   repository.TourRepo
}

func NewRatingService(ratingRepo repository.RatingRepo, tourRepo repository.TourRepo) *RatingService {
	return &RatingService{ratingRepo: ratingRepo, tourRepo: tourRepo}
}

type RatingInput struct {
	Score   int
	Comment string
}

func (s *RatingService) RateOrUpdate(ctx context.Context, userID, tourID uint, input RatingInput) (isNew bool, err error) {
	if input.Score < constants.RatingMinScore || input.Score > constants.RatingMaxScore {
		return false, appErrors.ErrInvalidScore
	}

	if _, err := s.tourRepo.FindByIDPublic(ctx, tourID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, appErrors.ErrTourNotFound
		}
		return false, fmt.Errorf("%s: %w", appErrors.ErrCtxRatingServiceRate, err)
	}

	_, findErr := s.ratingRepo.FindByUserAndTour(ctx, userID, tourID)
	if findErr != nil && !errors.Is(findErr, gorm.ErrRecordNotFound) {
		return false, fmt.Errorf("%s: %w", appErrors.ErrCtxRatingServiceRate, findErr)
	}
	isNew = errors.Is(findErr, gorm.ErrRecordNotFound)

	rating := &models.Rating{
		UserID:  userID,
		TourID:  tourID,
		Score:   input.Score,
		Comment: strings.TrimSpace(input.Comment),
	}

	if err := s.ratingRepo.Upsert(ctx, rating); err != nil {
		return false, fmt.Errorf("%s: %w", appErrors.ErrCtxRatingServiceRate, err)
	}

	avg, err := s.ratingRepo.CalcAvgByTourID(ctx, tourID)
	if err != nil {
		return isNew, fmt.Errorf("%s: %w", appErrors.ErrCtxRatingServiceRate, err)
	}
	avg = math.Round(avg*100) / 100
	if err := s.tourRepo.UpdateAvgRating(ctx, tourID, avg); err != nil {
		slog.Error(appErrors.ErrCtxRatingUpdateTourAvg, "context", appErrors.ErrCtxRatingUpdateTourAvg, "tour_id", tourID, "error", err)
	}

	return isNew, nil
}

func (s *RatingService) GetUserRating(ctx context.Context, userID, tourID uint) (*models.Rating, error) {
	if userID == 0 {
		return nil, nil
	}
	rating, err := s.ratingRepo.FindByUserAndTour(ctx, userID, tourID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("%s: %w", appErrors.ErrCtxRatingServiceRate, err)
	}
	return rating, nil
}

func (s *RatingService) ListByTour(ctx context.Context, tourID uint, page, limit int) ([]models.Rating, int64, error) {
	ratings, total, err := s.ratingRepo.FindByTourID(ctx, tourID, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", appErrors.ErrCtxRatingServiceList, err)
	}
	return ratings, total, nil
}
