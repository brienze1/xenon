package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type MCPGitConflict struct {
	Id         string                `json:"id"`
	Name       string                `json:"name"`
	Kind       string                `json:"kind"`
	Conflict   string                `json:"conflict"`
	Path       string                `json:"path,omitempty"`
	Resolution string                `json:"resolution,omitempty"`
	Choices    []string              `json:"choices"`
	Fields     []MCPGitConflictField `json:"fields,omitempty"`
}

func NewMCPGitConflict(conflict entity.ConflictItem) MCPGitConflict {
	choices := make([]string, 0, len(conflict.Choices()))
	for _, choice := range conflict.Choices() {
		choices = append(choices, choice.String())
	}
	fields := make([]MCPGitConflictField, 0, len(conflict.Fields))
	for _, field := range conflict.Fields {
		fields = append(fields, NewMCPGitConflictField(field))
	}
	return MCPGitConflict{
		Id:         conflict.ItemID,
		Name:       conflict.ItemName,
		Kind:       conflict.ItemKind.String(),
		Conflict:   conflict.Kind.String(),
		Path:       conflict.Path,
		Resolution: conflict.ChosenResolution().String(),
		Choices:    choices,
		Fields:     fields,
	}
}
