package model

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

type AIActivityTarget struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Path string `json:"path,omitempty"`
}

func NewAIActivity(activity entity.AIActivity) AIActivity {
	model := AIActivity{
		Tool:       activity.Tool,
		Action:     activity.Action,
		Outcome:    activity.Outcome,
		Code:       activity.Code,
		Args:       activity.Args,
		DurationMs: activity.DurationMs,
	}
	if activity.Target != (entity.AIActivityTarget{}) {
		model.Target = &AIActivityTarget{Id: activity.Target.ID, Name: activity.Target.Name, Path: activity.Target.Path}
	}
	return model
}

func (a *AIActivity) ToEntity() entity.AIActivity {
	activity := entity.AIActivity{
		Tool:       a.Tool,
		Action:     a.Action,
		Outcome:    a.Outcome,
		Code:       a.Code,
		Args:       a.Args,
		DurationMs: a.DurationMs,
	}
	if a.Target != nil {
		activity.Target = entity.AIActivityTarget{ID: a.Target.Id, Name: a.Target.Name, Path: a.Target.Path}
	}
	return activity
}
