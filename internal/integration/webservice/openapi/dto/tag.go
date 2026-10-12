package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type Tag struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
}

func NewTag(tag entity.ApiTag) Tag {
	return Tag{Name: tag.Name, Description: tag.Description}
}

func (t *Tag) ToEntity() entity.ApiTag {
	return entity.ApiTag{Name: t.Name, Description: t.Description}
}

func (t Tag) MarshalYAML() (any, error) {
	return mappingOfFields(always("name", t.Name), text("description", t.Description))
}
