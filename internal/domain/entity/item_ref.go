package entity

import "github.com/brienze1/xenon/internal/domain/enums"

type ItemRef struct {
	ID   string
	Name string
	Kind enums.ItemKind
}
