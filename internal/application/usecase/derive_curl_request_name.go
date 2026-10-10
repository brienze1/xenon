package usecase

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
)

const schemeSeparator = "://"

func DeriveCurlRequestName(request entity.Request) string {
	address := request.URL
	if _, rest, found := strings.Cut(address, schemeSeparator); found {
		address = rest
	}
	address, _, _ = strings.Cut(address, "#")
	address, _, _ = strings.Cut(address, "?")
	host, path, _ := strings.Cut(address, "/")
	if path = strings.TrimRight(path, "/"); path != "" {
		return request.MethodOrDefault() + " /" + path
	}
	return request.MethodOrDefault() + " " + host
}
