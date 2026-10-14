package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type MCPConsoleLine struct {
	Id      string `json:"id"`
	At      string `json:"at"`
	Type    string `json:"type"`
	Summary string `json:"summary"`
}

func NewMCPConsoleLine(entry entity.ConsoleEntry) MCPConsoleLine {
	converted := NewConsoleEntry(entry)
	return MCPConsoleLine{Id: converted.Id, At: converted.CreatedAt, Type: converted.Kind, Summary: converted.Summary}
}
