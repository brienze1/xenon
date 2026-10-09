package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type BodyRaw struct {
	Language string `json:"language" validate:"omitempty,oneof=json xml text html javascript"`
	Content  string `json:"content"`
}

func NewBodyRaw(raw entity.BodyRaw) BodyRaw {
	return BodyRaw{Language: raw.Language.String(), Content: raw.Content}
}

func (b *BodyRaw) ToEntity() entity.BodyRaw {
	return entity.NewBodyRaw(b.Language, b.Content)
}
