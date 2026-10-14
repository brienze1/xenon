package dto

import "github.com/brienze1/xenon/pkg/errs"

type MCPToolError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewMCPToolError(failure errs.BaseError) MCPToolError {
	return MCPToolError{Code: failure.Code(), Message: failure.Description()}
}
