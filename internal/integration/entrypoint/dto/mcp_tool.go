package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/aitooleffect"
)

type MCPTool struct {
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	Argument        string   `json:"argument"`
	Actions         []string `json:"actions"`
	RequiresAction  bool     `json:"requires_action"`
	ReadOnly        bool     `json:"read_only"`
	Destructive     bool     `json:"destructive"`
	Idempotent      bool     `json:"idempotent"`
	OpenWorld       bool     `json:"open_world"`
	WorkspaceScoped bool     `json:"workspace_scoped"`
}

func NewMCPTool(tool entity.AITool) MCPTool {
	return MCPTool{
		Name:            tool.Name,
		Description:     tool.Description,
		Argument:        tool.ArgumentName(),
		Actions:         tool.Actions,
		RequiresAction:  tool.RequiresAction(),
		ReadOnly:        tool.IsReadOnly(),
		Destructive:     tool.IsDestructive(),
		Idempotent:      tool.Idempotent,
		OpenWorld:       tool.OpenWorld,
		WorkspaceScoped: tool.WorkspaceScoped,
	}
}

func (m *MCPTool) ToEntity() entity.AITool {
	effect := aitooleffect.Write
	if m.ReadOnly {
		effect = aitooleffect.ReadOnly
	}
	if m.Destructive {
		effect = aitooleffect.Destructive
	}
	return entity.AITool{
		Name:            m.Name,
		Description:     m.Description,
		Selector:        m.Argument,
		Actions:         m.Actions,
		Effect:          effect,
		Idempotent:      m.Idempotent,
		OpenWorld:       m.OpenWorld,
		WorkspaceScoped: m.WorkspaceScoped,
	}
}

func (m *MCPTool) InputSchema() map[string]any {
	schema := map[string]any{"type": "object"}
	if len(m.Actions) == 0 {
		return schema
	}
	schema["properties"] = map[string]any{m.Argument: map[string]any{"type": "string", "enum": m.Actions}}
	if m.RequiresAction {
		schema["required"] = []string{m.Argument}
	}
	return schema
}
