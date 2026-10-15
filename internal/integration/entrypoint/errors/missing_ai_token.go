package errors

import "github.com/brienze1/xenon/pkg/errs"

func RejectMissingAIToken() error {
	return errs.NotFoundError("there is no OAuth 2.0 token for this auth", "MCP-13404")
}
