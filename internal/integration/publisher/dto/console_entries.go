package dto

import "github.com/brienze1/xenon/internal/domain/entity"

const (
	ConsoleEntriesEvent        = "console:entries"
	microsecondsPerMillisecond = 1000
)

type ConsoleEntries struct {
	Entries []ConsoleEntry `json:"entries"`
	Cleared bool           `json:"cleared"`
}

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

type ConsoleScript struct {
	OwnerKind string `json:"owner_kind"`
	OwnerId   string `json:"owner_id"`
	OwnerName string `json:"owner_name"`
	Phase     string `json:"phase"`
	Label     string `json:"label"`
}

type ConsoleRequest struct {
	Method string `json:"method"`
	Url    string `json:"url"`
	Nested bool   `json:"nested"`
}

type ConsoleResponse struct {
	Error      string  `json:"error,omitempty"`
	StatusText string  `json:"status_text"`
	TimeMs     float64 `json:"time_ms"`
	SizeBytes  int64   `json:"size_bytes"`
	StatusCode int     `json:"status_code"`
}

func NewConsoleEntries(entries []entity.ConsoleEntry) ConsoleEntries {
	result := ConsoleEntries{Entries: make([]ConsoleEntry, 0, len(entries))}
	for _, entry := range entries {
		result.Entries = append(result.Entries, newConsoleEntry(entry))
	}
	return result
}

func NewConsoleCleared() ConsoleEntries {
	return ConsoleEntries{Entries: []ConsoleEntry{}, Cleared: true}
}

func newConsoleEntry(entry entity.ConsoleEntry) ConsoleEntry {
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
		result.Script = &ConsoleScript{OwnerKind: entry.Script.OwnerKind.String(), OwnerId: entry.Script.ItemID, OwnerName: entry.Script.OwnerName, Phase: entry.Script.Phase.String(), Label: entry.Script.Label()}
	}
	if entry.Line > 0 {
		line, column := entry.Line, entry.Column
		result.Line, result.Column = &line, &column
	}
	if entry.Request != nil {
		result.Request = &ConsoleRequest{Method: entry.Request.Method, Url: entry.Request.URL, Nested: entry.Nested}
	}
	if entry.Response != nil {
		result.Response = newConsoleResponse(*entry.Response)
	}
	return result
}

func newConsoleResponse(response entity.Response) *ConsoleResponse {
	result := &ConsoleResponse{StatusCode: response.StatusCode, StatusText: response.StatusText, SizeBytes: response.SizeBytes, TimeMs: float64(response.Timings.Total.Microseconds()) / microsecondsPerMillisecond}
	if response.Error != nil {
		result.Error = response.Error.Message
	}
	return result
}
