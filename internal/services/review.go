package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"sun-booking-tours/internal/constants"
	appErrors "sun-booking-tours/internal/errors"
	"sun-booking-tours/internal/models"
	"sun-booking-tours/internal/repository"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type ReviewService struct {
	db       *gorm.DB
	repo     repository.ReviewRepo
	likeRepo repository.ReviewLikeRepo
	cmtRepo  repository.CommentRepo
}

func NewReviewService(db *gorm.DB, repo repository.ReviewRepo, likeRepo repository.ReviewLikeRepo, cmtRepo repository.CommentRepo) *ReviewService {
	return &ReviewService{db: db, repo: repo, likeRepo: likeRepo, cmtRepo: cmtRepo}
}

type ReviewCreateInput struct {
	Title   string
	Content string
	Type    string
	Images  []string
}

func (s *ReviewService) CreateReview(ctx context.Context, userID uint, input ReviewCreateInput) (*models.Review, error) {
	if err := s.validateReviewInput(input); err != nil {
		return nil, err
	}

	var imagesJSON datatypes.JSON
	if len(input.Images) > 0 {
		b, _ := json.Marshal(input.Images)
		imagesJSON = datatypes.JSON(b)
	}

	review := &models.Review{
		UserID:  userID,
		Title:   strings.TrimSpace(input.Title),
		Content: strings.TrimSpace(input.Content),
		Type:    input.Type,
		Status:  constants.ReviewStatusPending,
		Images:  imagesJSON,
	}

	if err := s.repo.Create(ctx, review); err != nil {
		return nil, fmt.Errorf("%s: %w", appErrors.ErrCtxReviewServiceCreate, err)
	}
	return review, nil
}

func (s *ReviewService) GetReview(ctx context.Context, id uint) (*models.Review, error) {
	review, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrReviewNotFound
		}
		return nil, fmt.Errorf("%s: %w", appErrors.ErrCtxReviewServiceGet, err)
	}
	return review, nil
}

func (s *ReviewService) GetPublicReview(ctx context.Context, id uint) (*models.Review, []models.Comment, error) {
	review, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, appErrors.ErrReviewNotFound
		}
		return nil, nil, fmt.Errorf("%s: %w", appErrors.ErrCtxReviewServiceGet, err)
	}
	if review.Status != constants.ReviewStatusApproved {
		return nil, nil, appErrors.ErrReviewNotFound
	}

	comments, err := s.cmtRepo.FindByReviewID(ctx, id)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", appErrors.ErrCtxReviewServiceGet, err)
	}

	return review, comments, nil
}

func (s *ReviewService) ListPublicReviews(ctx context.Context, filter repository.ReviewFilter) ([]models.Review, int64, error) {
	reviews, total, err := s.repo.FindAllPublic(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", appErrors.ErrCtxReviewServiceList, err)
	}
	return reviews, total, nil
}

func (s *ReviewService) ListMyReviews(ctx context.Context, userID uint, page, limit int) ([]models.Review, int64, error) {
	reviews, total, err := s.repo.FindByUserID(ctx, userID, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", appErrors.ErrCtxReviewServiceMyList, err)
	}
	return reviews, total, nil
}

func (s *ReviewService) UpdateReview(ctx context.Context, id, userID uint, input ReviewCreateInput) error {
	if err := s.validateReviewInput(input); err != nil {
		return err
	}

	review, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return appErrors.ErrReviewNotFound
		}
		return fmt.Errorf("%s: %w", appErrors.ErrCtxReviewServiceUpdate, err)
	}
	if review.UserID != userID {
		return appErrors.ErrReviewNotOwner
	}

	review.Title = strings.TrimSpace(input.Title)
	review.Content = strings.TrimSpace(input.Content)
	review.Type = input.Type
	review.Status = constants.ReviewStatusPending

	if len(input.Images) > 0 {
		b, _ := json.Marshal(input.Images)
		review.Images = datatypes.JSON(b)
	} else {
		review.Images = nil
	}

	if err := s.repo.Update(ctx, review); err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxReviewServiceUpdate, err)
	}
	return nil
}

func (s *ReviewService) DeleteReview(ctx context.Context, id, userID uint) error {
	review, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return appErrors.ErrReviewNotFound
		}
		return fmt.Errorf("%s: %w", appErrors.ErrCtxReviewServiceDelete, err)
	}
	if review.UserID != userID {
		return appErrors.ErrReviewNotOwner
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxReviewServiceDelete, err)
	}
	return nil
}

