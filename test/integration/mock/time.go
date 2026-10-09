package mock

import (
	"sync"
	"time"
)

type Time struct {
	currentTime time.Time
	updatedAt   time.Time
	mu          sync.RWMutex
}

func NewTime() *Time {
	now := time.Now().UTC()
	return &Time{currentTime: now, updatedAt: now}
}

func (t *Time) SetCurrentTime(currentTime time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.currentTime = currentTime.UTC()
	t.updatedAt = time.Now().UTC()
}

func (t *Time) Reset() {
	t.SetCurrentTime(time.Now())
}

func (t *Time) Now() time.Time {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.currentTime.Add(time.Since(t.updatedAt)).UTC()
}
