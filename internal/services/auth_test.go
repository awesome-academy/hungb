package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"sun-booking-tours/internal/constants"
	appErrors "sun-booking-tours/internal/errors"
	"sun-booking-tours/internal/models"
	"sun-booking-tours/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// --- helpers ----------------------------------------------------------

func setupAuthService(userRepo *testutil.MockUserRepo, socialRepo *testutil.MockSocialAccountRepo) *AuthService {
	return NewAuthService(nil, userRepo, socialRepo, nil, testutil.TestBaseURL)
}

// --- Login tests ------------------------------------------------------

func TestLogin_Success(t *testing.T) {
	userRepo := new(testutil.MockUserRepo)
	svc := setupAuthService(userRepo, nil)

	user := testutil.NewActiveUser()
	userRepo.On("FindByEmail", mock.Anything, testutil.TestEmail).Return(user, nil)

	got, err := svc.Login(context.Background(), &LoginForm{
		Email:    testutil.TestEmail,
		Password: testutil.TestPassword,
	})

	require.NoError(t, err)
	assert.Equal(t, user.ID, got.ID)
	userRepo.AssertExpectations(t)
}

// --- Login tests — error cases (table-driven) -------------------------

func TestLogin_ErrorCases(t *testing.T) {
	tests := []struct {
		name      string
		setupRepo func(repo *testutil.MockUserRepo)
		form      LoginForm
		wantErr   *appErrors.AppError
	}{
		{
			name: "email not found",
			setupRepo: func(repo *testutil.MockUserRepo) {
				repo.On("FindByEmail", mock.Anything, "unknown@example.com").
					Return(nil, fmt.Errorf("find user by email: %w", gorm.ErrRecordNotFound))
			},
			form:    LoginForm{Email: "unknown@example.com", Password: "any"},
			wantErr: appErrors.ErrInvalidCredentials,
		},
		{
			name: "wrong password",
			setupRepo: func(repo *testutil.MockUserRepo) {
				repo.On("FindByEmail", mock.Anything, testutil.TestEmail).Return(testutil.NewActiveUser(), nil)
			},
			form:    LoginForm{Email: testutil.TestEmail, Password: "wrong-password"},
			wantErr: appErrors.ErrInvalidCredentials,
		},
		{
			name: "banned user",
			setupRepo: func(repo *testutil.MockUserRepo) {
				repo.On("FindByEmail", mock.Anything, testutil.TestEmail).Return(testutil.NewBannedUser(), nil)
			},
			form:    LoginForm{Email: testutil.TestEmail, Password: testutil.TestPassword},
			wantErr: ErrAccountBanned,
		},
		{
			name: "inactive user",
			setupRepo: func(repo *testutil.MockUserRepo) {
				repo.On("FindByEmail", mock.Anything, testutil.TestEmail).Return(testutil.NewInactiveUser(), nil)
			},
			form:    LoginForm{Email: testutil.TestEmail, Password: testutil.TestPassword},
			wantErr: ErrAccountInactive,
		},
		{
			name: "admin must use portal",
			setupRepo: func(repo *testutil.MockUserRepo) {
				repo.On("FindByEmail", mock.Anything, testutil.TestAdminEmail).Return(testutil.NewAdminUser(), nil)
			},
			form:    LoginForm{Email: testutil.TestAdminEmail, Password: testutil.TestPassword},
			wantErr: ErrAdminMustUsePortal,
		},
		{
			name: "db error returns internal server error",
			setupRepo: func(repo *testutil.MockUserRepo) {
				repo.On("FindByEmail", mock.Anything, testutil.TestEmail).
					Return(nil, fmt.Errorf("connection refused"))
			},
			form:    LoginForm{Email: testutil.TestEmail, Password: testutil.TestPassword},
			wantErr: appErrors.ErrInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			userRepo := new(testutil.MockUserRepo)
			svc := setupAuthService(userRepo, nil)
			tc.setupRepo(userRepo)

			_, err := svc.Login(context.Background(), &tc.form)

			assert.True(t, appErrors.Is(err, tc.wantErr))
			userRepo.AssertExpectations(t)
		})
	}
}

func TestLogin_TrimsAndLowercasesEmail(t *testing.T) {
	userRepo := new(testutil.MockUserRepo)
	svc := setupAuthService(userRepo, nil)

	user := testutil.NewActiveUser()
	userRepo.On("FindByEmail", mock.Anything, testutil.TestEmail).Return(user, nil)

	got, err := svc.Login(context.Background(), &LoginForm{
		Email:    "  User@Example.COM  ",
		Password: testutil.TestPassword,
	})

	require.NoError(t, err)
	assert.Equal(t, user.ID, got.ID)
}

