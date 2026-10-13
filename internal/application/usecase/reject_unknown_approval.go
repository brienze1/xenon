package usecase

import "github.com/brienze1/xenon/pkg/errs"

func RejectUnknownApproval(id string) error {
	return errs.BadRequestError("no pending send with this id", "MCP-07400",
		errs.ErrorDetails{Attribute: "id", Messages: []string{id}})
}
