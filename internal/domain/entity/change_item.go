package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/changetype"
)

type ChangeItem struct {
	ItemID   string
	ItemName string
	ItemKind enums.ItemKind
	Change   enums.ChangeType
	Paths    []string
}

func NewChangeItem(ref ItemRef, change PathChange) ChangeItem {
	return ChangeItem{ItemID: ref.ID, ItemName: ref.Name, ItemKind: ref.Kind, Change: change.Change, Paths: []string{change.Path}}
}

func (c ChangeItem) Combined(change PathChange) ChangeItem {
	c.Paths = append(append([]string{}, c.Paths...), change.Path)
	if c.Change == change.Change {
		return c
	}
	if c.isMove(change) {
		c.Change = changetype.Renamed
		return c
	}
	c.Change = changetype.Modified
	return c
}

func (c ChangeItem) isMove(change PathChange) bool {
	return (c.Change == changetype.Deleted && change.Change == changetype.Added) ||
		(c.Change == changetype.Added && change.Change == changetype.Deleted)
}

func (c ChangeItem) IsAnyOf(ids []string) bool {
	for _, id := range ids {
		if c.ItemID == id {
			return true
		}
	}
	return false
}
