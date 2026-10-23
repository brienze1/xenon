package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type MCPRunStarted struct {
	RunId   string `json:"run_id"`
	Status  string `json:"status"`
	Planned int    `json:"planned"`
}

func NewMCPRunStarted(run entity.Run) MCPRunStarted {
	return MCPRunStarted{RunId: run.ID, Status: run.Status.String(), Planned: run.Progress.Planned}
}
