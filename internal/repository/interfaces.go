package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/spendly/backend/internal/domain"
)

// UserRepository contains all DB operations for Users.
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	DeleteSoft(ctx context.Context, id uuid.UUID) error
	UpdateCurrencyPreference(ctx context.Context, userID uuid.UUID, currency string) error
	GetAllActive(ctx context.Context) ([]domain.User, error)
}

// TransactionRepository contains database operations for Transactions.
type TransactionRepository interface {
	Insert(ctx context.Context, txn *domain.Transaction) error
	BulkInsert(ctx context.Context, txns []domain.Transaction) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Transaction, error)
	GetByDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]domain.Transaction, error)
	FindAllByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Transaction, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
}

// AnalysisRepository contains database operations for Analysis Snapshots.
type AnalysisRepository interface {
	UpsertSnapshot(ctx context.Context, snapshot *domain.AnalysisSnapshot) error
	GetByPeriod(ctx context.Context, userID uuid.UUID, period string) (*domain.AnalysisSnapshot, error)
	GetLatestByUserID(ctx context.Context, userID uuid.UUID, periodType string) (*domain.AnalysisSnapshot, error)
}

// InsightRepository contains database operations for AI Insights.
type InsightRepository interface {
	SaveInsight(ctx context.Context, insight *domain.AIInsight) error
	SaveBatch(ctx context.Context, insights []domain.AIInsight) error
	ListUnread(ctx context.Context, userID uuid.UUID) ([]domain.AIInsight, error)
	MarkRead(ctx context.Context, id int64) error
	SetHelpful(ctx context.Context, id int64, helpful bool) error
	GetLatestByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]domain.AIInsight, error)
}

// BudgetRepository contains database operations for Budgets.
type BudgetRepository interface {
	Create(ctx context.Context, budget *domain.Budget) error
	GetByID(ctx context.Context, id int64) (*domain.Budget, error)
	GetActiveBudgetsByUser(ctx context.Context, userID uuid.UUID) ([]domain.Budget, error)
	GetSpentAmount(ctx context.Context, budgetID int64) (float64, error)
	Update(ctx context.Context, budget *domain.Budget) error
	Delete(ctx context.Context, id int64) error
}

// CategoryRepository contains database operations for Categories.
type CategoryRepository interface {
	ListSystem(ctx context.Context) ([]domain.Category, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Category, error)
	GetByID(ctx context.Context, id int64) (*domain.Category, error)
}

// RecurringRepository handles scheduled/recurring transaction templates.
type RecurringRepository interface {
	Create(ctx context.Context, rec *domain.RecurringTransaction) error
	GetByID(ctx context.Context, id int64) (*domain.RecurringTransaction, error)
	GetActiveByUser(ctx context.Context, userID uuid.UUID) ([]domain.RecurringTransaction, error)
	GetDueBy(ctx context.Context, dueDate time.Time) ([]domain.RecurringTransaction, error)
	UpdateNextDueDate(ctx context.Context, id int64, nextDue time.Time) error
	SetActive(ctx context.Context, id int64, active bool) error
	Delete(ctx context.Context, id int64) error
}

// AccountRepository handles multiple financial accounts per user.
type AccountRepository interface {
	Create(ctx context.Context, account *domain.Account) error
	GetByID(ctx context.Context, id int64) (*domain.Account, error)
	GetAllByUser(ctx context.Context, userID uuid.UUID) ([]domain.Account, error)
	UpdateBalance(ctx context.Context, id int64, newBalance interface{}) error
	Update(ctx context.Context, account *domain.Account) error
	SoftDelete(ctx context.Context, id int64) error
	CreateTransfer(ctx context.Context, transfer *domain.AccountTransfer) error
	GetTransfersByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.AccountTransfer, error)
}
