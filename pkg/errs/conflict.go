package errs

import (
	"errors"
	"net/http"
)

type conflictError struct {
	message      string
	description  string
	code         string
	errorDetails []ErrorDetails
	statusCode   int
}

func ConflictError(description, code string, details ...ErrorDetails) BaseError {
	return &conflictError{
		statusCode:   http.StatusConflict,
		message:      "conflict",
		description:  description,
		code:         code,
		errorDetails: details,
	}
}

func IsConflictError(err error) bool {
	var target *conflictError
	return errors.As(err, &target)
}

func (c *conflictError) ToDto(transactionId string) Error {
	return toDto(transactionId, c.description, c.code, c.errorDetails)
}

func (c *conflictError) SetStatusCode(statusCode int, description string) {
	c.statusCode = statusCode
	c.description = description
}

func (c *conflictError) Is(target error) bool {
	var other *conflictError
	return errors.As(target, &other)
}

func (c *conflictError) Error() string                { return c.message }
func (c *conflictError) Description() string          { return c.description }
func (c *conflictError) StatusCode() int              { return c.statusCode }
func (c *conflictError) Code() string                 { return c.code }
func (c *conflictError) Retries() int                 { return 0 }
func (c *conflictError) InternalNotify() bool         { return true }
func (c *conflictError) Abort() bool                  { return true }
func (c *conflictError) ErrorDetails() map[string]any { return detailsToMap(c.errorDetails) }
