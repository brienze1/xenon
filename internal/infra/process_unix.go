//go:build !windows

package application

import "syscall"

func processRunning(pid int) bool {
	return pid > 0 && syscall.Kill(pid, 0) == nil
}
