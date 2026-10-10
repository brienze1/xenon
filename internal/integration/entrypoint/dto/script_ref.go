package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ScriptRef struct {
	OwnerKind string `json:"owner_kind"`
	OwnerId   string `json:"owner_id"`
	OwnerName string `json:"owner_name"`
	Phase     string `json:"phase"`
}

func NewScriptRef(script entity.ScriptRef) ScriptRef {
	return ScriptRef{OwnerKind: script.OwnerKind.String(), OwnerId: script.ItemID, OwnerName: script.OwnerName, Phase: script.Phase.String()}
}
