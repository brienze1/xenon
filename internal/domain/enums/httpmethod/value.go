package httpmethod

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Get     enums.HttpMethod = "GET"
	Post    enums.HttpMethod = "POST"
	Put     enums.HttpMethod = "PUT"
	Patch   enums.HttpMethod = "PATCH"
	Delete  enums.HttpMethod = "DELETE"
	Head    enums.HttpMethod = "HEAD"
	Options enums.HttpMethod = "OPTIONS"
	Unknown enums.HttpMethod = "UNKNOWN"
)

func GetHttpMethod(s string) enums.HttpMethod {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "GET":
		return Get
	case "POST":
		return Post
	case "PUT":
		return Put
	case "PATCH":
		return Patch
	case "DELETE":
		return Delete
	case "HEAD":
		return Head
	case "OPTIONS":
		return Options
	default:
		return Unknown
	}
}
