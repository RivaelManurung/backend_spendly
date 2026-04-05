package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/spendly/backend/internal/ai"
	"github.com/spendly/backend/internal/domain"
	"github.com/spendly/backend/internal/repository"
)

// RecurringService manages recurring transactions and provides AI advisory.
// Responsible for:
// 1. Processing due recurring transactions (auto-post or remind)
// 2. Running the recurring_advisor.prompt for proactive cash flow advice
type RecurringService struct {
	gemini      *ai.GeminiClient
	repoRecurring repository.RecurringRepository
	repoTxn     repository.TransactionRepository
	repoInsight repository.InsightRepository
}

func NewRecurringService(
	gemini *ai.GeminiClient,
	repoRecurring repository.RecurringRepository,
	repoTxn repository.TransactionRepository,
	repoInsight repository.InsightRepository,
) *RecurringService {
	return &RecurringService{
		gemini:        gemini,
		repoRecurring: repoRecurring,
		repoTxn:       repoTxn,
		repoInsight:   repoInsight,
	}
}

// ProcessDueTransactions checks all recurring transactions due today and either
// auto-posts them or creates a "reminder" insight for the user.
func (s *RecurringService) ProcessDueTransactions(ctx context.Context) error {
	today := time.Now().Truncate(24 * time.Hour)

	// Fetch all active recurring transactions due on or before today
	dueList, err := s.repoRecurring.GetDueBy(ctx, today)
	if err != nil {
		return fmt.Errorf("RecurringService.ProcessDueTransactions: %w", err)
	}

	for _, rec := range dueList {
		if rec.AutoPost {
			// Auto-create the transaction
			if err := s.autoPost(ctx, &rec); err != nil {
				fmt.Printf("RecurringService: failed to auto-post %q: %v\n", rec.Title, err)
			}
		} else {
			// Send reminder insight to user
			if err := s.sendReminder(ctx, &rec); err != nil {
				fmt.Printf("RecurringService: failed to send reminder for %q: %v\n", rec.Title, err)
			}
		}

		// Update NextDueDate
		if err := s.repoRecurring.UpdateNextDueDate(ctx, rec.ID, s.calcNextDue(rec)); err != nil {
			fmt.Printf("RecurringService: failed to update next_due_date for %d: %v\n", rec.ID, err)
		}
	}
	return nil
}

// autoPost creates a real transaction from a recurring template.
func (s *RecurringService) autoPost(ctx context.Context, rec *domain.RecurringTransaction) error {
	txn := &domain.Transaction{
		UserID:          rec.UserID,
		CategoryID:      rec.CategoryID,
		Amount:          rec.Amount,
		Currency:        rec.Currency,
		AmountInBase:    rec.Amount, // TODO: apply FX conversion if needed
		TransactionDate: time.Now(),
		Source:          "recurring",
	}
	title := rec.Title
	txn.Description = &title
	return s.repoTxn.Insert(ctx, txn)
}

// sendReminder creates an AI insight to remind the user of an upcoming payment.
func (s *RecurringService) sendReminder(ctx context.Context, rec *domain.RecurringTransaction) error {
	msg := fmt.Sprintf("Pengingat: %s sebesar %s jatuh tempo hari ini.", rec.Title, rec.Amount.String())
	insight := domain.AIInsight{
		UserID:   rec.UserID,
		Type:     "RECURRING_REMINDER",
		Title:    "🔔 Tagihan Jatuh Tempo",
		Content:  msg,
		Priority: 3, // High priority
	}
	return s.repoInsight.SaveInsight(ctx, &insight)
}

// calcNextDue calculates the next due date based on frequency.
func (s *RecurringService) calcNextDue(rec domain.RecurringTransaction) time.Time {
	switch rec.Frequency {
	case domain.FrequencyDaily:
		return rec.NextDueDate.AddDate(0, 0, 1)
	case domain.FrequencyWeekly:
		return rec.NextDueDate.AddDate(0, 0, 7)
	case domain.FrequencyMonthly:
		return rec.NextDueDate.AddDate(0, 1, 0)
	case domain.FrequencyYearly:
		return rec.NextDueDate.AddDate(1, 0, 0)
	default:
		return rec.NextDueDate.AddDate(0, 1, 0)
	}
}

// RunAdvisoryForUser runs recurring_advisor.prompt and saves AI insight.
func (s *RecurringService) RunAdvisoryForUser(ctx context.Context, userID uuid.UUID) error {
	// Fetch all active recurring
	allRecurring, err := s.repoRecurring.GetActiveByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("RecurringService.RunAdvisoryForUser: %w", err)
	}
	if len(allRecurring) == 0 {
		return nil
	}

	// Get upcoming 7 days
	now := time.Now()
	upcoming7Days := filterUpcoming(allRecurring, now, now.AddDate(0, 0, 7))

	recurringJSON, _ := json.Marshal(allRecurring)
	upcomingJSON, _ := json.Marshal(upcoming7Days)

	vars := map[string]string{
		"recurring_transactions": string(recurringJSON),
		"upcoming_due":           string(upcomingJSON),
		"accounts_balance":       `{}`, // TODO: inject real account balances
	}

	result, err := s.gemini.ExecutePrompt(ctx, "recurring_advisor", vars, true)
	if err != nil {
		return fmt.Errorf("RecurringService.RunAdvisoryForUser AI: %w", err)
	}

	var parsed struct {
		Warning string `json:"warning"`
	}
	_ = json.Unmarshal([]byte(result), &parsed)

	title := "Analisa Transaksi Rutin"
	if parsed.Warning != "" {
		title = "⚠️ " + title
	}

	insight := domain.AIInsight{
		UserID:   userID,
		Type:     "RECURRING_ADVISORY",
		Title:    title,
		Content:  result,
		Priority: 2,
	}
	return s.repoInsight.SaveInsight(ctx, &insight)
}

// filterUpcoming returns recurring transactions due within a date range.
func filterUpcoming(recs []domain.RecurringTransaction, from, to time.Time) []domain.RecurringTransaction {
	var result []domain.RecurringTransaction
	for _, r := range recs {
		if !r.NextDueDate.Before(from) && !r.NextDueDate.After(to) {
			result = append(result, r)
		}
	}
	return result
}
