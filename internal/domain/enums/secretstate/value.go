package secretstate

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Set     enums.SecretState = "set"
	Missing enums.SecretState = "missing"
	Locked  enums.SecretState = "locked"
	Unknown enums.SecretState = "unknown"
)

func GetSecretState(s string) enums.SecretState {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "set":
		return Set
	case "missing":
		return Missing
	case "locked":
		return Locked
	default:
		return Unknown
	}
}
