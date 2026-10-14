package model

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/aiactivityoutcome"
)

type AIActivity struct {
	Target     *AIActivityTarget `json:"target,omitempty"`
	Tool       string            `json:"tool"`
	Action     string            `json:"action,omitempty"`
	Outcome    string            `json:"outcome"`
	Code       string            `json:"code,omitempty"`
	Args       string            `json:"args,omitempty"`
	DurationMs int64             `json:"duration_ms"`
}

func NewAIActivity(activity entity.AIActivity) AIActivity {
	model := AIActivity{
		Tool:       activity.Tool,
		Action:     activity.Action,
		Outcome:    activity.Outcome.String(),
		Code:       activity.Code,
		Args:       activity.Args,
		DurationMs: activity.DurationMs,
	}
	if !activity.Target.IsEmpty() {
		target := NewAIActivityTarget(activity.Target)
		model.Target = &target
	}
	return model
}

func (a *AIActivity) ToEntity() entity.AIActivity {
	activity := entity.AIActivity{
		Tool:       a.Tool,
		Action:     a.Action,
		Outcome:    aiactivityoutcome.GetAIActivityOutcome(a.Outcome),
		Code:       a.Code,
		Args:       a.Args,
		DurationMs: a.DurationMs,
	}
	if a.Target != nil {
		activity.Target = a.Target.ToEntity()
	}
	return activity
}
