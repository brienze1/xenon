package approvaloutcome

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Pending  enums.ApprovalOutcome = "pending"
	Approved enums.ApprovalOutcome = "approved"
	Denied   enums.ApprovalOutcome = "denied"
	TimedOut enums.ApprovalOutcome = "timed_out"
	Unknown  enums.ApprovalOutcome = "unknown"
)

func GetApprovalOutcome(s string) enums.ApprovalOutcome {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "pending", "":
		return Pending
	case "approved":
		return Approved
	case "denied":
		return Denied
	case "timed_out":
		return TimedOut
	default:
		return Unknown
	}
}
