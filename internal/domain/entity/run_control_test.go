package entity

import "testing"

func TestRunControlGatesOnPauseAndStop(t *testing.T) {
	stopped := false
	control := NewRunControl(func() { stopped = true })
	if _, paused := control.Resumed(); paused {
		t.Fatal("a new run is not paused")
	}
	control.Pause()
	control.Pause()
	resumed, paused := control.Resumed()
	if !paused {
		t.Fatal("pause holds the run")
	}
	control.Resume()
	control.Resume()
	if _, open := <-resumed; open {
		t.Error("resume releases the waiting row")
	}
	control.Pause()
	resumed, _ = control.Resumed()
	control.Stop()
	if _, open := <-resumed; open || !stopped {
		t.Error("stop cancels the run and releases a paused one")
	}
	control.Finished()
	if _, open := <-control.Done(); open {
		t.Error("done is closed when the run finished")
	}
}
