package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type RunTarget struct {
	ItemId string `json:"item_id"`
	Kind   string `json:"kind"`
	Name   string `json:"name"`
}

func NewRunTarget(target entity.RunTarget) RunTarget {
	return RunTarget{ItemId: target.ItemID, Kind: target.Kind.String(), Name: target.Name}
}
