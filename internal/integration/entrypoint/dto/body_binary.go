package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type BodyBinary struct {
	Src string `json:"src"`
}

func NewBodyBinary(binary entity.FileRef) BodyBinary {
	return BodyBinary{Src: binary.Src}
}

func (b *BodyBinary) ToEntity() entity.FileRef {
	return entity.FileRef{Src: b.Src}
}