// --- AdminLogin tests -------------------------------------------------

func TestAdminLogin_Success(t *testing.T) {
	userRepo := new(testutil.MockUserRepo)
	svc := setupAuthService(userRepo, nil)

	admin := testutil.NewAdminUser()
	userRepo.On("FindByEmail", mock.Anything, testutil.TestAdminEmail).Return(admin, nil)

	got, err := svc.AdminLogin(context.Background(), &LoginForm{
		Email:    testutil.TestAdminEmail,
		Password: testutil.TestPassword,
	})

	require.NoError(t, err)
	assert.Equal(t, admin.ID, got.ID)
	assert.Equal(t, constants.RoleAdmin, got.Role)
}

// --- AdminLogin tests — error cases (table-driven) --------------------

func TestAdminLogin_ErrorCases(t *testing.T) {
	tests := []struct {
		name      string
		setupRepo func(repo *testutil.MockUserRepo)
		form      LoginForm
		wantErr   *appErrors.AppError
	}{
		{
			name: "email not found",
			setupRepo: func(repo *testutil.MockUserRepo) {
				repo.On("FindByEmail", mock.Anything, "nobody@example.com").
					Return(nil, fmt.Errorf("find user: %w", gorm.ErrRecordNotFound))
			},
			form:    LoginForm{Email: "nobody@example.com", Password: "any"},
			wantErr: appErrors.ErrInvalidCredentials,
		},
		{
			name: "wrong password",
			setupRepo: func(repo *testutil.MockUserRepo) {
				repo.On("FindByEmail", mock.Anything, testutil.TestAdminEmail).Return(testutil.NewAdminUser(), nil)
			},
			form:    LoginForm{Email: testutil.TestAdminEmail, Password: "wrong"},
			wantErr: appErrors.ErrInvalidCredentials,
		},
		{
			name: "not admin role",
			setupRepo: func(repo *testutil.MockUserRepo) {
				repo.On("FindByEmail", mock.Anything, testutil.TestEmail).Return(testutil.NewActiveUser(), nil)
			},
			form:    LoginForm{Email: testutil.TestEmail, Password: testutil.TestPassword},
			wantErr: appErrors.ErrForbidden,
		},
		{
			name: "banned admin",
			setupRepo: func(repo *testutil.MockUserRepo) {
				admin := testutil.NewAdminUser()
				admin.Status = constants.StatusBanned
				repo.On("FindByEmail", mock.Anything, testutil.TestAdminEmail).Return(admin, nil)
			},
			form:    LoginForm{Email: testutil.TestAdminEmail, Password: testutil.TestPassword},
			wantErr: ErrAccountBanned,
		},
		{
			name: "inactive admin",
			setupRepo: func(repo *testutil.MockUserRepo) {
				admin := testutil.NewAdminUser()
				admin.Status = constants.StatusInactive
				repo.On("FindByEmail", mock.Anything, testutil.TestAdminEmail).Return(admin, nil)
			},
			form:    LoginForm{Email: testutil.TestAdminEmail, Password: testutil.TestPassword},
			wantErr: ErrAccountInactive,
		},
		{
			name: "db error",
			setupRepo: func(repo *testutil.MockUserRepo) {
				repo.On("FindByEmail", mock.Anything, testutil.TestAdminEmail).
					Return(nil, fmt.Errorf("db is down"))
			},
			form:    LoginForm{Email: testutil.TestAdminEmail, Password: testutil.TestPassword},
			wantErr: appErrors.ErrInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			userRepo := new(testutil.MockUserRepo)
			svc := setupAuthService(userRepo, nil)
			tc.setupRepo(userRepo)

			_, err := svc.AdminLogin(context.Background(), &tc.form)

			assert.True(t, appErrors.Is(err, tc.wantErr))
			userRepo.AssertExpectations(t)
		})
	}
}

// --- Register tests ---------------------------------------------------

