package repository

import (
	"context"
	"fmt"

	"sun-booking-tours/internal/messages"
	"sun-booking-tours/internal/models"

	"gorm.io/gorm"
)

// UserFilter carries optional filtering and pagination parameters for user list queries.
type UserFilter struct {
	Role      string
	Status    string
	Keyword   string
	SortBy    string
	SortOrder string
	Page      int
	Limit     int
}

// UserRepo is the data-access contract for user records.
type UserRepo interface {
	FindByID(ctx context.Context, id uint) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	Create(ctx context.Context, user *models.User) error
	Update(ctx context.Context, user *models.User) error
	FindAll(ctx context.Context, filter UserFilter) ([]models.User, int64, error)
	FindByIDWithRelations(ctx context.Context, id uint) (*models.User, error)
	UpdateStatus(ctx context.Context, id uint, status string) error
	FindByVerifyToken(ctx context.Context, token string) (*models.User, error)
}

// userRepository is the GORM-backed implementation of UserRepo.
type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepo {
	return &userRepository{db: db}
}

func (r *userRepository) FindByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, fmt.Errorf("%s: %w", messages.ErrCtxFindUserByID, err)
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).
		Where("email = ?", email).
		First(&user).Error; err != nil {
		return nil, fmt.Errorf("%s: %w", messages.ErrCtxFindUserByEmail, err)
	}
	return &user, nil
}

func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.User{}).
		Where("email = ?", email).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("%s: %w", messages.ErrCtxCheckEmailExists, err)
	}
	return count > 0, nil
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return fmt.Errorf("%s: %w", messages.ErrCtxCreateUser, err)
	}
	return nil
}

func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		return fmt.Errorf("%s: %w", messages.ErrCtxUpdateUser, err)
	}
	return nil
}

func (r *userRepository) FindAll(ctx context.Context, filter UserFilter) ([]models.User, int64, error) {
	q := r.db.WithContext(ctx).Model(&models.User{})

	if filter.Role != "" {
		q = q.Where("role = ?", filter.Role)
	}
	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}
	if filter.Keyword != "" {
		like := "%" + filter.Keyword + "%"
		q = q.Where("full_name ILIKE ? OR email ILIKE ?", like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("%s: %w", messages.ErrCtxCountAllUsers, err)
	}

	allowedSortBy := map[string]bool{
		"created_at": true,
		"email":      true,
		"full_name":  true,
	}
	sortBy := "created_at"
	if allowedSortBy[filter.SortBy] {
		sortBy = filter.SortBy
	}
	sortOrder := "desc"
	if filter.SortOrder == "asc" {
		sortOrder = "asc"
	}
	q = q.Order(sortBy + " " + sortOrder)

	limit := filter.Limit
	if limit <= 0 {
		limit = 10
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	q = q.Limit(limit).Offset((page - 1) * limit)

	var users []models.User
	if err := q.Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("%s: %w", messages.ErrCtxFindAllUsers, err)
	}
	return users, total, nil
}

func (r *userRepository) FindByIDWithRelations(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).
		Preload("Bookings").
		Preload("Bookings.Tour").
		Preload("Reviews").
		Preload("BankAccounts").
		Preload("Ratings").
		Preload("Ratings.Tour").
		First(&user, id).Error; err != nil {
		return nil, fmt.Errorf("%s: %w", messages.ErrCtxFindUserByIDWithRelations, err)
	}
	return &user, nil
}

func (r *userRepository) UpdateStatus(ctx context.Context, id uint, status string) error {
	if err := r.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", id).
		Update("status", status).Error; err != nil {
		return fmt.Errorf("%s: %w", messages.ErrCtxUpdateUserStatus, err)
	}
	return nil
}

func (r *userRepository) FindByVerifyToken(ctx context.Context, token string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).
		Where("verify_token = ?", token).
		First(&user).Error; err != nil {
		return nil, fmt.Errorf("%s: %w", messages.ErrCtxFindUserByVerifyToken, err)
	}
	return &user, nil
}
