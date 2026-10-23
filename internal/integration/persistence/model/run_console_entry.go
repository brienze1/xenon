package model

import (
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/consoleentrykind"
	"github.com/brienze1/xenon/internal/domain/enums/consolelevel"
)

type RunConsoleEntry struct {
	CreatedAt time.Time     `json:"created_at"`
	Script    *RunScriptRef `json:"script,omitempty"`
	Id        string        `json:"id"`
	SendId    string        `json:"send_id"`
	Kind      string        `json:"kind"`
	Level     string        `json:"level"`
	Summary   string        `json:"summary"`
	Arguments []string      `json:"arguments"`
	Line      int           `json:"line"`
	Column    int           `json:"column"`
	Masked    bool          `json:"masked"`
	Nested    bool          `json:"nested"`
}

func NewRunConsoleEntry(entry entity.ConsoleEntry) RunConsoleEntry {
	model := RunConsoleEntry{
		CreatedAt: entry.CreatedAt.UTC(),
		Id:        entry.ID,
		SendId:    entry.SendID,
		Kind:      entry.Kind.String(),
		Level:     entry.Level.String(),
		Summary:   entry.Summary,
		Arguments: append([]string{}, entry.Arguments...),
		Line:      entry.Line,
		Column:    entry.Column,
		Masked:    entry.Masked,
		Nested:    entry.Nested,
	}
	if entry.Script != nil {
		script := NewRunScriptRef(*entry.Script)
		model.Script = &script
	}
	return model
}

func (r *RunConsoleEntry) ToEntity() entity.ConsoleEntry {
	entry := entity.ConsoleEntry{
		CreatedAt: r.CreatedAt,
		ID:        r.Id,
		SendID:    r.SendId,
		Kind:      consoleentrykind.GetConsoleEntryKind(r.Kind),
		Level:     consolelevel.GetConsoleLevel(r.Level),
		Summary:   r.Summary,
		Arguments: append([]string{}, r.Arguments...),
		Line:      r.Line,
		Column:    r.Column,
		Masked:    r.Masked,
		Nested:    r.Nested,
	}
	if r.Script != nil {
		script := r.Script.ToEntity()
		entry.Script = &script
	}
	return entry
}
