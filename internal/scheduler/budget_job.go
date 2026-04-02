package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/spendly/backend/internal/repository"
)

// AddDailyBudgetJob registers a daily budget monitoring job.
func (s *Scheduler) AddDailyBudgetJob(userRepo repository.UserRepository, budgetRepo repository.BudgetRepository) {
	_, err := s.cron.AddFunc("0 0 0 * * *", func() {
		log.Println("Starting daily budget monitor job...")
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
		defer cancel()

		users, err := userRepo.GetAllActive(ctx)
		if err != nil {
			log.Printf("Failed to get active users for budget check: %v", err)
			return
		}

		for _, user := range users {
			log.Printf("Checking budgets for user %s", user.ID)
			budgets, err := budgetRepo.GetActiveBudgetsByUser(ctx, user.ID)
			if err != nil {
				log.Printf("Failed to fetch budgets for user %s: %v", user.ID, err)
				continue
			}

			for _, b := range budgets {
				spent, _ := budgetRepo.GetSpentAmount(ctx, b.ID)
				limitF, _ := b.LimitAmount.Float64()
				if limitF > 0 && spent >= limitF {
					log.Printf("Budget exceeded for user %s, category %v", user.ID, b.CategoryID)
					// Trigger alert service if needed
				}
			}
		}
	})
	if err != nil {
		log.Printf("Error adding daily budget job: %v", err)
	}
}
