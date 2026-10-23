package model

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/envcolor"
)

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

func (h *HistoryEnvironment) ToEntity() *entity.HistoryEnvironment {
	if h == nil {
		return nil
	}
	return &entity.HistoryEnvironment{ID: h.Id, Name: h.Name, Color: envcolor.GetEnvColor(h.Color)}
}
