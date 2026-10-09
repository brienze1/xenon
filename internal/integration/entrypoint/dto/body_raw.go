package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/rawlanguage"
)

type BodyRaw struct {
	Language string `json:"language"`
	Content  string `json:"content"`
}

func NewBodyRaw(raw entity.BodyRaw) BodyRaw {
	return BodyRaw{Language: raw.Language.String(), Content: raw.Content}
}

func (b *BodyRaw) ToEntity() entity.BodyRaw {
	raw := entity.BodyRaw{Language: rawlanguage.Json, Content: b.Content}
	if b.Language != "" {
		raw.Language = rawlanguage.GetRawLanguage(b.Language)
	}
	return raw
}
