package entity

import "github.com/brienze1/xenon/internal/domain/enums"

type ScriptRef struct {
	Phase  enums.ScriptPhase
	ItemID string
	Path   string
	Text   string
}
