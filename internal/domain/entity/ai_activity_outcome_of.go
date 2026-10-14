package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/aiactivityoutcome"
)

// AIActivityOutcomeOf maps a failed call's error code to its outcome; no code is ok.
func AIActivityOutcomeOf(code string) enums.AIActivityOutcome {
	switch code {
	case "":
		return aiactivityoutcome.Ok
	case AISendDeniedCode:
		return aiactivityoutcome.Denied
	case AISendTimedOutCode:
		return aiactivityoutcome.TimedOut
	default:
		return aiactivityoutcome.Error
	}
}
