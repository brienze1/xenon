//go:build windows

package process

import "golang.org/x/sys/windows"

const stillActive = 259

// Running reports whether a process with the given id is alive.
func Running(pid int) bool {
	if pid <= 0 {
		return false
	}
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return false
	}
	defer func() { _ = windows.CloseHandle(handle) }()
	var code uint32
	return windows.GetExitCodeProcess(handle, &code) == nil && code == stillActive
}
