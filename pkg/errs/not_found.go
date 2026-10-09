package errs

import (
	"errors"
	"net/http"
)

type notFoundError struct {
	message     string
	description string
	code        string
	statusCode  int
}

func NotFoundError(message string, code string) BaseError {
	return &notFoundError{
		statusCode:  http.StatusNotFound,
		message:     "not found",
		description: message,
		code:        code,
	}
}

func IsNotFoundError(err error) bool {
	var target *notFoundError
	return errors.As(err, &target)
}

func (n *notFoundError) ToDto(transactionId string) Error {
	return toDto(transactionId, n.description, n.code, nil)
}

func (n *notFoundError) SetStatusCode(statusCode int, description string) {
	n.statusCode = statusCode
	n.description = description
}

func (n *notFoundError) Is(target error) bool {
	var other *notFoundError
	return errors.As(target, &other)
}

func (n *notFoundError) Error() string                { return n.message }
func (n *notFoundError) Description() string          { return n.description }
func (n *notFoundError) StatusCode() int              { return n.statusCode }
func (n *notFoundError) Code() string                 { return n.code }
func (n *notFoundError) Retries() int                 { return 0 }
func (n *notFoundError) InternalNotify() bool         { return true }
func (n *notFoundError) Abort() bool                  { return true }
func (n *notFoundError) ErrorDetails() map[string]any { return map[string]any{} }
