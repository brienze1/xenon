package dto

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
)

const mcpGitPathSeparator = ", "

type MCPGitChange struct {
	Id     string `json:"id"`
	Path   string `json:"path,omitempty"`
	Name   string `json:"name"`
	Kind   string `json:"kind"`
	Change string `json:"change"`
}

func NewMCPGitChange(change entity.ChangeItem) MCPGitChange {
	return MCPGitChange{
		Id:     change.ItemID,
		Path:   strings.Join(change.Paths, mcpGitPathSeparator),
		Name:   change.ItemName,
		Kind:   change.ItemKind.String(),
		Change: change.Change.String(),
	}
}

func NewMCPGitChangeList(changes []entity.ChangeItem) MCPPage[MCPGitChange] {
	lines := make([]MCPGitChange, 0, len(changes))
	for _, change := range changes {
		lines = append(lines, NewMCPGitChange(change))
	}
	return NewMCPList(lines)
}
