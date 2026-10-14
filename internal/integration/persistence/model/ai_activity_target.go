package model

import "github.com/brienze1/xenon/internal/domain/entity"

type AIActivityTarget struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Path string `json:"path,omitempty"`
}

func NewAIActivityTarget(target entity.AIActivityTarget) AIActivityTarget {
	return AIActivityTarget{Id: target.ID, Name: target.Name, Path: target.Path}
}

func (a *AIActivityTarget) ToEntity() entity.AIActivityTarget {
	return entity.AIActivityTarget{ID: a.Id, Name: a.Name, Path: a.Path}
}
