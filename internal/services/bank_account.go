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

type BankAccountForm struct {
	BankName      string `form:"bank_name"       binding:"required,max=255"`
	AccountNumber string `form:"account_number"  binding:"required,max=50"`
	AccountHolder string `form:"account_holder"  binding:"required,max=255"`
}

type BankAccountService struct {
	repo repository.BankAccountRepo
}

func NewBankAccountService(repo repository.BankAccountRepo) *BankAccountService {
	return &BankAccountService{repo: repo}
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
		return nil, appErrors.ErrFailedToFetch
	}
	if account.UserID != userID {
		return nil, appErrors.ErrForbidden
	}
	return account, nil
}

func (s *BankAccountService) Create(ctx context.Context, userID uint, form *BankAccountForm) error {
	// If this is the first account, auto-set as default
	count, err := s.repo.CountByUserID(ctx, userID)
	if err != nil {
		slog.ErrorContext(ctx, messages.LogBankAccountCreateFailed, "user_id", userID, "error", err)
		return appErrors.ErrInternalServerError
	}

	account := &models.BankAccount{
		UserID:        userID,
		BankName:      strings.TrimSpace(form.BankName),
		AccountNumber: strings.TrimSpace(form.AccountNumber),
		AccountHolder: strings.TrimSpace(form.AccountHolder),
		IsDefault:     count == 0,
	}

	if err := s.repo.Create(ctx, account); err != nil {
		slog.ErrorContext(ctx, messages.LogBankAccountCreateFailed, "user_id", userID, "error", err)
		return appErrors.ErrInternalServerError
	}
	return nil
}

func (s *BankAccountService) Update(ctx context.Context, id, userID uint, form *BankAccountForm) error {
	account, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return appErrors.ErrFailedToFetch
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
		return appErrors.ErrFailedToFetch
	}
	if account.UserID != userID {
		return appErrors.ErrForbidden
	}

	wasDefault := account.IsDefault

	if err := s.repo.Delete(ctx, id); err != nil {
		slog.ErrorContext(ctx, messages.LogBankAccountDeleteFailed, "id", id, "error", err)
		return appErrors.ErrInternalServerError
	}

	// If deleted account was default, set the first remaining account as default
	if wasDefault {
		accounts, err := s.repo.FindByUserID(ctx, userID)
		if err == nil && len(accounts) > 0 {
			_ = s.repo.ClearDefaultByUserID(ctx, userID)
			_ = s.repo.SetDefault(ctx, accounts[0].ID)
		}
	}

	return nil
}

func (s *BankAccountService) SetDefault(ctx context.Context, id, userID uint) error {
	account, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return appErrors.ErrFailedToFetch
	}
	if account.UserID != userID {
		return appErrors.ErrForbidden
	}

	if err := s.repo.ClearDefaultByUserID(ctx, userID); err != nil {
		slog.ErrorContext(ctx, messages.LogBankAccountSetDefaultFailed, "user_id", userID, "error", err)
		return appErrors.ErrInternalServerError
	}
	if err := s.repo.SetDefault(ctx, id); err != nil {
		slog.ErrorContext(ctx, messages.LogBankAccountSetDefaultFailed, "id", id, "error", err)
		return appErrors.ErrInternalServerError
	}

	return nil
}
