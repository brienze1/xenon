//go:build windows

package persistence

import (
	"errors"

	"golang.org/x/sys/windows"
)

const stillActive = 259

func isProcessRunning(processId int) bool {
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(processId))
	if err != nil {
		return errors.Is(err, windows.ERROR_ACCESS_DENIED)
	}
	defer func() { _ = windows.CloseHandle(handle) }()
	var exitCode uint32
	if err := windows.GetExitCodeProcess(handle, &exitCode); err != nil {
		return true
	}
	return exitCode == stillActive
}
