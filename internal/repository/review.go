package repository

import (
	"context"
	"fmt"

	"sun-booking-tours/internal/constants"
	appErrors "sun-booking-tours/internal/errors"
	"sun-booking-tours/internal/models"

	"gorm.io/gorm"
)

type ReviewFilter struct {
	Status  string
	Type    string
	UserID  uint
	Keyword string
	Sort    string
	Page    int
	Limit   int
}

type ReviewRepo interface {
	Create(ctx context.Context, review *models.Review) error
	FindByID(ctx context.Context, id uint) (*models.Review, error)
	FindByUserID(ctx context.Context, userID uint, page, limit int) ([]models.Review, int64, error)
	FindAllPublic(ctx context.Context, filter ReviewFilter) ([]models.Review, int64, error)
	FindAll(ctx context.Context, filter ReviewFilter) ([]models.Review, int64, error)
	Update(ctx context.Context, review *models.Review) error
	Delete(ctx context.Context, id uint) error
	UpdateStatus(ctx context.Context, id uint, status string) error
	IncrementLikeCount(ctx context.Context, id uint, delta int) error
}

type CommentRepo interface {
	Create(ctx context.Context, comment *models.Comment) error
	FindByReviewID(ctx context.Context, reviewID uint) ([]models.Comment, error)
	FindByID(ctx context.Context, id uint) (*models.Comment, error)
	Delete(ctx context.Context, id uint) error
}

type ReviewLikeRepo interface {
	Exists(ctx context.Context, userID, reviewID uint) (bool, error)
	FindByUserAndReviewIDs(ctx context.Context, userID uint, reviewIDs []uint) ([]uint, error)
	Create(ctx context.Context, like *models.ReviewLike) error
	Delete(ctx context.Context, userID, reviewID uint) error
}

type reviewRepository struct {
	db *gorm.DB
}

func NewReviewRepository(db *gorm.DB) ReviewRepo {
	return &reviewRepository{db: db}
}

func (r *reviewRepository) Create(ctx context.Context, review *models.Review) error {
	if err := r.db.WithContext(ctx).Create(review).Error; err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxReviewCreate, err)
	}
	return nil
}

func (r *reviewRepository) FindByID(ctx context.Context, id uint) (*models.Review, error) {
	var review models.Review
	if err := r.db.WithContext(ctx).
		Preload("User").
		First(&review, id).Error; err != nil {
		return nil, fmt.Errorf("%s: %w", appErrors.ErrCtxReviewFindByID, err)
	}
	return &review, nil
}

func (r *reviewRepository) FindByUserID(ctx context.Context, userID uint, page, limit int) ([]models.Review, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&models.Review{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("%s: %w", appErrors.ErrCtxReviewCountByUser, err)
	}

	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	var reviews []models.Review
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&reviews).Error; err != nil {
		return nil, 0, fmt.Errorf("%s: %w", appErrors.ErrCtxReviewFindByUser, err)
	}
	return reviews, total, nil
}

func (r *reviewRepository) FindAllPublic(ctx context.Context, filter ReviewFilter) ([]models.Review, int64, error) {
	baseScope := func(db *gorm.DB) *gorm.DB {
		db = db.Where("status = ?", constants.ReviewStatusApproved)
		if filter.Type != "" {
			db = db.Where("type = ?", filter.Type)
		}
		if filter.Keyword != "" {
			kw := "%" + filter.Keyword + "%"
			db = db.Where("(title ILIKE ? OR content ILIKE ?)", kw, kw)
		}
		return db
	}

	var total int64
	if err := r.db.WithContext(ctx).Model(&models.Review{}).Scopes(baseScope).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("%s: %w", appErrors.ErrCtxReviewCountPublic, err)
	}

	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = constants.DefaultPageLimit
	}
	offset := (filter.Page - 1) * filter.Limit

	orderClause := "created_at DESC"
	if filter.Sort == "most_liked" {
		orderClause = "like_count DESC, created_at DESC"
	}

	var reviews []models.Review
	if err := r.db.WithContext(ctx).
		Preload("User").
		Scopes(baseScope).
		Order(orderClause).
		Limit(filter.Limit).
		Offset(offset).
		Find(&reviews).Error; err != nil {
		return nil, 0, fmt.Errorf("%s: %w", appErrors.ErrCtxReviewFindAllPublic, err)
	}
	return reviews, total, nil
}

