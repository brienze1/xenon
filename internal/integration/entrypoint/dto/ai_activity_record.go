package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/aiactivityoutcome"
)

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

func NewAIActivityRecord(localId, client, tool, action string, target AIActivityTarget, outcome, code, args string, durationMs int64) AIActivityRecord {
	return AIActivityRecord{
		Target:     target,
		LocalId:    localId,
		Client:     client,
		Tool:       tool,
		Action:     action,
		Outcome:    outcome,
		Code:       code,
		Args:       args,
		DurationMs: durationMs,
	}
}

func (a *AIActivityRecord) ToEntity() entity.AIActivity {
	return entity.AIActivity{
		Tool:       a.Tool,
		Action:     a.Action,
		Target:     a.Target.ToEntity(),
		Outcome:    aiactivityoutcome.GetAIActivityOutcome(a.Outcome),
		Code:       a.Code,
		Args:       a.Args,
		DurationMs: a.DurationMs,
	}
}

func (a AIActivityRecord) LogValue() any {
	return map[string]any{"local_id": a.LocalId, "client": a.Client, "tool": a.Tool, "action": a.Action, "outcome": a.Outcome, "code": a.Code, "duration_ms": a.DurationMs}
}

func (a *AIActivityRecord) WorkspaceId() string {
	return a.LocalId
}

func (a *AIActivityRecord) ClientName() string {
	return a.Client
}
