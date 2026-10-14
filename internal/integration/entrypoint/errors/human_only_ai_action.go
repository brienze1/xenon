package errors

import "github.com/brienze1/xenon/pkg/errs"

func RejectHumanOnlyAIAction(what string) error {
	return errs.ForbiddenError(what+" is human-only — do it in Xenon.", "MCP-02403")
}
