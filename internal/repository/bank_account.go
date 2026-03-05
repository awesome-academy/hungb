package repository

import (
	"context"
	"errors"
	"fmt"

	appErrors "sun-booking-tours/internal/errors"
	"sun-booking-tours/internal/models"

	"gorm.io/gorm"
)

type BankAccountRepo interface {
	FindByID(ctx context.Context, id uint) (*models.BankAccount, error)
	FindByUserID(ctx context.Context, userID uint) ([]models.BankAccount, error)
	CountByUserID(ctx context.Context, userID uint) (int64, error)
	Create(ctx context.Context, account *models.BankAccount) error
	Update(ctx context.Context, account *models.BankAccount) error
	Delete(ctx context.Context, id, userID uint) error
	ClearDefaultByUserID(ctx context.Context, userID uint) error
	SetDefault(ctx context.Context, id, userID uint) error
}

type bankAccountRepository struct {
	db *gorm.DB
}

func NewBankAccountRepository(db *gorm.DB) BankAccountRepo {
	return &bankAccountRepository{db: db}
}

func (r *bankAccountRepository) FindByID(ctx context.Context, id uint) (*models.BankAccount, error) {
	var account models.BankAccount
	if err := r.db.WithContext(ctx).First(&account, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrBankAccountNotFound
		}
		return nil, fmt.Errorf("%s: %w", appErrors.ErrCtxBankAccountFindByID, err)
	}
	return &account, nil
}

func (r *bankAccountRepository) FindByUserID(ctx context.Context, userID uint) ([]models.BankAccount, error) {
	var accounts []models.BankAccount
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("is_default DESC, created_at DESC").
		Find(&accounts).Error; err != nil {
		return nil, fmt.Errorf("%s: %w", appErrors.ErrCtxBankAccountFindByUser, err)
	}
	return accounts, nil
}

func (r *bankAccountRepository) CountByUserID(ctx context.Context, userID uint) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.BankAccount{}).
		Where("user_id = ?", userID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("%s: %w", appErrors.ErrCtxBankAccountCount, err)
	}
	return count, nil
}

func (r *bankAccountRepository) Create(ctx context.Context, account *models.BankAccount) error {
	if err := r.db.WithContext(ctx).Create(account).Error; err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxBankAccountCreate, err)
	}
	return nil
}

func (r *bankAccountRepository) Update(ctx context.Context, account *models.BankAccount) error {
	if err := r.db.WithContext(ctx).Save(account).Error; err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxBankAccountUpdate, err)
	}
	return nil
}

func (r *bankAccountRepository) Delete(ctx context.Context, id, userID uint) error {
	result := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&models.BankAccount{})
	if result.Error != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxBankAccountDelete, result.Error)
	}
	if result.RowsAffected == 0 {
		return appErrors.ErrBankAccountNotFound
	}
	return nil
}

// ClearDefaultByUserID sets is_default=false for all accounts of a user.
func (r *bankAccountRepository) ClearDefaultByUserID(ctx context.Context, userID uint) error {
	if err := r.db.WithContext(ctx).
		Model(&models.BankAccount{}).
		Where("user_id = ?", userID).
		Update("is_default", false).Error; err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxBankAccountClearDefault, err)
	}
	return nil
}

func (r *bankAccountRepository) SetDefault(ctx context.Context, id, userID uint) error {
	result := r.db.WithContext(ctx).
		Model(&models.BankAccount{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_default", true)
	if result.Error != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxBankAccountSetDefault, result.Error)
	}
	if result.RowsAffected == 0 {
		return appErrors.ErrBankAccountNotFound
	}
	return nil
}
