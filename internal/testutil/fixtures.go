package testutil

import (
	"time"

	"sun-booking-tours/internal/constants"
	"sun-booking-tours/internal/models"

	"golang.org/x/crypto/bcrypt"
)

const (
	TestPassword      = "password123"
	TestEmail         = "user@example.com"
	TestAdminEmail    = "admin@example.com"
	TestFullName      = "Test User"
	TestAdminFullName = "Admin User"
	TestBaseURL       = "http://localhost:8080"
	TestSessionSecret = "test-secret-key-for-testing-only"
)

// HashedPassword returns a bcrypt hash of the given plain password.
func HashedPassword(plain string) string {
	h, _ := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.MinCost)
	return string(h)
}

// NewActiveUser builds a user model with role=user, status=active.
func NewActiveUser() *models.User {
	return &models.User{
		ID:            1,
		Email:         TestEmail,
		Password:      HashedPassword(TestPassword),
		FullName:      TestFullName,
		Role:          constants.RoleUser,
		Status:        constants.StatusActive,
		EmailVerified: true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// NewAdminUser builds a user model with role=admin, status=active.
func NewAdminUser() *models.User {
	return &models.User{
		ID:            2,
		Email:         TestAdminEmail,
		Password:      HashedPassword(TestPassword),
		FullName:      TestAdminFullName,
		Role:          constants.RoleAdmin,
		Status:        constants.StatusActive,
		EmailVerified: true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// NewBannedUser builds a user model with status=banned.
func NewBannedUser() *models.User {
	u := NewActiveUser()
	u.Status = constants.StatusBanned
	return u
}

// NewInactiveUser builds a user model with status=inactive.
func NewInactiveUser() *models.User {
	u := NewActiveUser()
	u.Status = constants.StatusInactive
	return u
}

// NewUnverifiedUser builds a user with email verification pending.
func NewUnverifiedUser() *models.User {
	u := NewActiveUser()
	u.EmailVerified = false
	u.Status = constants.StatusInactive
	u.VerifyToken = "valid-verify-token"
	exp := time.Now().Add(24 * time.Hour)
	u.VerifyTokenExpiry = &exp
	return u
}
