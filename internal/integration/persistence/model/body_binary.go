package model

import "github.com/brienze1/xenon/internal/domain/entity"

type BodyBinary struct {
	Src      string `yaml:"src" json:"src"`
	SrcState string `yaml:"srcState" json:"src_state,omitempty"`
}

func NewBodyBinary(binary entity.FileRef) BodyBinary {
	return BodyBinary{Src: binary.Src, SrcState: fileRefState(binary)}
}

func (b *BodyBinary) ToEntity() entity.FileRef {
	return fileRefToEntity(b.Src, b.SrcState)
}
