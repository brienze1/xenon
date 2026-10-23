package loadprofilekind

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Fixed   enums.LoadProfileKind = "fixed"
	RampUp  enums.LoadProfileKind = "ramp_up"
	Spike   enums.LoadProfileKind = "spike"
	Peak    enums.LoadProfileKind = "peak"
	Unknown enums.LoadProfileKind = "unknown"
)

func GetLoadProfileKind(s string) enums.LoadProfileKind {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "fixed":
		return Fixed
	case "ramp_up":
		return RampUp
	case "spike":
		return Spike
	case "peak":
		return Peak
	default:
		return Unknown
	}
}
