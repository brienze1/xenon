package webservice

import (
	"context"
	"sync"
	"time"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/pkg/logger"
)

type scheduledJob struct {
	stop       func() bool
	generation int
}

type syncScheduler struct {
	timers     adapter.Timers
	pending    map[string]scheduledJob
	generation int
	mu         sync.Mutex
}

func NewSyncScheduler(timers adapter.Timers) adapter.SyncScheduler {
	return &syncScheduler{timers: timers, pending: map[string]scheduledJob{}}
}

func (s *syncScheduler) Schedule(ctx context.Context, key string, delay time.Duration, job func(ctx context.Context)) {
	logger.Debug(ctx, "Started", map[string]any{"key": key, "delay": delay.String()})

	detached := context.WithoutCancel(ctx)
	s.mu.Lock()
	defer s.mu.Unlock()
	if previous, found := s.pending[key]; found {
		previous.stop()
	}
	s.generation++
	generation := s.generation
	stop := s.timers.AfterFunc(delay, func() {
		if s.take(key, generation) {
			job(detached)
		}
	})
	s.pending[key] = scheduledJob{stop: stop, generation: generation}

	logger.Debug(ctx, "Finished", map[string]any{"key": key, "delay": delay.String()})
}

func (s *syncScheduler) take(key string, generation int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.pending[key].generation != generation {
		return false
	}
	delete(s.pending, key)
	return true
}
