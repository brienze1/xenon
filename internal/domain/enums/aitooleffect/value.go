package aitooleffect

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	ReadOnly    enums.AIToolEffect = "read_only"
	Write       enums.AIToolEffect = "write"
	Destructive enums.AIToolEffect = "destructive"
	Unknown     enums.AIToolEffect = "unknown"
)

func GetAIToolEffect(s string) enums.AIToolEffect {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "read_only":
		return ReadOnly
	case "write":
		return Write
	case "destructive":
		return Destructive
	default:
		return Unknown
	}
}
