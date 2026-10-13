package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type EnvironmentList struct {
	ActiveId     string               `json:"active_id"`
	Environments []EnvironmentSummary `json:"environments"`
}

func NewEnvironmentList(list entity.EnvironmentList) EnvironmentList {
	result := EnvironmentList{ActiveId: list.ActiveID, Environments: make([]EnvironmentSummary, 0, len(list.Environments))}
	for _, environment := range list.Environments {
		result.Environments = append(result.Environments, NewEnvironmentSummary(environment))
	}
	return result
}
