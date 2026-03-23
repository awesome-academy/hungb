package services

import (
	"context"
	"errors"
	"fmt"

	"sun-booking-tours/internal/constants"
	appErrors "sun-booking-tours/internal/errors"
	"sun-booking-tours/internal/models"
	"sun-booking-tours/internal/repository"

	"gorm.io/gorm"
)

type AdminUserService struct {
	userRepo repository.UserRepo
}

func NewAdminUserService(userRepo repository.UserRepo) *AdminUserService {
	return &AdminUserService{userRepo: userRepo}
}

func (s *AdminUserService) ListUsers(ctx context.Context, filter repository.UserFilter) ([]models.User, int64, error) {
	users, total, err := s.userRepo.FindAll(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", appErrors.ErrCtxUserFindAll, err)
	}
	return users, total, nil
}

func (s *AdminUserService) GetUserDetail(ctx context.Context, id uint) (*models.User, error) {
	user, err := s.userRepo.FindByIDWithRelations(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrUserNotFound
		}
		return nil, fmt.Errorf("%s: %w", appErrors.ErrCtxUserFindByIDWithRelations, err)
	}
	return user, nil
}

// adminUser cannot change their own status, and no admin can change another admin's status.
func (s *AdminUserService) UpdateUserStatus(ctx context.Context, adminUser *models.User, targetID uint, status string) error {
	if adminUser.ID == targetID {
		return appErrors.ErrCannotBanSelf
	}

	target, err := s.userRepo.FindByID(ctx, targetID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return appErrors.ErrUserNotFound
		}
		return fmt.Errorf("%s: %w", appErrors.ErrCtxUserFindByID, err)
	}

	if target.Role == constants.RoleAdmin {
		return appErrors.ErrCannotBanAdmin
	}

	if err := s.userRepo.UpdateStatus(ctx, targetID, status); err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxUserUpdateStatus, err)
	}
	return nil
}
