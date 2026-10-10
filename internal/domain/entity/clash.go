package entity

import (
	"strconv"
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const importedSuffix = " (imported)"

type Clash struct {
	Kind         enums.ClashKind
	ItemKey      string
	ExistingID   string
	ExistingName string
	Name         string
}

func (c Clash) Description() string {
	return c.Kind.String() + " " + c.ExistingName
}

func ImportedName(name string, taken []string) string {
	candidate := name + importedSuffix
	for suffix := 2; isNameTaken(candidate, taken); suffix++ {
		candidate = name + " (imported " + strconv.Itoa(suffix) + ")"
	}
	return candidate
}

func isNameTaken(candidate string, taken []string) bool {
	for _, name := range taken {
		if strings.EqualFold(strings.TrimSpace(name), candidate) {
			return true
		}
	}
	return false
}
