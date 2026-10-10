package clashresolution

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Replace    enums.ClashResolution = "replace"
	KeepBoth   enums.ClashResolution = "keep_both"
	Skip       enums.ClashResolution = "skip"
	Unresolved enums.ClashResolution = ""
)

func GetClashResolution(s string) enums.ClashResolution {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "replace":
		return Replace
	case "keep_both":
		return KeepBoth
	case "skip":
		return Skip
	default:
		return Unresolved
	}
}
