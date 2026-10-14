package application

import (
	"os"

	"github.com/brienze1/xenon/internal/infra/mcpbridge"
	"github.com/brienze1/xenon/pkg/logger"
)

const mcpCommand = "mcp"

func Run() {
	if len(os.Args) > 1 && os.Args[1] == mcpCommand {
		logger.SetOutput(os.Stderr)
		os.Exit(mcpbridge.Run(os.Stdin, os.Stdout, os.Stderr))
	}
	Start()
}
