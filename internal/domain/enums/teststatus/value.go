package teststatus

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Passed  enums.TestStatus = "passed"
	Failed  enums.TestStatus = "failed"
	Skipped enums.TestStatus = "skipped"
	Unknown enums.TestStatus = "unknown"
)

func GetTestStatus(s string) enums.TestStatus {
	switch strings.ToLower(strings.TrimSpace(s)) {
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
