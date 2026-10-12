package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ApiTag struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func NewApiTag(tag entity.ApiTag) ApiTag {
	return ApiTag{Name: tag.Name, Description: tag.Description}
}
