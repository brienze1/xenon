package adapter

import "github.com/brienze1/xenon/internal/application/usecase"

type ConsolePersistence interface {
	usecase.AppendConsoleEntries
	usecase.FindConsoleEntries
	usecase.ClearConsole
}
