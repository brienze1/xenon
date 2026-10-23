package service

import (
	"context"
	"sync"
)

type runControl struct {
	stop      func()
	resumed   chan struct{}
	done      chan struct{}
	mutex     sync.Mutex
	paused    bool
	abandoned bool
}

func newRunControl(stop func()) *runControl {
	return &runControl{stop: stop, done: make(chan struct{})}
}

func (r *runControl) Pause() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if !r.paused {
		r.paused = true
		r.resumed = make(chan struct{})
	}
}

func (r *runControl) Resume() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.paused {
		r.paused = false
		close(r.resumed)
	}
}

func (r *runControl) Stop() {
	r.stop()
	r.Resume()
}

func (r *runControl) Abandon() {
	r.mutex.Lock()
	r.abandoned = true
	r.mutex.Unlock()
	r.Stop()
}

func (r *runControl) AwaitFinished(ctx context.Context) bool {
	select {
	case <-r.done:
		return true
	case <-ctx.Done():
		return false
	}
}

func (r *runControl) isAbandoned() bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.abandoned
}

func (r *runControl) awaitResume() (<-chan struct{}, bool) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.resumed, r.paused
}

func (r *runControl) finished() {
	close(r.done)
}
