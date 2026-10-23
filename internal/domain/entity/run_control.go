package entity

import "sync"

// RunControl is how callers steer a run's goroutine: pause and resume gate the next row, stop cancels it.
type RunControl struct {
	stop    func()
	resumed chan struct{}
	done    chan struct{}
	mutex   sync.Mutex
	paused  bool
}

func NewRunControl(stop func()) *RunControl {
	return &RunControl{stop: stop, done: make(chan struct{})}
}

func (r *RunControl) Pause() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if !r.paused {
		r.paused = true
		r.resumed = make(chan struct{})
	}
}

func (r *RunControl) Resume() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.paused {
		r.paused = false
		close(r.resumed)
	}
}

func (r *RunControl) Stop() {
	r.stop()
	r.Resume()
}

// Resumed returns the channel closed on resume, and false when the run is not paused.
func (r *RunControl) Resumed() (<-chan struct{}, bool) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.resumed, r.paused
}

func (r *RunControl) Finished() {
	close(r.done)
}

func (r *RunControl) Done() <-chan struct{} {
	return r.done
}
