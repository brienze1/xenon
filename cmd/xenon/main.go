package main

import (
	"os"

	application "github.com/brienze1/xenon/internal/infra"
	"github.com/brienze1/xenon/internal/infra/mcpbridge"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "mcp" {
		os.Exit(mcpbridge.Run(os.Stdin, os.Stdout, os.Stderr))
	}
	application.Start()
}
