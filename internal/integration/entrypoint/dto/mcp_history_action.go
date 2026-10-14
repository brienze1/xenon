package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type MCPHistoryAction struct {
	Target  *AIActivityTarget `json:"target,omitempty"`
	Tool    string            `json:"tool"`
	Action  string            `json:"action,omitempty"`
	Outcome string            `json:"outcome"`
	Code    string            `json:"code,omitempty"`
	Args    string            `json:"args,omitempty"`
	Ms      int64             `json:"ms"`
}

func NewMCPHistoryAction(aiActivity entity.AIActivity) MCPHistoryAction {
	activity := NewAIActivity(aiActivity, true)
	return MCPHistoryAction{
		Tool:    activity.Tool,
		Action:  activity.Action,
		Target:  activity.Target,
		Outcome: activity.Outcome,
		Code:    activity.Code,
		Ms:      activity.DurationMs,
		Args:    activity.Args,
	}
}
