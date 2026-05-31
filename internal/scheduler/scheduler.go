package scheduler

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/krishnassh/picostatus/internal/checker"
	"github.com/krishnassh/picostatus/internal/storage"
)

type Scheduler struct {
	repo *storage.Repository
}

func New(repo *storage.Repository) *Scheduler {
	return &Scheduler{repo: repo}
}

func (s *Scheduler) Start(ctx context.Context) {
	checks, err := s.repo.GetChecks()
	if err != nil {
		log.Printf("scheduler: failed to load checks: %v", err)
		return
	}

	var wg sync.WaitGroup
	for _, c := range checks {
		wg.Add(1)
		go func(c storage.Check) {
			defer wg.Done()
			s.runLoop(ctx, c)
		}(c)
	}

	wg.Wait()
}

func (s *Scheduler) runLoop(ctx context.Context, c storage.Check) {
	s.runOnce(ctx, c)

	ticker := time.NewTicker(c.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.runOnce(ctx, c)
		}
	}
}

func (s *Scheduler) runOnce(ctx context.Context, c storage.Check) {
	result := checker.HTTPChecker{URL: c.Target}.Run(ctx)

	status := storage.StatusUp
	if !result.Success {
		status = storage.StatusDown
	}

	if _, err := s.repo.InsertResult(c.ID, status, result.Success, result.LatencyMS, result.Error); err != nil {
		log.Printf("scheduler: failed to save result for %q: %v", c.Name, err)
	}
}
