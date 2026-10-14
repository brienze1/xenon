package dto

import (
	"slices"
	"strings"
)

func hasSuffixFold(name string, suffixes ...string) bool {
	lowered := strings.ToLower(name)
	return slices.ContainsFunc(suffixes, func(suffix string) bool { return strings.HasSuffix(lowered, suffix) })
}

func compareFold(a, b string) int {
	if byFold := strings.Compare(strings.ToLower(a), strings.ToLower(b)); byFold != 0 {
		return byFold
	}
	return strings.Compare(a, b)
}
