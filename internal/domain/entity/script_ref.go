package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/scriptphase"
)

type ScriptRef struct {
	Phase     enums.ScriptPhase
	OwnerKind enums.ScriptOwnerKind
	ItemID    string
	OwnerName string
	Path      string
	Text      string
}

func (s ScriptRef) IsPreRequest() bool {
	return s.Phase == scriptphase.PreRequest
}

func (s ScriptRef) Label() string {
	if s.IsPreRequest() {
		return s.OwnerName + " · Pre-request"
	}
	return s.OwnerName + " · Post-response"
}
