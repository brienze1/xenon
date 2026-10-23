package entity

import "github.com/brienze1/xenon/internal/domain/enums"

type RunTarget struct {
	Kind    enums.ItemKind
	LocalID string
	ItemID  string
	Name    string
}

func NewRunTarget(localId string, container TreeNode) RunTarget {
	return RunTarget{LocalID: localId, ItemID: container.ID, Kind: container.Kind, Name: container.Name}
}
