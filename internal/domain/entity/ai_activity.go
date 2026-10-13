package entity

import (
	"strings"
	"unicode/utf8"
)

const (
	MaxAIActivityArgsBytes = 2048
	aiActivityArgsEllipsis = "…"
	AIActivityOk           = "ok"
	AIActivityError        = "error"
	AIActivityDenied       = "denied"
	AIActivityTimedOut     = "timed_out"
)

type AIActivityTarget struct {
	ID   string
	Name string
	Path string
}

type AIActivity struct {
	Tool       string
	Action     string
	Target     AIActivityTarget
	Outcome    string
	Code       string
	Args       string
	DurationMs int64
}

func (a AIActivity) WithArgs(args string) AIActivity {
	a.Args = CutAIActivityArgs(args)
	return a
}

// CutAIActivityArgs keeps at most 2,048 bytes on a rune boundary and marks the cut with an ellipsis.
func CutAIActivityArgs(args string) string {
	if len(args) <= MaxAIActivityArgsBytes {
		return args
	}
	cut := MaxAIActivityArgsBytes
	for cut > 0 && !utf8.RuneStart(args[cut]) {
		cut--
	}
	return args[:cut] + aiActivityArgsEllipsis
}

func (a AIActivity) Mentions(needle string) bool {
	lowered := strings.ToLower(needle)
	for _, field := range []string{a.Tool, a.Action, a.Target.Name} {
		if strings.Contains(strings.ToLower(field), lowered) {
			return true
		}
	}
	return false
}

func (a AIActivity) LogValue() any {
	return map[string]any{"tool": a.Tool, "action": a.Action, "targetId": a.Target.ID, "outcome": a.Outcome, "code": a.Code, "durationMs": a.DurationMs}
}
