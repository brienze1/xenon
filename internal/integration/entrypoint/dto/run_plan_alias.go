package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type RunPlanAlias struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func NewRunPlanAlias(alias entity.RequestAlias) RunPlanAlias {
	return RunPlanAlias{Id: alias.ID, Name: alias.Name}
}