func TestRegister_Success_NoEmailVerification(t *testing.T) {
	userRepo := new(testutil.MockUserRepo)
	svc := setupAuthService(userRepo, nil) // nil emailService → no verification

	userRepo.On("ExistsByEmail", mock.Anything, "new@example.com").Return(false, nil)
	userRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).
		Run(func(args mock.Arguments) {
			u := args.Get(1).(*models.User)
			u.ID = 10
		}).
		Return(nil)

	got, err := svc.Register(context.Background(), &RegisterForm{
		FullName:        "New User",
		Email:           "new@example.com",
		Password:        "password123",
		PasswordConfirm: "password123",
	})

	require.NoError(t, err)
	assert.Equal(t, uint(10), got.ID)
	assert.Equal(t, constants.RoleUser, got.Role)
	assert.Equal(t, constants.StatusActive, got.Status)
	assert.True(t, got.EmailVerified)
	userRepo.AssertExpectations(t)
}

func TestRegister_PasswordMismatch(t *testing.T) {
	userRepo := new(testutil.MockUserRepo)
	svc := setupAuthService(userRepo, nil)

	_, err := svc.Register(context.Background(), &RegisterForm{
		FullName:        "User",
		Email:           "test@example.com",
		Password:        "password123",
		PasswordConfirm: "different",
	})

	require.Error(t, err)
	var appErr *appErrors.AppError
	assert.ErrorAs(t, err, &appErr)
	assert.Equal(t, 400, appErr.Status)
}

func TestRegister_EmailAlreadyExists(t *testing.T) {
	userRepo := new(testutil.MockUserRepo)
	svc := setupAuthService(userRepo, nil)

	userRepo.On("ExistsByEmail", mock.Anything, testutil.TestEmail).Return(true, nil)

	_, err := svc.Register(context.Background(), &RegisterForm{
		FullName:        "User",
		Email:           testutil.TestEmail,
		Password:        "password123",
		PasswordConfirm: "password123",
	})

	assert.True(t, appErrors.Is(err, appErrors.ErrEmailAlreadyTaken))
}

func TestRegister_ExistsByEmail_DBError(t *testing.T) {
	userRepo := new(testutil.MockUserRepo)
	svc := setupAuthService(userRepo, nil)

	userRepo.On("ExistsByEmail", mock.Anything, "x@y.com").Return(false, fmt.Errorf("db err"))

	_, err := svc.Register(context.Background(), &RegisterForm{
		FullName:        "User",
		Email:           "x@y.com",
		Password:        "password123",
		PasswordConfirm: "password123",
	})

	assert.True(t, appErrors.Is(err, appErrors.ErrInternalServerError))
}

func TestRegister_CreateUser_DBError(t *testing.T) {
	userRepo := new(testutil.MockUserRepo)
	svc := setupAuthService(userRepo, nil)

	userRepo.On("ExistsByEmail", mock.Anything, "new@example.com").Return(false, nil)
	userRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).
		Return(fmt.Errorf("insert error"))

	_, err := svc.Register(context.Background(), &RegisterForm{
		FullName:        "User",
		Email:           "new@example.com",
		Password:        "password123",
		PasswordConfirm: "password123",
	})

	require.Error(t, err)
	assert.True(t, appErrors.Is(err, appErrors.ErrInternalServerError))
}

func TestRegister_TrimsInput(t *testing.T) {
	userRepo := new(testutil.MockUserRepo)
	svc := setupAuthService(userRepo, nil)

	userRepo.On("ExistsByEmail", mock.Anything, "trimmed@example.com").Return(false, nil)
	userRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).
		Run(func(args mock.Arguments) {
			u := args.Get(1).(*models.User)
			assert.Equal(t, "Test User", u.FullName)
			assert.Equal(t, "trimmed@example.com", u.Email)
		}).
		Return(nil)

	_, err := svc.Register(context.Background(), &RegisterForm{
		FullName:        "  Test User  ",
		Email:           "  Trimmed@Example.COM  ",
		Password:        "password123",
		PasswordConfirm: "password123",
	})

	require.NoError(t, err)
}

// --- VerifyEmail tests ------------------------------------------------

func TestVerifyEmail_Success(t *testing.T) {
	userRepo := new(testutil.MockUserRepo)
	svc := setupAuthService(userRepo, nil)

	user := testutil.NewUnverifiedUser()
	userRepo.On("FindByVerifyToken", mock.Anything, "valid-verify-token").Return(user, nil)
	userRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)

	got, err := svc.VerifyEmail(context.Background(), "valid-verify-token")

	require.NoError(t, err)
	assert.True(t, got.EmailVerified)
	assert.Equal(t, constants.StatusActive, got.Status)
	assert.Empty(t, got.VerifyToken)
	userRepo.AssertExpectations(t)
}

