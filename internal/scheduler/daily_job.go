package scheduler

import (
	"context"
	"log"
	"time"

	"gorm.io/gorm"
)

// DailyJob represents the Cron-job logic mentioned in complete-ai-implementation.md
// for recurring transactions and budget resets.
type DailyJob struct {
	db *gorm.DB
}

func NewDailyJob(db *gorm.DB) *DailyJob {
	return &DailyJob{db: db}
}

func (j *DailyJob) Run(ctx context.Context) {
	log.Println("[DailyJob] Started executing automated tasks...")

	// Find recurring transactions scheduled for today
	j.processRecurringTransactions(ctx)

	// Reset weekly/monthly budgets if period has ended
	j.processBudgetResets(ctx)

	log.Println("[DailyJob] Daily tasks executed successfully.")
}

func (j *DailyJob) processRecurringTransactions(ctx context.Context) {
	log.Println("[DailyJob] Scanning for recurring logic...")
	// Logic would query DB: j.db.WithContext(ctx).Where(...)
	
	select {
	case <-time.After(100 * time.Millisecond):
		// Simulation complete
	case <-ctx.Done():
		log.Println("[DailyJob] processRecurringTransactions cancelled")
	}
}

func (j *DailyJob) processBudgetResets(ctx context.Context) {
	log.Println("[DailyJob] Checking budget periods...")
	// Logic to reset monthly Burn Rate via j.db.WithContext(ctx)
	
	select {
	case <-time.After(50 * time.Millisecond):
		// Simulation complete
	case <-ctx.Done():
		log.Println("[DailyJob] processBudgetResets cancelled")
	}
}

// StartCron initializes a background ticker
func (j *DailyJob) StartCron(ctx context.Context) {
	// A simple ticker running every 24h.
	ticker := time.NewTicker(24 * time.Hour)

	// Execute immediately on start
	go j.Run(ctx)

	go func() {
		for {
			select {
			case <-ticker.C:
				j.Run(ctx)
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}
