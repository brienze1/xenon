package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
)

type TreeExample struct {
	Id   string `json:"id"`
	Kind string `json:"kind"`
	Name string `json:"name"`
	Code int    `json:"code"`
}

func NewTreeExample(example entity.TreeExample) TreeExample {
	return TreeExample{Id: example.ID, Kind: itemkind.Example.String(), Name: example.Name, Code: example.Code}
}
