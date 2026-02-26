package repository

import (
	"context"
	"fmt"

	"sun-booking-tours/internal/models"

	"gorm.io/gorm"
)

// UserRepo is the data-access contract for user records.
// Depending on an interface (rather than the concrete struct) keeps services
// and handlers decoupled from the GORM implementation and makes unit-testing
// straightforward via simple fakes or mocks.
type UserRepo interface {
	FindByID(ctx context.Context, id uint) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	Create(ctx context.Context, user *models.User) error
	Update(ctx context.Context, user *models.User) error
}

// userRepository is the GORM-backed implementation of UserRepo.
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository returns a UserRepo backed by the given *gorm.DB.
func NewUserRepository(db *gorm.DB) UserRepo {
	return &userRepository{db: db}
}

// FindByID returns a user by primary key. Returns gorm.ErrRecordNotFound when missing.
func (r *userRepository) FindByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return &user, nil
}

// FindByEmail looks up a user by their exact (already-normalised) email.
// Emails are stored lower-cased on write, so a plain equality check hits the index.
// Returns gorm.ErrRecordNotFound when missing.
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).
		Where("email = ?", email).
		First(&user).Error; err != nil {
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	return &user, nil
}

// ExistsByEmail returns true when a user with that email already exists in the DB.
// Assumes the caller has already lower-cased the email (normalised on write).
func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.User{}).
		Where("email = ?", email).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("check email exists: %w", err)
	}
	return count > 0, nil
}

// Create inserts a new user; the model's ID field is populated by GORM on return.
func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

// Update saves all changed fields on the given user record.
func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}
