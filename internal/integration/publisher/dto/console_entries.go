package dto

import "github.com/brienze1/xenon/internal/domain/entity"

const ConsoleEntriesEvent = "console:entries"

type ConsoleEntries struct {
	Entries []ConsoleEntry `json:"entries"`
	Cleared bool           `json:"cleared"`
}

func NewConsoleEntries(entries []entity.ConsoleEntry) ConsoleEntries {
	result := ConsoleEntries{Entries: make([]ConsoleEntry, 0, len(entries))}
	for _, entry := range entries {
		result.Entries = append(result.Entries, NewConsoleEntry(entry))
	}
	return result
}

func NewConsoleCleared() ConsoleEntries {
	return ConsoleEntries{Entries: []ConsoleEntry{}, Cleared: true}
}
