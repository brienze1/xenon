package application

import (
	"os"
	"strconv"
	"time"

	"github.com/brienze1/xenon/internal/infra/mcpbridge"
	"github.com/brienze1/xenon/internal/infra/properties"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/process"
)

const (
	mcpCommand          = "mcp"
	restartWaitLimit    = 15 * time.Second
	restartWaitInterval = 100 * time.Millisecond
)

func Run() {
	if len(os.Args) > 1 && os.Args[1] == mcpCommand {
		logger.SetOutput(os.Stderr)
		os.Exit(mcpbridge.Run(os.Stdin, os.Stdout, os.Stderr))
	}
	waitForPreviousProcess(properties.Properties().Application.RestartAfterPid, process.Running, time.Sleep, restartWaitLimit)
	Start()
}

func waitForPreviousProcess(pid string, running func(int) bool, sleep func(time.Duration), limit time.Duration) {
	previous, err := strconv.Atoi(pid)
	if err != nil || previous <= 0 {
		return
	}
	for waited := time.Duration(0); waited < limit && running(previous); waited += restartWaitInterval {
		sleep(restartWaitInterval)
	}
}
