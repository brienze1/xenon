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

// Equals compares the identity of two targets: the same container of the same workspace.
func (r RunTarget) Equals(other RunTarget) bool {
	return r.LocalID == other.LocalID && r.ItemID == other.ItemID
}

func (r RunTarget) InWorkspace(localId string) bool {
	return r.LocalID == localId
}
