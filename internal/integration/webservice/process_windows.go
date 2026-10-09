//go:build windows

package webservice

import "os"

func isProcessRunning(processId int) bool {
	process, err := os.FindProcess(processId)
	if err != nil {
		return false
	}
	_ = process.Release()
	return true
}
