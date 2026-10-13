package oauth2grant

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	AuthorizationCode         enums.OAuth2Grant = "authorization_code"
	AuthorizationCodeWithPkce enums.OAuth2Grant = "authorization_code_with_pkce"
	Implicit                  enums.OAuth2Grant = "implicit"
	PasswordCredentials       enums.OAuth2Grant = "password_credentials"
	ClientCredentials         enums.OAuth2Grant = "client_credentials"
	Unknown                   enums.OAuth2Grant = "unknown"
)

func GetOAuth2Grant(s string) enums.OAuth2Grant {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "authorization_code":
		return AuthorizationCode
	case "authorization_code_with_pkce":
		return AuthorizationCodeWithPkce
	case "implicit":
		return Implicit
	case "password_credentials":
		return PasswordCredentials
	case "client_credentials":
		return ClientCredentials
	default:
		return Unknown
	}
}
