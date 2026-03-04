package repository

import (
	"context"
	"fmt"

	"sun-booking-tours/internal/models"

	"gorm.io/gorm"
)

type SocialAccountRepo interface {
	FindByProvider(ctx context.Context, provider, providerID string) (*models.SocialAccount, error)
	Create(ctx context.Context, account *models.SocialAccount) error
}

type socialAccountRepository struct {
	db *gorm.DB
}

func NewSocialAccountRepository(db *gorm.DB) SocialAccountRepo {
	return &socialAccountRepository{db: db}
}

func (r *socialAccountRepository) FindByProvider(ctx context.Context, provider, providerID string) (*models.SocialAccount, error) {
	var acct models.SocialAccount
	if err := r.db.WithContext(ctx).
		Where("provider = ? AND provider_id = ?", provider, providerID).
		First(&acct).Error; err != nil {
		return nil, fmt.Errorf("find social account: %w", err)
	}
	return &acct, nil
}

func (r *socialAccountRepository) Create(ctx context.Context, account *models.SocialAccount) error {
	if err := r.db.WithContext(ctx).Create(account).Error; err != nil {
		return fmt.Errorf("create social account: %w", err)
	}
	return nil
}
