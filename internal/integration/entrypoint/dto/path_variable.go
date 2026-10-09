package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type PathVariable struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Description string `json:"description"`
}

func NewPathVariable(variable entity.PathVariable) PathVariable {
	return PathVariable{Key: variable.Key, Value: variable.Value, Description: variable.Description}
}

func (p *PathVariable) ToEntity() entity.PathVariable {
	return entity.PathVariable{Key: p.Key, Value: p.Value, Description: p.Description}
}
