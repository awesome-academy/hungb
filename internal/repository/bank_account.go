package repository

import (
	"context"
	"fmt"

	"sun-booking-tours/internal/models"

	"gorm.io/gorm"
)

type BankAccountRepo interface {
	FindByID(ctx context.Context, id uint) (*models.BankAccount, error)
	FindByUserID(ctx context.Context, userID uint) ([]models.BankAccount, error)
	CountByUserID(ctx context.Context, userID uint) (int64, error)
	Create(ctx context.Context, account *models.BankAccount) error
	Update(ctx context.Context, account *models.BankAccount) error
	Delete(ctx context.Context, id uint) error
	ClearDefaultByUserID(ctx context.Context, userID uint) error
	SetDefault(ctx context.Context, id uint) error
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
		return nil, fmt.Errorf("find bank account by id: %w", err)
	}
	return &account, nil
}

func (r *bankAccountRepository) FindByUserID(ctx context.Context, userID uint) ([]models.BankAccount, error) {
	var accounts []models.BankAccount
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("is_default DESC, created_at DESC").
		Find(&accounts).Error; err != nil {
		return nil, fmt.Errorf("find bank accounts by user: %w", err)
	}
	return accounts, nil
}

func (r *bankAccountRepository) CountByUserID(ctx context.Context, userID uint) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.BankAccount{}).
		Where("user_id = ?", userID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count bank accounts: %w", err)
	}
	return count, nil
}

func (r *bankAccountRepository) Create(ctx context.Context, account *models.BankAccount) error {
	if err := r.db.WithContext(ctx).Create(account).Error; err != nil {
		return fmt.Errorf("create bank account: %w", err)
	}
	return nil
}

func (r *bankAccountRepository) Update(ctx context.Context, account *models.BankAccount) error {
	if err := r.db.WithContext(ctx).Save(account).Error; err != nil {
		return fmt.Errorf("update bank account: %w", err)
	}
	return nil
}

func (r *bankAccountRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&models.BankAccount{}, id).Error; err != nil {
		return fmt.Errorf("delete bank account: %w", err)
	}
	return nil
}

// ClearDefaultByUserID sets is_default=false for all accounts of a user.
func (r *bankAccountRepository) ClearDefaultByUserID(ctx context.Context, userID uint) error {
	if err := r.db.WithContext(ctx).
		Model(&models.BankAccount{}).
		Where("user_id = ?", userID).
		Update("is_default", false).Error; err != nil {
		return fmt.Errorf("clear default bank accounts: %w", err)
	}
	return nil
}

// SetDefault sets is_default=true for a specific account.
func (r *bankAccountRepository) SetDefault(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).
		Model(&models.BankAccount{}).
		Where("id = ?", id).
		Update("is_default", true).Error; err != nil {
		return fmt.Errorf("set default bank account: %w", err)
	}
	return nil
}
