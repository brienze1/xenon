package dto

import (
	"slices"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type TagGroup struct {
	Name string   `yaml:"name"`
	Tags []string `yaml:"tags"`
}

func NewTagGroup(group entity.ApiTagGroup) TagGroup {
	return TagGroup{Name: group.Name, Tags: slices.Clone(group.Tags)}
}

func (t *TagGroup) ToEntity() entity.ApiTagGroup {
	return entity.ApiTagGroup{Name: t.Name, Tags: append([]string{}, t.Tags...)}
}

func (t TagGroup) MarshalYAML() (any, error) {
	return mappingOfFields(always("name", t.Name), always("tags", t.Tags))
}
