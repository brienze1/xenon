package adapter

import (
	"testing"
	"time"
)

func TestEmitWithoutRunningApplicationIsDropped(t *testing.T) {
	NewEventEmitter().Emit("workspace:tree-changed", nil)
}

func TestPickFolderWithoutRunningApplicationFails(t *testing.T) {
	path, err := NewDialog().PickFolder("Choose a folder")
	if err == nil || path != "" {
		t.Fatalf("PickFolder() = %q, %v; want an error", path, err)
	}
}

func TestPickSourcesWithoutRunningApplicationFails(t *testing.T) {
	paths, err := NewDialog().PickSources("Import")
	if err == nil || paths != nil {
		t.Fatalf("PickSources() = %v, %v; want an error", paths, err)
	}
}

func TestPickFileAndSaveFileWithoutRunningApplicationFail(t *testing.T) {
	for name, pick := range map[string]func() (string, error){
		"file": func() (string, error) { return NewDialog().PickFile("Choose a file") },
		"save": func() (string, error) { return NewDialog().PickSaveFile("Save", "body.json") },
	} {
		path, err := pick()
		if err == nil || path != "" {
			t.Fatalf("%s: = %q, %v; want an error", name, path, err)
		}
	}
}

func TestOpenUrlWithoutRunningApplicationFails(t *testing.T) {
	if err := NewBrowser().OpenURL("https://example.test"); err == nil {
		t.Fatal("OpenURL() without a running application must fail")
	}
}

func TestLoopbackListensOnTheGivenAddress(t *testing.T) {
	listener, err := NewLoopback().Listen("127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	defer func() { _ = listener.Close() }()
	if _, busy := NewLoopback().Listen(listener.Addr().String()); busy == nil {
		t.Fatal("a second Listen() on a busy address must fail")
	}
}

func TestTimersRunAfterTheDelayUnlessStopped(t *testing.T) {
	timers := NewTimers()
	ran := make(chan struct{})
	timers.AfterFunc(time.Millisecond, func() { close(ran) })
	select {
	case <-ran:
	case <-time.After(5 * time.Second):
		t.Fatal("the timer did not run")
	}
	stop := timers.AfterFunc(time.Hour, func() { t.Error("a stopped timer ran") })
	if !stop() || stop() {
		t.Fatal("stop reports whether the timer was still pending")
	}
}
