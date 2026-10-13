package senderrorkind

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	None              enums.SendErrorKind = ""
	Dns               enums.SendErrorKind = "dns"
	ConnectionRefused enums.SendErrorKind = "connection_refused"
	Timeout           enums.SendErrorKind = "timeout"
	Tls               enums.SendErrorKind = "tls"
	Cancelled         enums.SendErrorKind = "cancelled"
	MissingFile       enums.SendErrorKind = "missing_file"
	InvalidUrl        enums.SendErrorKind = "invalid_url"
	Other             enums.SendErrorKind = "other"
	Unknown           enums.SendErrorKind = "unknown"
)

func GetSendErrorKind(s string) enums.SendErrorKind {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "":
		return None
	case "dns":
		return Dns
	case "connection_refused":
		return ConnectionRefused
	case "timeout":
		return Timeout
	case "tls":
		return Tls
	case "cancelled":
		return Cancelled
	case "missing_file":
		return MissingFile
	case "invalid_url":
		return InvalidUrl
	case "other":
		return Other
	default:
		return Unknown
	}
}
