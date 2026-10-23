package service

import (
	"context"
	"testing"
)

func TestRunControlGatesOnPauseAndStop(t *testing.T) {
	stopped := false
	control := newRunControl(func() { stopped = true })
	if _, paused := control.awaitResume(); paused {
		t.Fatal("a new run is not paused")
	}
	control.Pause()
	control.Pause()
	resumed, paused := control.awaitResume()
	if !paused {
		t.Fatal("pause holds the run")
	}
	control.Resume()
	control.Resume()
	if _, open := <-resumed; open {
		t.Error("resume releases the waiting row")
	}
	control.Pause()
	resumed, _ = control.awaitResume()
	control.Stop()
	if _, open := <-resumed; open || !stopped {
		t.Error("stop cancels the run and releases a paused one")
	}
	control.finished()
	if !control.AwaitFinished(context.Background()) {
		t.Error("done is closed when the run finished")
	}
}

func TestRunControlAwaitFinishedEndsWithTheContext(t *testing.T) {
	control := newRunControl(func() {})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if control.AwaitFinished(ctx) {
		t.Error("an unfinished run is not awaited past its context")
	}
}

func TestRunControlAbandonStopsTheRun(t *testing.T) {
	stopped := false
	control := newRunControl(func() { stopped = true })
	if control.isAbandoned() {
		t.Fatal("a new run is not abandoned")
	}
	control.Abandon()
	if !control.isAbandoned() || !stopped {
		t.Error("abandon marks the run and stops it")
	}
}
