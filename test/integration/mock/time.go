package mock

import (
	"sort"
	"sync"
	"time"
)

type timer struct {
	due     time.Time
	run     func()
	id      int
	stopped bool
}

type Time struct {
	currentTime time.Time
	updatedAt   time.Time
	timers      []*timer
	nextTimer   int
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
	t.mu.Lock()
	defer t.mu.Unlock()
	t.timers = nil
}

func (t *Time) Now() time.Time {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.now()
}

func (t *Time) now() time.Time {
	return t.currentTime.Add(time.Since(t.updatedAt)).UTC()
}

func (t *Time) AfterFunc(delay time.Duration, run func()) func() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.nextTimer++
	scheduled := &timer{id: t.nextTimer, due: t.now().Add(delay), run: run}
	t.timers = append(t.timers, scheduled)
	return func() bool {
		t.mu.Lock()
		defer t.mu.Unlock()
		pending := !scheduled.stopped
		scheduled.stopped = true
		return pending
	}
}

func (t *Time) Advance(duration time.Duration) {
	t.mu.Lock()
	t.currentTime = t.currentTime.Add(duration)
	t.mu.Unlock()
	for {
		next := t.nextDue()
		if next == nil {
			return
		}
		next.run()
	}
}

func (t *Time) nextDue() *timer {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := t.now()
	sort.SliceStable(t.timers, func(i, j int) bool { return t.timers[i].due.Before(t.timers[j].due) })
	for index, scheduled := range t.timers {
		if scheduled.stopped || scheduled.due.After(now) {
			continue
		}
		t.timers = append(t.timers[:index], t.timers[index+1:]...)
		scheduled.stopped = true
		return scheduled
	}
	return nil
}
