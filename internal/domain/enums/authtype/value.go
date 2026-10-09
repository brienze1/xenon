package authtype

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Inherit enums.AuthType = "inherit"
	None    enums.AuthType = "none"
	Bearer  enums.AuthType = "bearer"
	Basic   enums.AuthType = "basic"
	ApiKey  enums.AuthType = "apikey"
	OAuth2  enums.AuthType = "oauth2"
	AwsV4   enums.AuthType = "awsv4"
	Digest  enums.AuthType = "digest"
	Unknown enums.AuthType = "unknown"
)

func GetAuthType(s string) enums.AuthType {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "inherit":
		return Inherit
	case "none":
		return None
	case "bearer":
		return Bearer
	case "basic":
		return Basic
	case "apikey":
		return ApiKey
	case "oauth2":
		return OAuth2
	case "awsv4":
		return AwsV4
	case "digest":
		return Digest
	default:
		return Unknown
	}
}
