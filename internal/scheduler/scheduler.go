package scheduler

import (
	"github.com/robfig/cron/v3"
	"github.com/spendly/backend/internal/repository"
)

type Scheduler struct {
	cron *cron.Cron
	repo repository.AnalysisRepository // etc.
}

func NewScheduler(repo repository.AnalysisRepository) *Scheduler {
	c := cron.New(cron.WithSeconds())
	return &Scheduler{
		cron: c,
		repo: repo,
	}
}

func (s *Scheduler) Start() {
	s.cron.Start()
}

func (s *Scheduler) Stop() {
	s.cron.Stop()
}
