package model

import "github.com/brienze1/xenon/internal/domain/entity"

type TreeIndex struct {
	Items        map[string]IndexedItem
	Reidentified map[string]string
	Tree         entity.Tree
}
