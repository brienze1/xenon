package bodyencoding

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Text    enums.BodyEncoding = "text"
	Base64  enums.BodyEncoding = "base64"
	Unknown enums.BodyEncoding = "unknown"
)

func GetBodyEncoding(s string) enums.BodyEncoding {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "text":
		return Text
	case "base64":
		return Base64
	default:
		return Unknown
	}
}
