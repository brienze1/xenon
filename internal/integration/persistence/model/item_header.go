package model

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
)

type ItemHeader struct {
	Kind string `yaml:"kind"`
	Id   string `yaml:"id"`
	Name string `yaml:"name"`
}

func (i *ItemHeader) IsIdentified() bool {
	return i.Id != "" && itemkind.GetItemKind(i.Kind) != itemkind.Unknown
}

func (i *ItemHeader) ToEntity() entity.ItemRef {
	return entity.ItemRef{ID: i.Id, Name: i.Name, Kind: itemkind.GetItemKind(i.Kind)}
}
