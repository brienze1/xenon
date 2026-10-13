package bodymode

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	None       enums.BodyMode = "none"
	FormData   enums.BodyMode = "formdata"
	UrlEncoded enums.BodyMode = "urlencoded"
	Raw        enums.BodyMode = "raw"
	Binary     enums.BodyMode = "binary"
	GraphQl    enums.BodyMode = "graphql"
	Unknown    enums.BodyMode = "unknown"
)

func GetBodyMode(s string) enums.BodyMode {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "none":
		return None
	case "formdata":
		return FormData
	case "urlencoded":
		return UrlEncoded
	case "raw":
		return Raw
	case "binary":
		return Binary
	case "graphql":
		return GraphQl
	default:
		return Unknown
	}
}
