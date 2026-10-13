package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type BodyBinary struct {
	Src      string `json:"src"`
	SrcState string `json:"src_state" validate:"omitempty,oneof=ok repick"`
}

func NewBodyBinary(binary entity.FileRef) BodyBinary {
	return BodyBinary{Src: binary.Src, SrcState: newFileState(binary)}
}

func (b *BodyBinary) ToEntity() entity.FileRef {
	return fileRefToEntity(b.Src, b.SrcState)
}
