package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ConsoleEntry struct {
	Script    *ConsoleScript   `json:"script,omitempty"`
	Request   *ConsoleRequest  `json:"request,omitempty"`
	Response  *ConsoleResponse `json:"response,omitempty"`
	Line      *int             `json:"line,omitempty"`
	Column    *int             `json:"column,omitempty"`
	Id        string           `json:"id"`
	SendId    string           `json:"send_id"`
	Kind      string           `json:"kind"`
	Level     string           `json:"level"`
	Summary   string           `json:"summary"`
	CreatedAt string           `json:"created_at"`
	Arguments []string         `json:"arguments"`
	Masked    bool             `json:"masked"`
}

func NewConsoleEntry(entry entity.ConsoleEntry) ConsoleEntry {
	result := ConsoleEntry{
		Id:        entry.ID,
		SendId:    entry.SendID,
		Kind:      entry.Kind.String(),
		Level:     entry.Level.String(),
		Summary:   entry.Summary,
		CreatedAt: entry.CreatedAt.UTC().Format("2006-01-02T15:04:05.000Z"),
		Arguments: append([]string{}, entry.Arguments...),
		Masked:    entry.Masked,
	}
	if entry.Script != nil {
		script := NewConsoleScript(*entry.Script)
		result.Script = &script
	}
	if entry.Line > 0 {
		line, column := entry.Line, entry.Column
		result.Line, result.Column = &line, &column
	}
	if entry.Request != nil {
		request := NewConsoleRequest(*entry.Request, entry.Nested)
		result.Request = &request
	}
	if entry.Response != nil {
		response := NewConsoleResponse(*entry.Response)
		result.Response = &response
	}
	return result
}
