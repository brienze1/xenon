package webservice

import (
	"context"
	"testing"
	"time"
)

type fakeTimer struct {
	run     func()
	delay   time.Duration
	stopped bool
}

type fakeTimers struct {
	scheduled []*fakeTimer
}

func (f *fakeTimers) AfterFunc(delay time.Duration, run func()) func() bool {
	timer := &fakeTimer{run: run, delay: delay}
	f.scheduled = append(f.scheduled, timer)
	return func() bool {
		pending := !timer.stopped
		timer.stopped = true
		return pending
	}
}

type contextKey struct{}

func TestSyncSchedulerRunsTheLatestJobPerKeyDetachedFromCancellation(t *testing.T) {
	timers := &fakeTimers{}
	scheduler := NewSyncScheduler(timers)
	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), contextKey{}, "trace"))
	runs := map[string]int{}
	var jobContext context.Context
	job := func(name string) func(context.Context) {
		return func(ctx context.Context) {
			runs[name]++
			jobContext = ctx
		}
	}
	scheduler.Schedule(ctx, "push:A", 3*time.Second, job("first"))
	scheduler.Schedule(ctx, "push:A", 2*time.Second, job("second"))
	scheduler.Schedule(ctx, "push:B", 0, job("other"))
	cancel()
	if len(timers.scheduled) != 3 || timers.scheduled[1].delay != 2*time.Second || timers.scheduled[2].delay != 0 {
		t.Fatalf("scheduled = %+v", timers.scheduled)
	}
	if !timers.scheduled[0].stopped || timers.scheduled[1].stopped || timers.scheduled[2].stopped {
		t.Fatal("a new job replaces the pending one under the same key only")
	}
	timers.scheduled[0].run()
	if runs["first"] != 0 {
		t.Fatal("a replaced job that fires anyway does not run")
	}
	timers.scheduled[1].run()
	timers.scheduled[1].run()
	timers.scheduled[2].run()
	if runs["second"] != 1 || runs["other"] != 1 {
		t.Fatalf("runs = %v", runs)
	}
	if jobContext.Err() != nil || jobContext.Value(contextKey{}) != "trace" {
		t.Fatal("jobs keep the caller's values but not its cancellation")
	}
	scheduler.Schedule(context.Background(), "push:A", time.Second, job("third"))
	if timers.scheduled[1].stopped {
		t.Fatal("a job that already ran is not pending any more")
	}
	timers.scheduled[3].run()
	if runs["third"] != 1 {
		t.Fatalf("runs = %v", runs)
	}
}
