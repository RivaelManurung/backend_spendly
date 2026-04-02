package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/spendly/backend/internal/domain"
	"github.com/spendly/backend/internal/repository"
)

var _ repository.UserRepository = (*userRepository)(nil)

type userRepository struct {
	db *sqlx.DB
}

// NewUserRepository creates a new instance of postgres repo for users.
func NewUserRepository(db *sqlx.DB) repository.UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	now := time.Now()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}
	user.UpdatedAt = now

	query := `
		INSERT INTO users (id, email, name, avatar_url, status, currency_preference, created_at, updated_at)
		VALUES (:id, :email, :name, :avatar_url, :status, :currency_preference, :created_at, :updated_at)
	`
	_, err := r.db.NamedExecContext(ctx, query, user)
	if err != nil {
		return fmt.Errorf("userRepository.Create: %w", err)
	}

	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT 	id, email, name, avatar_url, status, currency_preference, created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`
	var user domain.User
	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("userRepository.GetByID: %w", err)
	}

	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT 	id, email, name, avatar_url, status, currency_preference, created_at, updated_at, deleted_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`
	var user domain.User
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("userRepository.GetByEmail: %w", err)
	}

	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	user.UpdatedAt = time.Now()
	query := `
		UPDATE users
		SET name = :name,
			avatar_url = :avatar_url,
			status = :status,
			currency_preference = :currency_preference,
			updated_at = :updated_at
		WHERE id = :id AND deleted_at IS NULL
	`
	_, err := r.db.NamedExecContext(ctx, query, user)
	if err != nil {
		return fmt.Errorf("userRepository.Update: %w", err)
	}
	return nil
}

func (r *userRepository) DeleteSoft(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE users SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("userRepository.DeleteSoft: %w", err)
	}
	return nil
}

func (r *userRepository) UpdateCurrencyPreference(ctx context.Context, userID uuid.UUID, currency string) error {
	query := `UPDATE users SET currency_preference = $1, updated_at = NOW() WHERE id = $2 AND deleted_at IS NULL`
	_, err := r.db.ExecContext(ctx, query, currency, userID)
	if err != nil {
		return fmt.Errorf("userRepository.UpdateCurrencyPreference: %w", err)
	}
	return nil
}

func (r *userRepository) GetAllActive(ctx context.Context) ([]domain.User, error) {
	query := `SELECT id, email, name, avatar_url, status, currency_preference, created_at, updated_at FROM users WHERE deleted_at IS NULL`
	var users []domain.User
	if err := r.db.SelectContext(ctx, &users, query); err != nil {
		return nil, fmt.Errorf("userRepository.GetAllActive: %w", err)
	}
	if users == nil {
		users = []domain.User{}
	}
	return users, nil
}
