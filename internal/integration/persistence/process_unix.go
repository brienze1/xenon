//go:build !windows

package persistence

import (
	"errors"
	"os"
	"syscall"
)

func isProcessRunning(processId int) bool {
	process, _ := os.FindProcess(processId)
	err := process.Signal(syscall.Signal(0))
	return err == nil || errors.Is(err, syscall.EPERM)
}
