package model

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/scriptownerkind"
	"github.com/brienze1/xenon/internal/domain/enums/scriptphase"
)

type RunScriptRef struct {
	Phase     string `json:"phase"`
	OwnerKind string `json:"owner_kind"`
	ItemId    string `json:"item_id"`
	OwnerName string `json:"owner_name"`
}

func NewRunScriptRef(script entity.ScriptRef) RunScriptRef {
	return RunScriptRef{Phase: script.Phase.String(), OwnerKind: script.OwnerKind.String(), ItemId: script.ItemID, OwnerName: script.OwnerName}
}

func (r *RunScriptRef) ToEntity() entity.ScriptRef {
	return entity.ScriptRef{Phase: scriptphase.GetScriptPhase(r.Phase), OwnerKind: scriptownerkind.GetScriptOwnerKind(r.OwnerKind), ItemID: r.ItemId, OwnerName: r.OwnerName}
}