func (r *reviewRepository) FindAll(ctx context.Context, filter ReviewFilter) ([]models.Review, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.Review{})

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}
	if filter.UserID > 0 {
		query = query.Where("user_id = ?", filter.UserID)
	}
	if filter.Keyword != "" {
		kw := "%" + filter.Keyword + "%"
		query = query.Where("(title ILIKE ? OR content ILIKE ?)", kw, kw)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("%s: %w", appErrors.ErrCtxReviewCountAll, err)
	}

	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = constants.DefaultPageLimit
	}
	offset := (filter.Page - 1) * filter.Limit

	var reviews []models.Review
	if err := query.
		Preload("User").
		Order("created_at DESC").
		Limit(filter.Limit).
		Offset(offset).
		Find(&reviews).Error; err != nil {
		return nil, 0, fmt.Errorf("%s: %w", appErrors.ErrCtxReviewFindAll, err)
	}
	return reviews, total, nil
}

func (r *reviewRepository) Update(ctx context.Context, review *models.Review) error {
	if err := r.db.WithContext(ctx).Save(review).Error; err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxReviewUpdate, err)
	}
	return nil
}

func (r *reviewRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&models.Review{}, id).Error; err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxReviewDelete, err)
	}
	return nil
}

func (r *reviewRepository) UpdateStatus(ctx context.Context, id uint, status string) error {
	if err := r.db.WithContext(ctx).
		Model(&models.Review{}).
		Where("id = ?", id).
		Update("status", status).Error; err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxReviewUpdateStatus, err)
	}
	return nil
}

func (r *reviewRepository) IncrementLikeCount(ctx context.Context, id uint, delta int) error {
	if err := r.db.WithContext(ctx).
		Model(&models.Review{}).
		Where("id = ?", id).
		UpdateColumn("like_count", gorm.Expr("like_count + ?", delta)).Error; err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxReviewUpdateLikeCount, err)
	}
	return nil
}

type commentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) CommentRepo {
	return &commentRepository{db: db}
}

func (r *commentRepository) Create(ctx context.Context, comment *models.Comment) error {
	if err := r.db.WithContext(ctx).Create(comment).Error; err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxCommentCreate, err)
	}
	return nil
}

func (r *commentRepository) FindByReviewID(ctx context.Context, reviewID uint) ([]models.Comment, error) {
	var comments []models.Comment
	if err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Children", func(db *gorm.DB) *gorm.DB {
			return db.Preload("User").Order("created_at ASC")
		}).
		Where("review_id = ? AND parent_id IS NULL", reviewID).
		Order("created_at DESC").
		Find(&comments).Error; err != nil {
		return nil, fmt.Errorf("%s: %w", appErrors.ErrCtxCommentFindByReview, err)
	}
	return comments, nil
}

func (r *commentRepository) FindByID(ctx context.Context, id uint) (*models.Comment, error) {
	var comment models.Comment
	if err := r.db.WithContext(ctx).First(&comment, id).Error; err != nil {
		return nil, fmt.Errorf("%s: %w", appErrors.ErrCtxCommentFindByID, err)
	}
	return &comment, nil
}

func (r *commentRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).
		Where("id = ? OR parent_id = ?", id, id).
		Delete(&models.Comment{}).Error; err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxCommentDelete, err)
	}
	return nil
}

type reviewLikeRepository struct {
	db *gorm.DB
}

func NewReviewLikeRepository(db *gorm.DB) ReviewLikeRepo {
	return &reviewLikeRepository{db: db}
}

func (r *reviewLikeRepository) Exists(ctx context.Context, userID, reviewID uint) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.ReviewLike{}).
		Where("user_id = ? AND review_id = ?", userID, reviewID).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("%s: %w", appErrors.ErrCtxLikeCheck, err)
	}
	return count > 0, nil
}

func (r *reviewLikeRepository) FindByUserAndReviewIDs(ctx context.Context, userID uint, reviewIDs []uint) ([]uint, error) {
	if len(reviewIDs) == 0 || userID == 0 {
		return nil, nil
	}
	var likedIDs []uint
	if err := r.db.WithContext(ctx).Model(&models.ReviewLike{}).
		Where("user_id = ? AND review_id IN ?", userID, reviewIDs).
		Pluck("review_id", &likedIDs).Error; err != nil {
		return nil, fmt.Errorf("%s: %w", appErrors.ErrCtxLikeCheck, err)
	}
	return likedIDs, nil
}

func (r *reviewLikeRepository) Create(ctx context.Context, like *models.ReviewLike) error {
	if err := r.db.WithContext(ctx).Create(like).Error; err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxLikeToggle, err)
	}
	return nil
}

func (r *reviewLikeRepository) Delete(ctx context.Context, userID, reviewID uint) error {
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND review_id = ?", userID, reviewID).
		Delete(&models.ReviewLike{}).Error; err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxLikeToggle, err)
	}
	return nil
}
