package entity

import (
	"slices"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/aitooleffect"
)

const aiToolActionArgument = "action"

type AITool struct {
	Name            string
	Description     string
	Selector        string
	Effect          enums.AIToolEffect
	Actions         []string
	Idempotent      bool
	OpenWorld       bool
	WorkspaceScoped bool
}

func (a AITool) ArgumentName() string {
	if a.Selector != "" {
		return a.Selector
	}
	return aiToolActionArgument
}

func (a AITool) HasActions() bool {
	return len(a.Actions) > 0
}

// DefaultAction is the only action of a single-action tool, else empty.
func (a AITool) DefaultAction() string {
	if len(a.Actions) == 1 {
		return a.Actions[0]
	}
	return ""
}

func (a AITool) RequiresAction() bool {
	return a.HasActions() && a.DefaultAction() == ""
}

func (a AITool) ActionOrDefault(action string) string {
	if action == "" {
		return a.DefaultAction()
	}
	return action
}

func (a AITool) IsWorkspaceScoped() bool {
	return a.WorkspaceScoped
}

func (a AITool) MissesWorkspace(localId string) bool {
	return a.IsWorkspaceScoped() && !HasActiveWorkspace(localId)
}

func (a AITool) Allows(action string) bool {
	return slices.Contains(a.Actions, action)
}

func (a AITool) IsReadOnly() bool {
	return a.Effect == aitooleffect.ReadOnly
}

func (a AITool) IsDestructive() bool {
	return a.Effect == aitooleffect.Destructive
}
