package adapter

import "github.com/brienze1/xenon/internal/application/usecase"

type ConsoleEntriesPublisher interface {
	usecase.PublishConsoleEntries
}
