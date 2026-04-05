package scheduler

import (
	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron *cron.Cron
}

func NewScheduler() *Scheduler {
	// withSeconds: true allows cron to run frequently for testing if needed
	c := cron.New(cron.WithSeconds())
	return &Scheduler{
		cron: c,
	}
}

func (s *Scheduler) Start() {
	s.cron.Start()
}

func (s *Scheduler) Stop() {
	s.cron.Stop()
}
