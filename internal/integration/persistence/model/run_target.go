package model

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
)

type RunTarget struct {
	ItemId string `json:"item_id"`
	Kind   string `json:"kind"`
	Name   string `json:"name"`
}

func NewRunTarget(target entity.RunTarget) RunTarget {
	return RunTarget{ItemId: target.ItemID, Kind: target.Kind.String(), Name: target.Name}
}

func (r *RunTarget) ToEntity(localId string) entity.RunTarget {
	return entity.RunTarget{LocalID: localId, ItemID: r.ItemId, Kind: itemkind.GetItemKind(r.Kind), Name: r.Name}
}
