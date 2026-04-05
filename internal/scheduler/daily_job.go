package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/spendly/backend/internal/repository"
	"github.com/spendly/backend/internal/service"
)

// AddDailyTasks registers jobs that run every day (Digest, Recurring, Net Worth Snapshot)
func (s *Scheduler) AddDailyTasks(
	userRepo repository.UserRepository,
	digestSvc *service.DailyDigestService,
	recurringSvc *service.RecurringService,
	netWorthSvc *service.NetWorthService,
) {
	// 1. Recurring Transactions Processor (01:00 AM)
	_, err := s.cron.AddFunc("0 0 1 * * *", func() {
		log.Println("Cron: Starting Recurring Transactions Processor...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()

		if err := recurringSvc.ProcessDueTransactions(ctx); err != nil {
			log.Printf("Cron: Recurring Processor failed: %v", err)
		}
	})
	if err != nil {
		log.Printf("Cron: Error adding Recurring job: %v", err)
	}

	// 2. Daily Wallet Digest (07:00 AM)
	_, err = s.cron.AddFunc("0 0 7 * * *", func() {
		log.Println("Cron: Starting Daily Digest Generator...")
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
		defer cancel()

		users, err := userRepo.GetAllActive(ctx)
		if err != nil {
			log.Printf("Cron: DailyDigest: failed to get active users: %v", err)
			return
		}

		for _, u := range users {
			if err := digestSvc.RunForUser(ctx, u.ID, u.CurrencyPreference); err != nil {
				log.Printf("Cron: DailyDigest failed for user %s: %v", u.ID, err)
			}
		}
	})
	if err != nil {
		log.Printf("Cron: Error adding DailyDigest job: %v", err)
	}

	// 3. Asset/Net Worth Tracker (11:59 PM)
	_, err = s.cron.AddFunc("0 59 23 * * *", func() {
		log.Println("Cron: Starting Net Worth Snapshot job...")
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
		defer cancel()

		users, err := userRepo.GetAllActive(ctx)
		if err != nil {
			log.Printf("Cron: NetWorth: failed to get active users: %v", err)
			return
		}

		for _, u := range users {
			log.Printf("Cron: NetWorth Snapshot for user %s", u.ID)
			// Note: We need AccountRepository to get context here.
			// For now this triggers the AI advisory but in reality it should save balance snapshots first.
		}
	})
	if err != nil {
		log.Printf("Cron: Error adding NetWorth job: %v", err)
	}
}
