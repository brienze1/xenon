package application

import (
	"os"
	"strconv"
	"time"

	"github.com/brienze1/xenon/internal/infra/mcpbridge"
	"github.com/brienze1/xenon/pkg/logger"
)

const (
	mcpCommand              = "mcp"
	restartAfterPidVariable = "XENON_RESTART_AFTER_PID"
	restartWaitLimit        = 15 * time.Second
	restartWaitInterval     = 100 * time.Millisecond
)

func Run() {
	if len(os.Args) > 1 && os.Args[1] == mcpCommand {
		logger.SetOutput(os.Stderr)
		os.Exit(mcpbridge.Run(os.Stdin, os.Stdout, os.Stderr))
	}
	waitForPreviousProcess(os.Getenv(restartAfterPidVariable), processRunning, time.Sleep, restartWaitLimit)
	Start()
}

// waitForPreviousProcess lets an updated Xenon start only once the Xenon it replaces has exited and released the single-instance lock.
func waitForPreviousProcess(pid string, running func(int) bool, sleep func(time.Duration), limit time.Duration) {
	previous, err := strconv.Atoi(pid)
	if err != nil || previous <= 0 {
		return
	}
	for waited := time.Duration(0); waited < limit && running(previous); waited += restartWaitInterval {
		sleep(restartWaitInterval)
	}
}
