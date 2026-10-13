package entity

import (
	"errors"

	"github.com/brienze1/xenon/pkg/errs"
)

const (
	nextStepAttribute = "next_step"
	attributeKey      = "attribute"
	messagesKey       = "messages"
)

type SyncError struct {
	Code     string
	Message  string
	NextStep string
}

func NewSyncError(err error) SyncError {
	var custom errs.BaseError
	if !errors.As(err, &custom) {
		return SyncError{Message: err.Error()}
	}
	failure := SyncError{Code: custom.Code(), Message: custom.Description()}
	details := custom.ErrorDetails()
	if messages, ok := details[messagesKey].([]string); ok && details[attributeKey] == nextStepAttribute && len(messages) > 0 {
		failure.NextStep = messages[0]
	}
	return failure
}
