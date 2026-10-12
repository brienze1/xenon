package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ApiTagGroup struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

func NewApiTagGroup(group entity.ApiTagGroup) ApiTagGroup {
	return ApiTagGroup{Name: group.Name, Tags: append([]string{}, group.Tags...)}
}
