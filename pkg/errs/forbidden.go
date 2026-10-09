package errs

import (
	"errors"
	"net/http"
)

type forbiddenError struct {
	message     string
	description string
	code        string
	statusCode  int
}

func ForbiddenError(description, code string) BaseError {
	return &forbiddenError{
		statusCode:  http.StatusForbidden,
		message:     "forbidden",
		description: description,
		code:        code,
	}
}

func (f *forbiddenError) ToDto(transactionId string) Error {
	return toDto(transactionId, f.description, f.code, nil)
}

func (f *forbiddenError) SetStatusCode(statusCode int, description string) {
	f.statusCode = statusCode
	f.description = description
}

func (f *forbiddenError) Is(target error) bool {
	var other *forbiddenError
	return errors.As(target, &other)
}

func (f *forbiddenError) Error() string                { return f.message }
func (f *forbiddenError) Description() string          { return f.description }
func (f *forbiddenError) StatusCode() int              { return f.statusCode }
func (f *forbiddenError) Code() string                 { return f.code }
func (f *forbiddenError) Retries() int                 { return 0 }
func (f *forbiddenError) InternalNotify() bool         { return true }
func (f *forbiddenError) Abort() bool                  { return true }
func (f *forbiddenError) ErrorDetails() map[string]any { return map[string]any{} }
