package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type HistoryEnvironment struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

func NewHistoryEnvironment(environment *entity.HistoryEnvironment) *HistoryEnvironment {
	if environment == nil {
		return nil
	}
	return &HistoryEnvironment{Id: environment.ID, Name: environment.Name, Color: environment.Color.String()}
}
