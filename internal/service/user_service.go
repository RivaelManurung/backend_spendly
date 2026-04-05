package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/spendly/backend/internal/domain"
	"github.com/spendly/backend/internal/repository"
)

// UserService mendefinisikan contract untuk logic pengguna.
type UserService interface {
	RegisterUser(ctx context.Context, email, name, avatarURL, currencyPref string) (*domain.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	UpdateUser(ctx context.Context, user *domain.User) error
}

type userService struct {
	repo repository.UserRepository
}

// NewUserService menginisialisasi user service dengan dependency inject.
func NewUserService(repo repository.UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}

func (s *userService) RegisterUser(ctx context.Context, email, name, avatarURL, currencyPref string) (*domain.User, error) {
	if email == "" {
		return nil, errors.New("email cannot be empty")
	}

	// Cek apakah user sudah ada
	existingUser, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	if currencyPref == "" {
		currencyPref = "IDR"
	}

	newUser := &domain.User{
		ID:                 uuid.New(),
		Email:              email,
		Status:             "active",
		CurrencyPreference: currencyPref,
		SalaryCycleDay:     1, // Default gajian tanggal 1
		RiskProfile:        "conservative",
		AIAnalystPersona:   "supportive",
	}
	if name != "" {
		newUser.Name = &name
	}
	if avatarURL != "" {
		newUser.AvatarURL = &avatarURL
	}
	if newUser.FinancialGoals == nil {
		newUser.FinancialGoals = []string{}
	}

	if err := s.repo.Create(ctx, newUser); err != nil {
		return nil, fmt.Errorf("failed to register user: %w", err)
	}

	return newUser, nil
}

func (s *userService) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

func (s *userService) UpdateUser(ctx context.Context, user *domain.User) error {
	// Validasi business rules jika ada (misal status aktif/non-aktif)
	if user.Status != "active" && user.Status != "deactivated" {
		return errors.New("invalid user status")
	}

	if user.SalaryCycleDay < 1 || user.SalaryCycleDay > 31 {
		return errors.New("invalid salary cycle day: must be between 1 and 31")
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}
