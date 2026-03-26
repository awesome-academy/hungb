package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"sun-booking-tours/internal/constants"
	appErrors "sun-booking-tours/internal/errors"
	"sun-booking-tours/internal/messages"
	"sun-booking-tours/internal/models"
	"sun-booking-tours/internal/repository"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

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
	db             *gorm.DB
	userRepo       repository.UserRepo
	socialAcctRepo repository.SocialAccountRepo
	emailService   *EmailService
	baseURL        string
}

func NewAuthService(db *gorm.DB, userRepo repository.UserRepo, socialAcctRepo repository.SocialAccountRepo, emailService *EmailService, baseURL string) *AuthService {
	return &AuthService{db: db, userRepo: userRepo, socialAcctRepo: socialAcctRepo, emailService: emailService, baseURL: baseURL}
}

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

func (s *AuthService) OAuthLogin(ctx context.Context, provider, providerID, email, name, avatarURL string) (*models.User, error) {
	email = strings.TrimSpace(strings.ToLower(email))

	if email == "" {
		return nil, appErrors.NewAppError(422, messages.ErrOAuthMissingEmail)
	}

	socialAcct, err := s.socialAcctRepo.FindByProvider(ctx, provider, providerID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			slog.ErrorContext(ctx, messages.LogOAuthFindSocial, "provider", provider, "error", err)
			return nil, appErrors.ErrInternalServerError
		}
	} else {
		user, err := s.userRepo.FindByID(ctx, socialAcct.UserID)
		if err != nil {
			slog.ErrorContext(ctx, messages.LogOAuthFindLinkedUser, "user_id", socialAcct.UserID, "error", err)
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

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		slog.ErrorContext(ctx, messages.LogOAuthFindByEmail, "email", email, "error", err)
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
		if err := s.socialAcctRepo.Create(ctx, &models.SocialAccount{
			UserID:     user.ID,
			Provider:   provider,
			ProviderID: providerID,
		}); err != nil {
			slog.ErrorContext(ctx, messages.LogOAuthLinkSocial, "error", err)
			return nil, appErrors.ErrInternalServerError
		}
		return user, nil
	}

	var newUser *models.User
	txErr := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txUserRepo := repository.NewUserRepository(tx)
		txSocialRepo := repository.NewSocialAccountRepository(tx)

		newUser = &models.User{
			FullName:  name,
			Email:     email,
			AvatarURL: avatarURL,
			Role:      constants.RoleUser,
			Status:    constants.StatusActive,
		}
		if err := txUserRepo.Create(ctx, newUser); err != nil {
			return err
		}

		return txSocialRepo.Create(ctx, &models.SocialAccount{
			UserID:     newUser.ID,
			Provider:   provider,
			ProviderID: providerID,
		})
	})
	if txErr != nil {
		if appErrors.IsDuplicateEntryError(txErr) {
			return nil, appErrors.ErrEmailAlreadyTaken
		}
		slog.ErrorContext(ctx, messages.LogOAuthCreateUser, "error", txErr)
		return nil, appErrors.ErrInternalServerError
	}

	return newUser, nil
}

func (s *AuthService) Register(ctx context.Context, form *RegisterForm) (*models.User, error) {
	form.Email = strings.TrimSpace(strings.ToLower(form.Email))
	form.FullName = strings.TrimSpace(form.FullName)

	if form.Password != form.PasswordConfirm {
		return nil, &appErrors.AppError{
			Status:  400,
			Message: messages.ErrPasswordMismatch,
		}
	}

	exists, err := s.userRepo.ExistsByEmail(ctx, form.Email)
	if err != nil {
		slog.ErrorContext(ctx, messages.LogRegisterCheckEmailExists, "error", err)
		return nil, appErrors.ErrInternalServerError
	}
	if exists {
		return nil, appErrors.ErrEmailAlreadyTaken
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(form.Password), bcrypt.DefaultCost)
	if err != nil {
		slog.ErrorContext(ctx, messages.LogRegisterHashPassword, "error", err)
		return nil, appErrors.ErrInternalServerError
	}

	status := constants.StatusActive
	emailVerified := true
	var verifyToken string
	var tokenExpiry *time.Time

	if s.emailService != nil && s.emailService.IsEnabled() {
		status = constants.StatusInactive
		emailVerified = false
		token, err := generateVerifyToken()
		if err != nil {
			slog.ErrorContext(ctx, messages.LogRegisterGenToken, "error", err)
			return nil, appErrors.ErrInternalServerError
		}
		verifyToken = token
		exp := time.Now().Add(24 * time.Hour)
		tokenExpiry = &exp
	}

	user := &models.User{
		FullName:          form.FullName,
		Email:             form.Email,
		Password:          string(hashed),
		Role:              constants.RoleUser,
		Status:            status,
		EmailVerified:     emailVerified,
		VerifyToken:       verifyToken,
		VerifyTokenExpiry: tokenExpiry,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		if appErrors.IsDuplicateEntryError(err) {
			return nil, appErrors.ErrEmailAlreadyTaken
		}
		slog.ErrorContext(ctx, messages.LogRegisterCreateUser, "error", err)
		return nil, appErrors.ErrInternalServerError
	}

	if verifyToken != "" {
		verifyPath, err := url.JoinPath(s.baseURL, constants.RouteVerifyEmail)
		if err != nil {
			slog.ErrorContext(ctx, "register: build verify URL", "error", err)
			return nil, appErrors.ErrInternalServerError
		}
		verifyURL := verifyPath + "?token=" + url.QueryEscape(verifyToken)
		if err := s.emailService.SendVerificationEmail(user.Email, user.FullName, verifyURL); err != nil {
			slog.ErrorContext(ctx, messages.LogRegisterSendEmail, "email", user.Email, "error", err)
			return nil, fmt.Errorf("send verification email: %w", err)
		}
	}

	return user, nil
}

func (s *AuthService) VerifyEmail(ctx context.Context, token string) (*models.User, error) {
	if token == "" {
		return nil, appErrors.NewAppError(400, messages.ErrVerifyTokenInvalid)
	}

	user, err := s.userRepo.FindByVerifyToken(ctx, token)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NewAppError(400, messages.ErrVerifyTokenInvalid)
		}
		slog.ErrorContext(ctx, messages.LogVerifyFindByToken, "error", err)
		return nil, appErrors.ErrInternalServerError
	}

	if user.VerifyTokenExpiry != nil && time.Now().After(*user.VerifyTokenExpiry) {
		return nil, appErrors.NewAppError(400, messages.ErrVerifyTokenExpired)
	}

	if user.EmailVerified {
		return user, nil
	}

	user.EmailVerified = true
	user.Status = constants.StatusActive
	user.VerifyToken = ""
	user.VerifyTokenExpiry = nil

	if err := s.userRepo.Update(ctx, user); err != nil {
		slog.ErrorContext(ctx, messages.LogVerifyUpdateUser, "error", err)
		return nil, appErrors.ErrInternalServerError
	}

	return user, nil
}

func (s *AuthService) EmailVerificationRequired() bool {
	return s.emailService != nil && s.emailService.IsEnabled()
}

func generateVerifyToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
