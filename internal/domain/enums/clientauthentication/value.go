package clientauthentication

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Header  enums.ClientAuthentication = "header"
	Body    enums.ClientAuthentication = "body"
	Unknown enums.ClientAuthentication = "unknown"
)

func GetClientAuthentication(s string) enums.ClientAuthentication {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "header":
		return Header
	case "body":
		return Body
	default:
		return Unknown
	}
}
