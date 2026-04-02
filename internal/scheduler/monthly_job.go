package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/spendly/backend/internal/repository"
	"github.com/spendly/backend/internal/service"
)

// AddMonthlyAnalysisJob registers a monthly analysis job (e.g. 1st of every month)
func (s *Scheduler) AddMonthlyAnalysisJob(pipeline *service.AnalysisPipeline, userRepo repository.UserRepository) {
	_, err := s.cron.AddFunc("0 0 0 1 * *", func() {
		log.Println("Starting monthly analysis job...")
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
		defer cancel()

		period := time.Now().AddDate(0, -1, 0).Format("2006-01") // Previous month

		// Implementation: Fetch active users and run the pipeline
		users, err := userRepo.GetAllActive(ctx)
		if err != nil {
			log.Printf("Failed to get active users: %v", err)
			return
		}

		for _, user := range users {
			log.Printf("Starting pipeline for user %s", user.ID)
			if err := pipeline.RunBackgroundMonthlyJobs(ctx, user.ID, period); err != nil {
				log.Printf("Pipeline failed for user %s: %v", user.ID, err)
			}
		}
	})
	if err != nil {
		log.Printf("Error adding monthly analysis job: %v", err)
	}
}
