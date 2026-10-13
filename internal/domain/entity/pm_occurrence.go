package entity

import "github.com/brienze1/xenon/internal/domain/enums"

type PmOccurrence struct {
	Phase   enums.ScriptPhase
	ItemKey string
	ItemID  string
	LocalID string
	Path    string
	Line    int
}
