package application

import (
	"testing"
	"time"
)

func TestWaitForPreviousProcessWaitsUntilItExits(t *testing.T) {
	checks := 0
	slept := time.Duration(0)
	waitForPreviousProcess("42", func(pid int) bool {
		checks++
		return pid == 42 && checks < 3
	}, func(d time.Duration) { slept += d }, time.Second)
	if checks != 3 || slept != 2*restartWaitInterval {
		t.Fatalf("checks = %d, slept = %s", checks, slept)
	}
}

func TestWaitForPreviousProcessGivesUpAtTheLimit(t *testing.T) {
	slept := time.Duration(0)
	waitForPreviousProcess("42", func(int) bool { return true }, func(d time.Duration) { slept += d }, 5*restartWaitInterval)
	if slept != 5*restartWaitInterval {
		t.Fatalf("slept = %s", slept)
	}
}

func TestWaitForPreviousProcessIgnoresAMissingPid(t *testing.T) {
	for _, pid := range []string{"", "abc", "0", "-3"} {
		waitForPreviousProcess(pid, func(int) bool {
			t.Fatalf("pid %q must not be checked", pid)
			return false
		}, func(time.Duration) {}, time.Second)
	}
}