func TestVerifyEmail_EmptyToken(t *testing.T) {
	svc := setupAuthService(nil, nil)

	_, err := svc.VerifyEmail(context.Background(), "")

	require.Error(t, err)
	var appErr *appErrors.AppError
	assert.ErrorAs(t, err, &appErr)
	assert.Equal(t, 400, appErr.Status)
}

func TestVerifyEmail_TokenNotFound(t *testing.T) {
	userRepo := new(testutil.MockUserRepo)
	svc := setupAuthService(userRepo, nil)

	userRepo.On("FindByVerifyToken", mock.Anything, "bad-token").
		Return(nil, fmt.Errorf("find: %w", gorm.ErrRecordNotFound))

	_, err := svc.VerifyEmail(context.Background(), "bad-token")

	require.Error(t, err)
	var appErr *appErrors.AppError
	assert.ErrorAs(t, err, &appErr)
	assert.Equal(t, 400, appErr.Status)
}

func TestVerifyEmail_TokenExpired(t *testing.T) {
	userRepo := new(testutil.MockUserRepo)
	svc := setupAuthService(userRepo, nil)

	user := testutil.NewUnverifiedUser()
	expired := time.Now().Add(-1 * time.Hour)
	user.VerifyTokenExpiry = &expired
	userRepo.On("FindByVerifyToken", mock.Anything, "expired-token").Return(user, nil)

	_, err := svc.VerifyEmail(context.Background(), "expired-token")

	require.Error(t, err)
	var appErr *appErrors.AppError
	assert.ErrorAs(t, err, &appErr)
	assert.Equal(t, 400, appErr.Status)
}

