package services

import (
	"context"
	"fmt"
	"strings"

	appErrors "sun-booking-tours/internal/errors"
	"sun-booking-tours/internal/models"
	"sun-booking-tours/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

// RegisterForm holds data submitted from the public registration form.
type RegisterForm struct {
	FullName        string `form:"full_name"        binding:"required,min=2"`
	Email           string `form:"email"            binding:"required,email"`
	Password        string `form:"password"         binding:"required,min=8"`
	PasswordConfirm string `form:"password_confirm" binding:"required"`
}

type AuthService struct {
	userRepo *repository.UserRepository
}

func NewAuthService(userRepo *repository.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

// Register validates form data, hashes the password, and creates a new user.
// Returns the created user or an *AppError describing what went wrong.
func (s *AuthService) Register(ctx context.Context, form *RegisterForm) (*models.User, error) {
	// Normalise email
	form.Email = strings.TrimSpace(strings.ToLower(form.Email))
	form.FullName = strings.TrimSpace(form.FullName)

	// Password confirmation check
	if form.Password != form.PasswordConfirm {
		return nil, fmt.Errorf("%w", &appErrors.AppError{
			Status:  400,
			Message: "Mật khẩu xác nhận không khớp",
		})
	}

	// Unique email check
	exists, err := s.userRepo.ExistsByEmail(ctx, form.Email)
	if err != nil {
		return nil, fmt.Errorf("check email: %w", err)
	}
	if exists {
		return nil, appErrors.ErrEmailAlreadyTaken
	}

	// Hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(form.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &models.User{
		FullName: form.FullName,
		Email:    form.Email,
		Password: string(hashed),
		Role:     "user",
		Status:   "active",
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}
