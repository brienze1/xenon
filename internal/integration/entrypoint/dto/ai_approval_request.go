package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type AIApprovalRequest struct {
	Environment AIApprovalEnvironment `json:"environment"`
	SessionId   string                `json:"session_id"`
	Client      string                `json:"client"`
	Method      string                `json:"method" validate:"required,not_blank"`
	Url         string                `json:"url"`
}

func NewAIApprovalRequest(sessionId, client, method, url string, environment AIApprovalEnvironment) AIApprovalRequest {
	return AIApprovalRequest{Environment: environment, SessionId: sessionId, Client: client, Method: method, Url: url}
}

func (a *AIApprovalRequest) ToEntity() entity.AIApproval {
	return entity.AIApproval{Client: a.Client, Method: a.Method, RawURL: a.Url, Environment: a.Environment.ToEntity()}
}

func (a AIApprovalRequest) LogValue() any {
	return map[string]any{"session_id": a.SessionId, "client": a.Client, "method": a.Method, "environment_id": a.Environment.Id}
}

func (a *AIApprovalRequest) Session() string {
	return a.SessionId
}
