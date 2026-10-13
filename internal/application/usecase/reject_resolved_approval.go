package usecase

import "github.com/brienze1/xenon/pkg/errs"

func RejectResolvedApproval(id string) error {
	return errs.ConflictError("this send was already approved, denied or timed out", "MCP-05409",
		errs.ErrorDetails{Attribute: "id", Messages: []string{id}})
}
