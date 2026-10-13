package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ConsoleScript struct {
	OwnerKind string `json:"owner_kind"`
	OwnerId   string `json:"owner_id"`
	OwnerName string `json:"owner_name"`
	Phase     string `json:"phase"`
	Label     string `json:"label"`
}

func NewConsoleScript(script entity.ScriptRef) ConsoleScript {
	return ConsoleScript{OwnerKind: script.OwnerKind.String(), OwnerId: script.ItemID, OwnerName: script.OwnerName, Phase: script.Phase.String(), Label: script.Label()}
}