func (s *ReviewService) ToggleLike(ctx context.Context, userID, reviewID uint) (liked bool, err error) {
	exists, err := s.likeRepo.Exists(ctx, userID, reviewID)
	if err != nil {
		return false, fmt.Errorf("%s: %w", appErrors.ErrCtxReviewServiceToggleLike, err)
	}

	if exists {
		if err := s.likeRepo.Delete(ctx, userID, reviewID); err != nil {
			return false, fmt.Errorf("%s: %w", appErrors.ErrCtxReviewServiceToggleLike, err)
		}
		_ = s.repo.IncrementLikeCount(ctx, reviewID, -1)
		return false, nil
	}

	like := &models.ReviewLike{UserID: userID, ReviewID: reviewID}
	if err := s.likeRepo.Create(ctx, like); err != nil {
		return false, fmt.Errorf("%s: %w", appErrors.ErrCtxReviewServiceToggleLike, err)
	}
	_ = s.repo.IncrementLikeCount(ctx, reviewID, 1)
	return true, nil
}

func (s *ReviewService) HasUserLiked(ctx context.Context, userID, reviewID uint) bool {
	if userID == 0 {
		return false
	}
	exists, _ := s.likeRepo.Exists(ctx, userID, reviewID)
	return exists
}

func (s *ReviewService) AddComment(ctx context.Context, userID, reviewID uint, parentID *uint, content string) error {
	content = strings.TrimSpace(content)
	if content == "" {
		return appErrors.ErrInvalidInput
	}

	if parentID != nil {
		parent, err := s.cmtRepo.FindByID(ctx, *parentID)
		if err != nil {
			return appErrors.ErrCommentNotFound
		}
		if parent.ReviewID != reviewID {
			return appErrors.ErrInvalidInput
		}
	}

	comment := &models.Comment{
		UserID:   userID,
		ReviewID: reviewID,
		ParentID: parentID,
		Content:  content,
	}
	if err := s.cmtRepo.Create(ctx, comment); err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxReviewServiceAddComment, err)
	}
	return nil
}

func (s *ReviewService) DeleteComment(ctx context.Context, commentID, userID uint, isAdmin bool) error {
	comment, err := s.cmtRepo.FindByID(ctx, commentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return appErrors.ErrCommentNotFound
		}
		return fmt.Errorf("%s: %w", appErrors.ErrCtxReviewServiceDelComment, err)
	}
	if !isAdmin && comment.UserID != userID {
		return appErrors.ErrCommentNotOwner
	}

	if err := s.cmtRepo.Delete(ctx, commentID); err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxReviewServiceDelComment, err)
	}
	return nil
}

func (s *ReviewService) AdminListReviews(ctx context.Context, filter repository.ReviewFilter) ([]models.Review, int64, error) {
	reviews, total, err := s.repo.FindAll(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", appErrors.ErrCtxReviewServiceAdminList, err)
	}
	return reviews, total, nil
}

func (s *ReviewService) AdminApproveReview(ctx context.Context, id uint) error {
	review, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return appErrors.ErrReviewNotFound
		}
		return fmt.Errorf("%s: %w", appErrors.ErrCtxReviewServiceApprove, err)
	}
	if review.Status == constants.ReviewStatusApproved {
		return appErrors.ErrReviewCannotApprove
	}

	if err := s.repo.UpdateStatus(ctx, id, constants.ReviewStatusApproved); err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxReviewServiceApprove, err)
	}
	return nil
}

func (s *ReviewService) AdminRejectReview(ctx context.Context, id uint) error {
	review, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return appErrors.ErrReviewNotFound
		}
		return fmt.Errorf("%s: %w", appErrors.ErrCtxReviewServiceReject, err)
	}
	if review.Status == constants.ReviewStatusRejected {
		return appErrors.ErrReviewCannotReject
	}

	if err := s.repo.UpdateStatus(ctx, id, constants.ReviewStatusRejected); err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxReviewServiceReject, err)
	}
	return nil
}

func (s *ReviewService) validateReviewInput(input ReviewCreateInput) error {
	if strings.TrimSpace(input.Title) == "" {
		return appErrors.NewAppError(400, appErrors.ErrInvalidInput.Message)
	}
	if strings.TrimSpace(input.Content) == "" {
		return appErrors.NewAppError(400, appErrors.ErrInvalidInput.Message)
	}
	if len(strings.TrimSpace(input.Content)) < 10 {
		return appErrors.NewAppError(400, appErrors.ErrInvalidInput.Message)
	}
	switch input.Type {
	case constants.ReviewTypePlace, constants.ReviewTypeFood, constants.ReviewTypeNews:
	default:
		return appErrors.NewAppError(400, appErrors.ErrInvalidInput.Message)
	}
	return nil
}
