package entity

import "github.com/brienze1/xenon/internal/domain/enums"

// AIRunStart is the outcome of a run an AI client asked for: the started run, the refused approval or an unknown environment.
type AIRunStart struct {
	Target             AIActivityTarget
	Run                *Run
	Refused            *AIApproval
	Changes            []enums.AIChangeKind
	UnknownEnvironment bool
}

func NewAIRunStart(target AIActivityTarget) AIRunStart {
	return AIRunStart{Target: target}
}

func (a AIRunStart) WithUnknownEnvironment() AIRunStart {
	a.UnknownEnvironment = true
	return a
}

// WithOutcome records the started run, or the approval the user refused instead of it.
func (a AIRunStart) WithOutcome(run *Run, refused *AIApproval, changes []enums.AIChangeKind) AIRunStart {
	a.Run = run
	a.Refused = refused
	a.Changes = changes
	return a
}

func (a AIRunStart) HasUnknownEnvironment() bool {
	return a.UnknownEnvironment
}

func (a AIRunStart) IsRefused() bool {
	return a.Refused != nil
}
