package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type AIApprovalRequest struct {
	Environment AIApprovalEnvironment `json:"environment"`
	SessionId   string                `json:"session_id"`
	Client      string                `json:"client"`
	Method      string                `json:"method" validate:"required_without=Summary"`
	Url         string                `json:"url"`
	Summary     string                `json:"summary"`
}

func NewAIApprovalRequest(sessionId, client, method, url string, environment AIApprovalEnvironment) AIApprovalRequest {
	return AIApprovalRequest{Environment: environment, SessionId: sessionId, Client: client, Method: method, Url: url}
}

func NewAIRunApprovalRequest(sessionId, client, summary string, environment AIApprovalEnvironment) AIApprovalRequest {
	return AIApprovalRequest{Environment: environment, SessionId: sessionId, Client: client, Summary: summary}
}

func (a *AIApprovalRequest) ToEntity() entity.AIApproval {
	return entity.AIApproval{Client: a.Client, Method: a.Method, RawURL: a.Url, Summary: a.Summary, Environment: a.Environment.ToEntity()}
}

func (a AIApprovalRequest) LogValue() any {
	return map[string]any{"session_id": a.SessionId, "client": a.Client, "method": a.Method, "summary": a.Summary, "environment_id": a.Environment.Id}
}

func (a *AIApprovalRequest) Session() string {
	return a.SessionId
}
