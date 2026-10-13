package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type AIActivityRecord struct {
	Target     AIActivityTarget `json:"target"`
	LocalId    string           `json:"local_id"`
	Client     string           `json:"client"`
	Tool       string           `json:"tool" validate:"required,not_blank"`
	Action     string           `json:"action"`
	Outcome    string           `json:"outcome" validate:"required,oneof=ok error denied timed_out"`
	Code       string           `json:"code"`
	Args       string           `json:"args"`
	DurationMs int64            `json:"duration_ms" validate:"min=0"`
}

type AIActivity struct {
	Target     *AIActivityTarget `json:"target,omitempty"`
	Tool       string            `json:"tool"`
	Action     string            `json:"action,omitempty"`
	Outcome    string            `json:"outcome"`
	Code       string            `json:"code,omitempty"`
	Args       string            `json:"args,omitempty"`
	DurationMs int64             `json:"duration_ms"`
}

type AIActivityTarget struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Path string `json:"path,omitempty"`
}

func NewAIActivity(activity entity.AIActivity, withArgs bool) AIActivity {
	result := AIActivity{Tool: activity.Tool, Action: activity.Action, Outcome: activity.Outcome, Code: activity.Code, DurationMs: activity.DurationMs}
	if withArgs {
		result.Args = activity.Args
	}
	if activity.Target != (entity.AIActivityTarget{}) {
		result.Target = &AIActivityTarget{Id: activity.Target.ID, Name: activity.Target.Name, Path: activity.Target.Path}
	}
	return result
}

func (a *AIActivityRecord) ToEntity() entity.AIActivity {
	return entity.AIActivity{
		Tool:       a.Tool,
		Action:     a.Action,
		Target:     entity.AIActivityTarget{ID: a.Target.Id, Name: a.Target.Name, Path: a.Target.Path},
		Outcome:    a.Outcome,
		Code:       a.Code,
		Args:       a.Args,
		DurationMs: a.DurationMs,
	}
}

func (a AIActivityRecord) LogValue() any {
	return map[string]any{"local_id": a.LocalId, "client": a.Client, "tool": a.Tool, "action": a.Action, "outcome": a.Outcome, "code": a.Code, "duration_ms": a.DurationMs}
}
