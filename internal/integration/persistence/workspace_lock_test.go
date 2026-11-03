package persistence

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestWorkspaceLockSerialisesTheSameWorkspace(t *testing.T) {
	lock := NewWorkspaceLockPersistence()
	_, unlock := lock.LockWorkspace(context.Background(), "ws-1")
	acquired := make(chan struct{})
	go func() {
		_, unlockSecond := lock.LockWorkspace(context.Background(), "ws-1")
		close(acquired)
		unlockSecond()
	}()
	select {
	case <-acquired:
		t.Fatal("second LockWorkspace() for the same workspace did not wait")
	case <-time.After(50 * time.Millisecond):
	}
	unlock()
	select {
	case <-acquired:
	case <-time.After(time.Second):
		t.Fatal("second LockWorkspace() never acquired after unlock")
	}
}

func TestWorkspaceLockDoesNotBlockOtherWorkspaces(t *testing.T) {
	lock := NewWorkspaceLockPersistence()
	_, unlock := lock.LockWorkspace(context.Background(), "ws-1")
	defer unlock()
	acquired := make(chan struct{})
	go func() {
		_, unlockOther := lock.LockWorkspace(context.Background(), "ws-2")
		unlockOther()
		close(acquired)
	}()
	select {
	case <-acquired:
	case <-time.After(time.Second):
		t.Fatal("LockWorkspace() for another workspace blocked")
	}
}

func TestWorkspaceLockIsReentrantForTheReturnedContext(t *testing.T) {
	lock := NewWorkspaceLockPersistence()
	ctx, unlock := lock.LockWorkspace(context.Background(), "ws-1")
	done := make(chan struct{})
	go func() {
		nested, unlockNested := lock.LockWorkspace(ctx, "ws-1")
		_, unlockOther := lock.LockWorkspace(nested, "ws-2")
		unlockOther()
		unlockNested()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("nested LockWorkspace() with the returned context deadlocked")
	}
	unlock()
	unlock()
	_, unlockAgain := lock.LockWorkspace(context.Background(), "ws-1")
	unlockAgain()
}

func TestWorkspaceLockLocksAgainWithAReleasedContext(t *testing.T) {
	lock := NewWorkspaceLockPersistence()
	released, unlock := lock.LockWorkspace(context.Background(), "ws-1")
	unlock()
	_, unlockHolder := lock.LockWorkspace(context.Background(), "ws-1")
	acquired := make(chan struct{})
	go func() {
		_, unlockStale := lock.LockWorkspace(released, "ws-1")
		close(acquired)
		unlockStale()
	}()
	select {
	case <-acquired:
		t.Fatal("LockWorkspace() with a released context skipped the lock")
	case <-time.After(50 * time.Millisecond):
	}
	unlockHolder()
	<-acquired
}

func TestWorkspaceLockKeepsConcurrentWritersExclusive(t *testing.T) {
	lock := NewWorkspaceLockPersistence()
	var wait sync.WaitGroup
	counter := 0
	for range 50 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			_, unlock := lock.LockWorkspace(context.Background(), "ws-1")
			defer unlock()
			value := counter
			time.Sleep(time.Microsecond)
			counter = value + 1
		}()
	}
	wait.Wait()
	if counter != 50 {
		t.Fatalf("counter = %d, want 50", counter)
	}
}
