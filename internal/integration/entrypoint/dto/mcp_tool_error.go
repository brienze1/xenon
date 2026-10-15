package dto

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
)

const (
	invalidRequestPrefix = "invalid request: "
	detailFieldJoiner    = "; "
)

type MCPToolError struct {
	Code    string               `json:"code"`
	Message string               `json:"message"`
	Details []MCPToolErrorDetail `json:"details,omitempty"`
}

// NewMCPToolError names each failing field by the argument the client sent for the action; a validation failure's
// message is rebuilt from those names.
func NewMCPToolError(failure errs.BaseError, tool MCPTool, action string) MCPToolError {
	details := failure.ToDto("").Error.ErrorDetails
	toolError := MCPToolError{Code: failure.Code(), Message: failure.Description()}
	problems := make([]string, 0, len(details))
	for _, detail := range details {
		mapped := NewMCPToolErrorDetail(tool.ArgumentOf(action, detail.Attribute), detail.Messages)
		toolError.Details = append(toolError.Details, mapped)
		problems = append(problems, mapped.Field+" "+mapped.Problem)
	}
	if len(problems) > 0 && strings.HasPrefix(toolError.Message, invalidRequestPrefix) {
		toolError.Message = invalidRequestPrefix + strings.Join(problems, detailFieldJoiner)
	}
	return toolError
}

func NewMCPPanicToolError(failure errs.BaseError) MCPToolError {
	return NewMCPToolError(failure, NewMCPTool(entity.AITool{}), "")
}
