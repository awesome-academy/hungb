package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"sun-booking-tours/internal/constants"
	appErrors "sun-booking-tours/internal/errors"
	"sun-booking-tours/internal/messages"
	"sun-booking-tours/internal/models"
	"sun-booking-tours/internal/repository"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Sentinel errors for Login — distinguishable by appErrors.Is.
var (
	ErrAdminMustUsePortal = appErrors.NewAppError(403, messages.AuthErrAdminMustUsePortal)
	ErrAccountBanned      = appErrors.NewAppError(403, messages.AuthErrAccountBanned)
	ErrAccountInactive    = appErrors.NewAppError(403, messages.AuthErrAccountInactive)
)

type LoginForm struct {
	Email    string `form:"email"    binding:"required,email"`
	Password string `form:"password" binding:"required"`
}

type RegisterForm struct {
	FullName        string `form:"full_name"        binding:"required,min=2"`
	Email           string `form:"email"            binding:"required,email"`
	Password        string `form:"password"         binding:"required,min=8"`
	PasswordConfirm string `form:"password_confirm" binding:"required"`
}

type AuthService struct {
	userRepo       repository.UserRepo
	socialAcctRepo repository.SocialAccountRepo
}

func NewAuthService(userRepo repository.UserRepo, socialAcctRepo repository.SocialAccountRepo) *AuthService {
	return &AuthService{userRepo: userRepo, socialAcctRepo: socialAcctRepo}
}

// Login verifies credentials and returns the authenticated user.
// Returns ErrInvalidCredentials for wrong email/password, ErrAccountBanned,
// ErrAccountInactive, or ErrAdminMustUsePortal for other rejection cases.
func (s *AuthService) Login(ctx context.Context, form *LoginForm) (*models.User, error) {
	form.Email = strings.TrimSpace(strings.ToLower(form.Email))

	user, err := s.userRepo.FindByEmail(ctx, form.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrInvalidCredentials
		}
		slog.ErrorContext(ctx, messages.LogLoginFindUser, "email", form.Email, "error", err)
		return nil, appErrors.ErrInternalServerError
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(form.Password)); err != nil {
		return nil, appErrors.ErrInvalidCredentials
	}

	switch user.Status {
	case constants.StatusBanned:
		return nil, ErrAccountBanned
	case constants.StatusInactive:
		return nil, ErrAccountInactive
	}

	if user.Role == constants.RoleAdmin {
		return nil, ErrAdminMustUsePortal
	}

	return user, nil
}

// AdminLogin verifies credentials and ensures the user has admin role.
func (s *AuthService) AdminLogin(ctx context.Context, form *LoginForm) (*models.User, error) {
	form.Email = strings.TrimSpace(strings.ToLower(form.Email))

	user, err := s.userRepo.FindByEmail(ctx, form.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrInvalidCredentials
		}
		slog.ErrorContext(ctx, messages.LogAdminLoginFindUser, "email", form.Email, "error", err)
		return nil, appErrors.ErrInternalServerError
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(form.Password)); err != nil {
		return nil, appErrors.ErrInvalidCredentials
	}

	if user.Role != constants.RoleAdmin {
		return nil, appErrors.ErrForbidden
	}

	switch user.Status {
	case constants.StatusBanned:
		return nil, ErrAccountBanned
	case constants.StatusInactive:
		return nil, ErrAccountInactive
	}

	return user, nil
}

