package runresultfilter

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	All     enums.RunResultFilter = "all"
	Passed  enums.RunResultFilter = "passed"
	Failed  enums.RunResultFilter = "failed"
	Skipped enums.RunResultFilter = "skipped"
	Unknown enums.RunResultFilter = "unknown"
)

func GetRunResultFilter(s string) enums.RunResultFilter {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "all":
		return All
	case "passed":
		return Passed
	case "failed":
		return Failed
	case "skipped":
		return Skipped
	default:
		return Unknown
	}
}
