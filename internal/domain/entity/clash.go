package entity

import "github.com/brienze1/xenon/internal/domain/enums"

type Clash struct {
	Kind         enums.ClashKind
	ItemKey      string
	ExistingID   string
	ExistingName string
	Name         string
}

func (c Clash) Description() string {
	return c.Kind.String() + " " + c.ExistingName
}