// OAuthLogin finds or creates a user from OAuth provider data.
// If the social account exists, returns the linked user.
// If the email matches an existing user, links the social account.
// Otherwise creates a new user with an empty password.
func (s *AuthService) OAuthLogin(ctx context.Context, provider, providerID, email, name, avatarURL string) (*models.User, error) {
	email = strings.TrimSpace(strings.ToLower(email))

	// Check if social account already exists
	socialAcct, err := s.socialAcctRepo.FindByProvider(ctx, provider, providerID)
	if err == nil && socialAcct != nil {
		user, err := s.userRepo.FindByID(ctx, socialAcct.UserID)
		if err != nil {
			return nil, appErrors.ErrInternalServerError
		}
		if user.Role == constants.RoleAdmin {
			return nil, ErrAdminMustUsePortal
		}
		switch user.Status {
		case constants.StatusBanned:
			return nil, ErrAccountBanned
		case constants.StatusInactive:
			return nil, ErrAccountInactive
		}
		return user, nil
	}

	// Check if user with same email exists
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, appErrors.ErrInternalServerError
	}

	if user != nil {
		if user.Role == constants.RoleAdmin {
			return nil, ErrAdminMustUsePortal
		}
		switch user.Status {
		case constants.StatusBanned:
			return nil, ErrAccountBanned
		case constants.StatusInactive:
			return nil, ErrAccountInactive
		}
		// Link social account to existing user
		if err := s.socialAcctRepo.Create(ctx, &models.SocialAccount{
			UserID:     user.ID,
			Provider:   provider,
			ProviderID: providerID,
		}); err != nil {
			slog.ErrorContext(ctx, "oauth: link social account", "error", err)
		}
		return user, nil
	}

	// Create new user
	newUser := &models.User{
		FullName:  name,
		Email:     email,
		AvatarURL: avatarURL,
		Role:      constants.RoleUser,
		Status:    constants.StatusActive,
	}
	if err := s.userRepo.Create(ctx, newUser); err != nil {
		if appErrors.IsDuplicateEntryError(err) {
			return nil, appErrors.ErrEmailAlreadyTaken
		}
		return nil, appErrors.ErrInternalServerError
	}

	if err := s.socialAcctRepo.Create(ctx, &models.SocialAccount{
		UserID:     newUser.ID,
		Provider:   provider,
		ProviderID: providerID,
	}); err != nil {
		slog.ErrorContext(ctx, "oauth: create social account", "error", err)
	}

	return newUser, nil
}

// Register validates form data, hashes the password, and creates a new user.
// Returns the created user or an *AppError describing what went wrong.
func (s *AuthService) Register(ctx context.Context, form *RegisterForm) (*models.User, error) {
	// Normalise email and name so stored values are always consistent.
	form.Email = strings.TrimSpace(strings.ToLower(form.Email))
	form.FullName = strings.TrimSpace(form.FullName)

	// Password confirmation check.
	if form.Password != form.PasswordConfirm {
		return nil, &appErrors.AppError{
			Status:  400,
			Message: messages.ErrPasswordMismatch,
		}
	}

	// Best-effort uniqueness pre-check (optimistic path; not a hard guarantee).
	exists, err := s.userRepo.ExistsByEmail(ctx, form.Email)
	if err != nil {
		slog.ErrorContext(ctx, messages.LogRegisterCheckEmailExists, "error", err)
		return nil, appErrors.ErrInternalServerError
	}
	if exists {
		return nil, appErrors.ErrEmailAlreadyTaken
	}

	// Hash password.
	hashed, err := bcrypt.GenerateFromPassword([]byte(form.Password), bcrypt.DefaultCost)
	if err != nil {
		slog.ErrorContext(ctx, messages.LogRegisterHashPassword, "error", err)
		return nil, appErrors.ErrInternalServerError
	}

	user := &models.User{
		FullName: form.FullName,
		Email:    form.Email,
		Password: string(hashed),
		Role:     constants.RoleUser,
		Status:   constants.StatusActive,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		// Two concurrent registrations can both pass ExistsByEmail and then one
		// will hit the DB unique constraint. Map that to the correct sentinel so
		// the user sees a proper message and no internal details leak.
		if appErrors.IsDuplicateEntryError(err) {
			return nil, appErrors.ErrEmailAlreadyTaken
		}
		slog.ErrorContext(ctx, messages.LogRegisterCreateUser, "error", err)
		return nil, fmt.Errorf("%w", appErrors.ErrInternalServerError)
	}

	return user, nil
}
