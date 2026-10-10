package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ConsoleEntry struct {
	Script    *ScriptRef `json:"script,omitempty"`
	Id        string     `json:"id"`
	SendId    string     `json:"send_id"`
	Kind      string     `json:"kind"`
	Level     string     `json:"level"`
	Summary   string     `json:"summary"`
	CreatedAt string     `json:"created_at"`
	Arguments []string   `json:"arguments"`
	Masked    bool       `json:"masked"`
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
		script := NewScriptRef(*entry.Script)
		result.Script = &script
	}
	return result
}

func NewConsoleEntries(entries []entity.ConsoleEntry) []ConsoleEntry {
	result := make([]ConsoleEntry, 0, len(entries))
	for _, entry := range entries {
		result = append(result, NewConsoleEntry(entry))
	}
	return result
}

func (c ConsoleEntry) LogValue() any {
	return map[string]any{"id": c.Id, "send_id": c.SendId, "kind": c.Kind}
}
