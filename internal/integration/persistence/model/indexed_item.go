package model

import "github.com/brienze1/xenon/internal/domain/enums"

type IndexedItem struct {
	Path string
	Kind enums.ItemKind
}
