//go:build !windows

package process

import "syscall"

// Running reports whether a process with the given id is alive.
func Running(pid int) bool {
	return pid > 0 && syscall.Kill(pid, 0) == nil
}
