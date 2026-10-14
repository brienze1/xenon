package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type MCPHistoryLine struct {
	Status  *int   `json:"status,omitempty"`
	Id      string `json:"id"`
	At      string `json:"at"`
	Kind    string `json:"kind"`
	Source  string `json:"source"`
	Client  string `json:"client,omitempty"`
	Method  string `json:"method,omitempty"`
	Url     string `json:"url,omitempty"`
	Tool    string `json:"tool,omitempty"`
	Action  string `json:"action,omitempty"`
	Target  string `json:"target,omitempty"`
	Outcome string `json:"outcome,omitempty"`
	Ms      int64  `json:"ms,omitempty"`
}

func NewMCPHistoryLine(historyEntry entity.HistoryEntry) MCPHistoryLine {
	entry := NewHistoryEntry(historyEntry)
	line := MCPHistoryLine{Id: entry.Id, At: entry.SentAt, Kind: entry.Kind, Source: entry.Source, Client: entry.Client, Ms: entry.TotalMs}
	if entry.Activity == nil {
		line.Method, line.Url, line.Status, line.Outcome = entry.Method, entry.Url, entry.StatusCode, entry.Outcome
		return line
	}
	line.Tool, line.Action, line.Outcome = entry.Activity.Tool, entry.Activity.Action, entry.Activity.Outcome
	if entry.Activity.Target != nil {
		line.Target = entry.Activity.Target.Name
	}
	return line
}
