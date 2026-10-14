package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type AIActivity struct {
	Target     *AIActivityTarget `json:"target,omitempty"`
	Tool       string            `json:"tool"`
	Action     string            `json:"action,omitempty"`
	Outcome    string            `json:"outcome"`
	Code       string            `json:"code,omitempty"`
	Args       string            `json:"args,omitempty"`
	DurationMs int64             `json:"duration_ms"`
}

func NewAIActivity(activity entity.AIActivity, withArgs bool) AIActivity {
	result := AIActivity{Tool: activity.Tool, Action: activity.Action, Outcome: activity.Outcome.String(), Code: activity.Code, DurationMs: activity.DurationMs}
	if withArgs {
		result.Args = activity.Args
	}
	if !activity.Target.IsEmpty() {
		target := NewAIActivityTarget(activity.Target)
		result.Target = &target
	}
	return result
}
