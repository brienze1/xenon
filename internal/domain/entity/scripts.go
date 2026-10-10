package entity

import (
	"slices"
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/scriptphase"
)

type Scripts struct {
	PreRequest   string
	PostResponse string
	Disabled     []enums.ScriptPhase
}

func (s Scripts) Text(phase enums.ScriptPhase) string {
	if phase == scriptphase.PreRequest {
		return s.PreRequest
	}
	return s.PostResponse
}

func (s Scripts) HasText(phase enums.ScriptPhase) bool {
	return strings.TrimSpace(s.Text(phase)) != ""
}

func (s Scripts) IsDisabled(phase enums.ScriptPhase) bool {
	return slices.Contains(s.Disabled, phase)
}

func (s Scripts) Phases() []enums.ScriptPhase {
	phases := make([]enums.ScriptPhase, 0, 2)
	for _, phase := range []enums.ScriptPhase{scriptphase.PreRequest, scriptphase.PostResponse} {
		if s.HasText(phase) {
			phases = append(phases, phase)
		}
	}
	return phases
}

func (s Scripts) IsEmpty() bool {
	return len(s.Phases()) == 0
}
