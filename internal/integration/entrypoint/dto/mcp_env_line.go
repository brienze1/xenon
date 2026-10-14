package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type MCPEnvLine struct {
	Id                string `json:"id"`
	Name              string `json:"name"`
	Color             string `json:"color"`
	Active            bool   `json:"active,omitempty"`
	ConfirmBeforeSend bool   `json:"confirm_before_send,omitempty"`
}

func NewMCPEnvLine(environment entity.Environment, active bool) MCPEnvLine {
	return MCPEnvLine{Id: environment.ID, Name: environment.Name, Color: environment.Color.String(), Active: active, ConfirmBeforeSend: environment.ConfirmBeforeSend}
}

func NewMCPEnvList(list entity.EnvironmentList) MCPPage[MCPEnvLine] {
	lines := make([]MCPEnvLine, 0, len(list.Environments))
	for _, environment := range list.Environments {
		lines = append(lines, NewMCPEnvLine(environment, list.IsActive(environment.ID)))
	}
	return NewMCPList(lines)
}
