package digestqop

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Auth    enums.DigestQop = "auth"
	AuthInt enums.DigestQop = "auth-int"
	None    enums.DigestQop = ""
	Unknown enums.DigestQop = "unknown"
)

func GetDigestQop(s string) enums.DigestQop {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "auth":
		return Auth
	case "auth-int":
		return AuthInt
	case "":
		return None
	default:
		return Unknown
	}
}
