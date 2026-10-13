package signinstatus

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Pending enums.SignInStatus = "pending"
	Done    enums.SignInStatus = "done"
	Expired enums.SignInStatus = "expired"
	Denied  enums.SignInStatus = "denied"
	Unknown enums.SignInStatus = "unknown"
)

func GetSignInStatus(s string) enums.SignInStatus {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "pending":
		return Pending
	case "done":
		return Done
	case "expired":
		return Expired
	case "denied":
		return Denied
	default:
		return Unknown
	}
}
