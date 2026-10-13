package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type AIApproval struct {
	Environment AIApprovalEnvironment `json:"environment"`
	Id          string                `json:"id"`
	Client      string                `json:"client"`
	Method      string                `json:"method"`
	Url         string                `json:"url"`
	CreatedAt   string                `json:"created_at"`
}

type AIApprovalEnvironment struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type AIApprovalRequest struct {
	Environment AIApprovalEnvironment `json:"environment"`
	SessionId   string                `json:"session_id"`
	Client      string                `json:"client"`
	Method      string                `json:"method" validate:"required,not_blank"`
	Url         string                `json:"url"`
}

type AIApprovalOutcome struct {
	Id      string `json:"id"`
	Outcome string `json:"outcome"`
}

type AIApprovalDenial struct {
	SessionId string `json:"session_id"`
}

func NewAIApproval(approval entity.AIApproval) AIApproval {
	return AIApproval{
		Id:          approval.ID,
		Client:      approval.Client,
		Method:      approval.Method,
		Url:         approval.RawURL,
		CreatedAt:   approval.CreatedAt.UTC().Format("2006-01-02T15:04:05.000Z"),
		Environment: AIApprovalEnvironment{Id: approval.Environment.ID, Name: approval.Environment.Name, Color: approval.Environment.Color},
	}
}

func NewAIApprovals(approvals []entity.AIApproval) []AIApproval {
	rows := make([]AIApproval, 0, len(approvals))
	for _, approval := range approvals {
		rows = append(rows, NewAIApproval(approval))
	}
	return rows
}

func NewAIApprovalOutcome(approval entity.AIApproval) AIApprovalOutcome {
	return AIApprovalOutcome{Id: approval.ID, Outcome: approval.Outcome.String()}
}

func (a *AIApprovalRequest) ToEntity() entity.AIApproval {
	environment := entity.AIApprovalEnvironment{ID: a.Environment.Id, Name: a.Environment.Name, Color: a.Environment.Color}
	return entity.AIApproval{Client: a.Client, Method: a.Method, RawURL: a.Url, Environment: environment}
}

func (a AIApprovalDenial) IsAll() bool {
	return a.SessionId == ""
}

func (a AIApprovalRequest) LogValue() any {
	return map[string]any{"session_id": a.SessionId, "client": a.Client, "method": a.Method, "environment_id": a.Environment.Id}
}

func (a AIApprovalOutcome) LogValue() any {
	return map[string]any{"id": a.Id, "outcome": a.Outcome}
}
