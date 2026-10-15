package process

import (
	"os"
	"testing"
)

func TestRunningKnowsTheCurrentProcess(t *testing.T) {
	if Running(-1) || Running(0) {
		t.Fatal("no process has a negative or zero id")
	}
	if !Running(os.Getpid()) {
		t.Fatal("the current process is running")
	}
}
