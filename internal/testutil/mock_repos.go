package testutil

import (
	"context"

	"sun-booking-tours/internal/models"
	"sun-booking-tours/internal/repository"

	"github.com/stretchr/testify/mock"
)

// MockUserRepo is a testify mock for repository.UserRepo.
type MockUserRepo struct {
	mock.Mock
}

var _ repository.UserRepo = (*MockUserRepo)(nil)

func (m *MockUserRepo) FindByID(ctx context.Context, id uint) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepo) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepo) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepo) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepo) FindAll(ctx context.Context, filter repository.UserFilter) ([]models.User, int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]models.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserRepo) FindByIDWithRelations(ctx context.Context, id uint) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepo) UpdateStatus(ctx context.Context, id uint, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockUserRepo) FindByVerifyToken(ctx context.Context, token string) (*models.User, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// MockSocialAccountRepo is a testify mock for repository.SocialAccountRepo.
type MockSocialAccountRepo struct {
	mock.Mock
}

var _ repository.SocialAccountRepo = (*MockSocialAccountRepo)(nil)

func (m *MockSocialAccountRepo) FindByProvider(ctx context.Context, provider, providerID string) (*models.SocialAccount, error) {
	args := m.Called(ctx, provider, providerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SocialAccount), args.Error(1)
}

func (m *MockSocialAccountRepo) Create(ctx context.Context, account *models.SocialAccount) error {
	args := m.Called(ctx, account)
	return args.Error(0)
}
