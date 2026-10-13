package model

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/yamlnode"
)

type ItemCopy struct {
	Id      string `yaml:"id"`
	Name    string `yaml:"name"`
	Renamed bool   `yaml:"-"`
}

func NewItemCopy(copied entity.ItemCopy) ItemCopy {
	return ItemCopy{Id: copied.ID, Name: copied.Name, Renamed: copied.Renames()}
}

func (i *ItemCopy) Fields() []yamlnode.Field {
	fields := []yamlnode.Field{yamlnode.String("id", i.Id)}
	if i.Renamed {
		fields = append(fields, yamlnode.String("name", i.Name))
	}
	return fields
}
