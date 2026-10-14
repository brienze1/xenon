package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
)

type MCPToolResult struct {
	Payload  any              `json:"payload"`
	Target   AIActivityTarget `json:"target"`
	Changes  []AIChange       `json:"changes"`
	Recorded bool             `json:"recorded"`
}

func NewMCPToolResult(payload any) MCPToolResult {
	return MCPToolResult{Payload: payload}
}

func NewMCPTargetResult(target entity.AIActivityTarget) MCPToolResult {
	return MCPToolResult{Target: NewAIActivityTarget(target)}
}

func (m MCPToolResult) WithTarget(target entity.AIActivityTarget) MCPToolResult {
	m.Target = NewAIActivityTarget(target)
	return m
}

func (m MCPToolResult) WithChanges(kinds ...enums.AIChangeKind) MCPToolResult {
	changes := append(make([]AIChange, 0, len(m.Changes)+len(kinds)), m.Changes...)
	for _, kind := range kinds {
		changes = append(changes, NewAIChange(kind))
	}
	m.Changes = changes
	return m
}

func (m MCPToolResult) WithChange(change AIChange) MCPToolResult {
	m.Changes = append(append(make([]AIChange, 0, len(m.Changes)+1), m.Changes...), change)
	return m
}

func (m MCPToolResult) Prepended(kind enums.AIChangeKind) MCPToolResult {
	m.Changes = append([]AIChange{NewAIChange(kind)}, m.Changes...)
	return m
}

func (m MCPToolResult) AsRecorded() MCPToolResult {
	m.Recorded = true
	return m
}

func (m MCPToolResult) NeedsRecording() bool {
	return !m.Recorded
}

func (m MCPToolResult) NotifiesChanges(failed bool, localId string) bool {
	return !failed && len(m.Changes) > 0 && entity.HasActiveWorkspace(localId)
}

func (m MCPToolResult) LogValue() any {
	return map[string]any{"target_id": m.Target.Id, "changes": m.Changes, "recorded": m.Recorded}
}
