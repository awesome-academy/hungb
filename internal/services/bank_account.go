package services

import (
	"context"
	"log/slog"
	"strings"

	appErrors "sun-booking-tours/internal/errors"
	"sun-booking-tours/internal/messages"
	"sun-booking-tours/internal/models"
	"sun-booking-tours/internal/repository"

	"gorm.io/gorm"
)

type BankAccountForm struct {
	BankName      string `form:"bank_name"       binding:"required,max=255"`
	AccountNumber string `form:"account_number"  binding:"required,max=50"`
	AccountHolder string `form:"account_holder"  binding:"required,max=255"`
}

type BankAccountService struct {
	db   *gorm.DB
	repo repository.BankAccountRepo
}

func NewBankAccountService(db *gorm.DB, repo repository.BankAccountRepo) *BankAccountService {
	return &BankAccountService{db: db, repo: repo}
}

func (s *BankAccountService) ListByUser(ctx context.Context, userID uint) ([]models.BankAccount, error) {
	accounts, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		slog.ErrorContext(ctx, messages.LogBankAccountLoadFailed, "user_id", userID, "error", err)
		return nil, appErrors.ErrInternalServerError
	}
	return accounts, nil
}

func (s *BankAccountService) GetByID(ctx context.Context, id, userID uint) (*models.BankAccount, error) {
	account, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if appErrors.Is(err, appErrors.ErrBankAccountNotFound) {
			return nil, err
		}
		slog.ErrorContext(ctx, messages.LogBankAccountLoadFailed, "id", id, "error", err)
		return nil, appErrors.ErrInternalServerError
	}
	if account.UserID != userID {
		return nil, appErrors.ErrForbidden
	}
	return account, nil
}

func (s *BankAccountService) Create(ctx context.Context, userID uint, form *BankAccountForm) error {
	account := &models.BankAccount{
		UserID:        userID,
		BankName:      strings.TrimSpace(form.BankName),
		AccountNumber: strings.TrimSpace(form.AccountNumber),
		AccountHolder: strings.TrimSpace(form.AccountHolder),
	}

	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := repository.NewBankAccountRepository(tx)
		count, err := txRepo.CountByUserID(ctx, userID)
		if err != nil {
			return err
		}
		account.IsDefault = count == 0
		return txRepo.Create(ctx, account)
	}); err != nil {
		slog.ErrorContext(ctx, messages.LogBankAccountCreateFailed, "user_id", userID, "error", err)
		return appErrors.ErrInternalServerError
	}
	return nil
}

func (s *BankAccountService) Update(ctx context.Context, id, userID uint, form *BankAccountForm) error {
	account, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if appErrors.Is(err, appErrors.ErrBankAccountNotFound) {
			return err
		}
		slog.ErrorContext(ctx, messages.LogBankAccountLoadFailed, "id", id, "error", err)
		return appErrors.ErrInternalServerError
	}
	if account.UserID != userID {
		return appErrors.ErrForbidden
	}

	account.BankName = strings.TrimSpace(form.BankName)
	account.AccountNumber = strings.TrimSpace(form.AccountNumber)
	account.AccountHolder = strings.TrimSpace(form.AccountHolder)

	if err := s.repo.Update(ctx, account); err != nil {
		slog.ErrorContext(ctx, messages.LogBankAccountUpdateFailed, "id", id, "error", err)
		return appErrors.ErrInternalServerError
	}
	return nil
}

func (s *BankAccountService) Delete(ctx context.Context, id, userID uint) error {
	account, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if appErrors.Is(err, appErrors.ErrBankAccountNotFound) {
			return err
		}
		slog.ErrorContext(ctx, messages.LogBankAccountLoadFailed, "id", id, "error", err)
		return appErrors.ErrInternalServerError
	}
	if account.UserID != userID {
		return appErrors.ErrForbidden
	}

	wasDefault := account.IsDefault

	if err := s.repo.Delete(ctx, id, userID); err != nil {
		slog.ErrorContext(ctx, messages.LogBankAccountDeleteFailed, "id", id, "error", err)
		return appErrors.ErrInternalServerError
	}

	// If deleted account was default, set the first remaining account as default
	if wasDefault {
		accounts, err := s.repo.FindByUserID(ctx, userID)
		if err != nil {
			slog.ErrorContext(ctx, messages.LogBankAccountSetDefaultFailed, "user_id", userID, "error", err)
			return nil
		}
		if len(accounts) == 0 {
			return nil
		}
		if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			txRepo := repository.NewBankAccountRepository(tx)
			if err := txRepo.ClearDefaultByUserID(ctx, userID); err != nil {
				return err
			}
			return txRepo.SetDefault(ctx, accounts[0].ID, userID)
		}); err != nil {
			slog.ErrorContext(ctx, messages.LogBankAccountSetDefaultFailed, "id", accounts[0].ID, "error", err)
		}
	}

	return nil
}

func (s *BankAccountService) SetDefault(ctx context.Context, id, userID uint) error {
	account, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if appErrors.Is(err, appErrors.ErrBankAccountNotFound) {
			return err
		}
		slog.ErrorContext(ctx, messages.LogBankAccountLoadFailed, "id", id, "error", err)
		return appErrors.ErrInternalServerError
	}
	if account.UserID != userID {
		return appErrors.ErrForbidden
	}

	// Atomic: clear all defaults + set new default in one transaction
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := repository.NewBankAccountRepository(tx)
		if err := txRepo.ClearDefaultByUserID(ctx, userID); err != nil {
			return err
		}
		return txRepo.SetDefault(ctx, id, userID)
	}); err != nil {
		slog.ErrorContext(ctx, messages.LogBankAccountSetDefaultFailed, "id", id, "error", err)
		return appErrors.ErrInternalServerError
	}

	return nil
}
