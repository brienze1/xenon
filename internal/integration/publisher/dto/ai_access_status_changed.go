package dto

import "github.com/brienze1/xenon/internal/domain/entity"

const AIAccessStatusChangedEvent = "ai:access-status-changed"

type AIAccessStatusChanged struct {
	State   string `json:"state"`
	Address string `json:"address,omitempty"`
	Reason  string `json:"reason,omitempty"`
	Code    string `json:"code,omitempty"`
	Clients int    `json:"clients"`
}

func NewAIAccessStatusChanged(status entity.AIAccessStatus) AIAccessStatusChanged {
	return AIAccessStatusChanged{State: status.State.String(), Address: status.Address(), Reason: status.Reason, Code: status.Code, Clients: status.Clients}
}
