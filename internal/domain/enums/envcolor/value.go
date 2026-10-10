package envcolor

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	None    enums.EnvColor = "none"
	Green   enums.EnvColor = "green"
	Cyan    enums.EnvColor = "cyan"
	Magenta enums.EnvColor = "magenta"
	Yellow  enums.EnvColor = "yellow"
	Red     enums.EnvColor = "red"
	Unknown enums.EnvColor = "unknown"
)

func GetEnvColor(s string) enums.EnvColor {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "none":
		return None
	case "green":
		return Green
	case "cyan":
		return Cyan
	case "magenta":
		return Magenta
	case "yellow":
		return Yellow
	case "red":
		return Red
	default:
		return Unknown
	}
}
