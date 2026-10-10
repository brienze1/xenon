package entity

import (
	"time"
	"unicode/utf8"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/consoleentrykind"
)

const consoleSummaryLimit = 200

type ConsoleEntry struct {
	CreatedAt time.Time
	Request   *EffectiveRequest
	Response  *Response
	Script    *ScriptRef
	Kind      enums.ConsoleEntryKind
	Level     enums.ConsoleLevel
	ID        string
	SendID    string
	Summary   string
	Arguments []string
	Masked    bool
}

func NewConsoleLogEntry(id, sendId string, level enums.ConsoleLevel, script ScriptRef, text string, arguments []string, createdAt time.Time) ConsoleEntry {
	return ConsoleEntry{
		ID:        id,
		SendID:    sendId,
		Kind:      consoleentrykind.Log,
		Level:     level,
		Script:    &script,
		Summary:   ConsoleSummary(text),
		Arguments: append([]string{}, arguments...),
		CreatedAt: createdAt,
	}
}

func ConsoleSummary(text string) string {
	if utf8.RuneCountInString(text) <= consoleSummaryLimit {
		return text
	}
	return string([]rune(text)[:consoleSummaryLimit])
}

func (c ConsoleEntry) IsRequest() bool {
	return c.Kind == consoleentrykind.Request
}

func (c ConsoleEntry) IsLog() bool {
	return c.Kind == consoleentrykind.Log || c.Kind == consoleentrykind.Info
}

func (c ConsoleEntry) IsError() bool {
	return c.Kind == consoleentrykind.Error || c.Kind == consoleentrykind.TestFailure
}
