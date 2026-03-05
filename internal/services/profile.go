package services

import (
	"context"
	"log/slog"
	"strings"

	appErrors "sun-booking-tours/internal/errors"
	"sun-booking-tours/internal/messages"
	"sun-booking-tours/internal/models"
	"sun-booking-tours/internal/repository"
)

type ProfileUpdateForm struct {
	FullName  string `form:"full_name" binding:"required,min=2,max=255"`
	Phone     string `form:"phone"     binding:"omitempty,max=20"`
	AvatarURL string `form:"avatar_url" binding:"omitempty,max=500"`
}

type ProfileService struct {
	userRepo repository.UserRepo
}

func NewProfileService(userRepo repository.UserRepo) *ProfileService {
	return &ProfileService{userRepo: userRepo}
}

func (s *ProfileService) GetProfile(ctx context.Context, userID uint) (*models.User, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		slog.ErrorContext(ctx, messages.LogProfileLoadFailed, "user_id", userID, "error", err)
		return nil, appErrors.ErrInternalServerError
	}
	return user, nil
}

func (s *ProfileService) UpdateProfile(ctx context.Context, userID uint, form *ProfileUpdateForm) (*models.User, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		slog.ErrorContext(ctx, messages.LogProfileLoadFailed, "user_id", userID, "error", err)
		return nil, appErrors.ErrInternalServerError
	}

	user.FullName = strings.TrimSpace(form.FullName)
	user.Phone = strings.TrimSpace(form.Phone)
	user.AvatarURL = strings.TrimSpace(form.AvatarURL)

	if err := s.userRepo.Update(ctx, user); err != nil {
		slog.ErrorContext(ctx, messages.LogProfileUpdateFailed, "user_id", userID, "error", err)
		return nil, appErrors.ErrInternalServerError
	}

	return user, nil
}
