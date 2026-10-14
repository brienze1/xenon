package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type AIApprovalEnvironment struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

func NewAIApprovalEnvironment(environment entity.AIApprovalEnvironment) AIApprovalEnvironment {
	return AIApprovalEnvironment{Id: environment.ID, Name: environment.Name, Color: environment.Color}
}

func (a *AIApprovalEnvironment) ToEntity() entity.AIApprovalEnvironment {
	return entity.NewAIApprovalEnvironment(a.Id, a.Name, a.Color)
}