func TestVerifyEmail_AlreadyVerified(t *testing.T) {
	userRepo := new(testutil.MockUserRepo)
	svc := setupAuthService(userRepo, nil)

	user := testutil.NewActiveUser()
	user.VerifyToken = "some-token"
	userRepo.On("FindByVerifyToken", mock.Anything, "some-token").Return(user, nil)

	got, err := svc.VerifyEmail(context.Background(), "some-token")

	require.NoError(t, err)
	assert.True(t, got.EmailVerified)
	// Update should NOT be called since already verified
	userRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

func TestVerifyEmail_UpdateFails(t *testing.T) {
	userRepo := new(testutil.MockUserRepo)
	svc := setupAuthService(userRepo, nil)

	user := testutil.NewUnverifiedUser()
	userRepo.On("FindByVerifyToken", mock.Anything, "valid-verify-token").Return(user, nil)
	userRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.User")).
		Return(fmt.Errorf("db error"))

	_, err := svc.VerifyEmail(context.Background(), "valid-verify-token")

	assert.True(t, appErrors.Is(err, appErrors.ErrInternalServerError))
}

// --- OAuthLogin tests -------------------------------------------------

func TestOAuthLogin_ExistingSocialAccount(t *testing.T) {
	userRepo := new(testutil.MockUserRepo)
	socialRepo := new(testutil.MockSocialAccountRepo)
	svc := setupAuthService(userRepo, socialRepo)

	user := testutil.NewActiveUser()
	socialAcct := &models.SocialAccount{ID: 1, UserID: user.ID, Provider: "google", ProviderID: "g123"}

	socialRepo.On("FindByProvider", mock.Anything, "google", "g123").Return(socialAcct, nil)
	userRepo.On("FindByID", mock.Anything, user.ID).Return(user, nil)

	got, err := svc.OAuthLogin(context.Background(), "google", "g123", "user@example.com", "User", "")

	require.NoError(t, err)
	assert.Equal(t, user.ID, got.ID)
	socialRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestOAuthLogin_ExistingSocialAccount_BannedUser(t *testing.T) {
	userRepo := new(testutil.MockUserRepo)
	socialRepo := new(testutil.MockSocialAccountRepo)
	svc := setupAuthService(userRepo, socialRepo)

	user := testutil.NewBannedUser()
	socialAcct := &models.SocialAccount{ID: 1, UserID: user.ID, Provider: "google", ProviderID: "g123"}

	socialRepo.On("FindByProvider", mock.Anything, "google", "g123").Return(socialAcct, nil)
	userRepo.On("FindByID", mock.Anything, user.ID).Return(user, nil)

	_, err := svc.OAuthLogin(context.Background(), "google", "g123", "user@example.com", "User", "")

	assert.True(t, appErrors.Is(err, ErrAccountBanned))
}

func TestOAuthLogin_ExistingSocialAccount_AdminUser(t *testing.T) {
	userRepo := new(testutil.MockUserRepo)
	socialRepo := new(testutil.MockSocialAccountRepo)
	svc := setupAuthService(userRepo, socialRepo)

	admin := testutil.NewAdminUser()
	socialAcct := &models.SocialAccount{ID: 1, UserID: admin.ID, Provider: "google", ProviderID: "g123"}

	socialRepo.On("FindByProvider", mock.Anything, "google", "g123").Return(socialAcct, nil)
	userRepo.On("FindByID", mock.Anything, admin.ID).Return(admin, nil)

	_, err := svc.OAuthLogin(context.Background(), "google", "g123", "admin@example.com", "Admin", "")

	assert.True(t, appErrors.Is(err, ErrAdminMustUsePortal))
}

func TestOAuthLogin_ExistingEmailUser_LinksSocialAccount(t *testing.T) {
	userRepo := new(testutil.MockUserRepo)
	socialRepo := new(testutil.MockSocialAccountRepo)
	svc := setupAuthService(userRepo, socialRepo)

	user := testutil.NewActiveUser()

	socialRepo.On("FindByProvider", mock.Anything, "facebook", "fb456").
		Return(nil, fmt.Errorf("find social account: %w", gorm.ErrRecordNotFound))
	userRepo.On("FindByEmail", mock.Anything, testutil.TestEmail).Return(user, nil)
	socialRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.SocialAccount")).Return(nil)

	got, err := svc.OAuthLogin(context.Background(), "facebook", "fb456", testutil.TestEmail, "User", "")

	require.NoError(t, err)
	assert.Equal(t, user.ID, got.ID)
	socialRepo.AssertCalled(t, "Create", mock.Anything, mock.AnythingOfType("*models.SocialAccount"))
}

func TestOAuthLogin_ExistingEmailUser_Banned(t *testing.T) {
	userRepo := new(testutil.MockUserRepo)
	socialRepo := new(testutil.MockSocialAccountRepo)
	svc := setupAuthService(userRepo, socialRepo)

	user := testutil.NewBannedUser()

	socialRepo.On("FindByProvider", mock.Anything, "facebook", "fb456").
		Return(nil, fmt.Errorf("find social account: %w", gorm.ErrRecordNotFound))
	userRepo.On("FindByEmail", mock.Anything, testutil.TestEmail).Return(user, nil)

	_, err := svc.OAuthLogin(context.Background(), "facebook", "fb456", testutil.TestEmail, "User", "")

	assert.True(t, appErrors.Is(err, ErrAccountBanned))
}

func TestOAuthLogin_EmptyEmail(t *testing.T) {
	svc := setupAuthService(nil, nil)

	_, err := svc.OAuthLogin(context.Background(), "google", "g123", "", "User", "")

	require.Error(t, err)
	var appErr *appErrors.AppError
	assert.ErrorAs(t, err, &appErr)
	assert.Equal(t, 422, appErr.Status)
}

func TestOAuthLogin_SocialAccountDBError(t *testing.T) {
	userRepo := new(testutil.MockUserRepo)
	socialRepo := new(testutil.MockSocialAccountRepo)
	svc := setupAuthService(userRepo, socialRepo)

	socialRepo.On("FindByProvider", mock.Anything, "google", "g123").
		Return(nil, fmt.Errorf("db error"))

	_, err := svc.OAuthLogin(context.Background(), "google", "g123", "user@example.com", "User", "")

	assert.True(t, appErrors.Is(err, appErrors.ErrInternalServerError))
}

func TestOAuthLogin_ExistingSocialAccount_InactiveUser(t *testing.T) {
	userRepo := new(testutil.MockUserRepo)
	socialRepo := new(testutil.MockSocialAccountRepo)
	svc := setupAuthService(userRepo, socialRepo)

	user := testutil.NewInactiveUser()
	socialAcct := &models.SocialAccount{ID: 1, UserID: user.ID, Provider: "google", ProviderID: "g123"}

	socialRepo.On("FindByProvider", mock.Anything, "google", "g123").Return(socialAcct, nil)
	userRepo.On("FindByID", mock.Anything, user.ID).Return(user, nil)

	_, err := svc.OAuthLogin(context.Background(), "google", "g123", "user@example.com", "User", "")

	assert.True(t, appErrors.Is(err, ErrAccountInactive))
}

// --- EmailVerificationRequired ----------------------------------------

func TestEmailVerificationRequired_NilEmailService(t *testing.T) {
	svc := setupAuthService(nil, nil)
	assert.False(t, svc.EmailVerificationRequired())
}
