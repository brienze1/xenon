package usecase

import (
	"strconv"
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
	"github.com/brienze1/xenon/pkg/fileutils"
)

const requestFileExtension = ".yaml"

var reservedRequestFileNames = []string{"collection.yaml", "folder.yaml"}

func DeriveFileName(name string, kind enums.ItemKind, siblings []string) string {
	base := fileutils.SafeName(name)
	extension := fileExtension(kind)
	candidate := base + extension
	for suffix := 2; isTaken(candidate, siblings); suffix++ {
		candidate = base + "-" + strconv.Itoa(suffix) + extension
	}
	return candidate
}

func fileExtension(kind enums.ItemKind) string {
	if kind == itemkind.Request {
		return requestFileExtension
	}
	return ""
}

func isTaken(candidate string, siblings []string) bool {
	return containsFold(siblings, candidate) || containsFold(reservedRequestFileNames, candidate)
}

func containsFold(names []string, candidate string) bool {
	for _, name := range names {
		if strings.EqualFold(name, candidate) {
			return true
		}
	}
	return false
}
