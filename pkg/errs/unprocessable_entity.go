package errs

import (
	"errors"
	"net/http"
)

type unprocessableEntityError struct {
	message     string
	description string
	code        string
	statusCode  int
}

func UnprocessableEntityError(description, code string) BaseError {
	return &unprocessableEntityError{
		statusCode:  http.StatusUnprocessableEntity,
		message:     "unprocessable entity",
		description: description,
		code:        code,
	}
}

func (u *unprocessableEntityError) ToDto(transactionId string) Error {
	return toDto(transactionId, u.description, u.code, nil)
}

func (u *unprocessableEntityError) SetStatusCode(statusCode int, description string) {
	u.statusCode = statusCode
	u.description = description
}

func (u *unprocessableEntityError) Is(target error) bool {
	var other *unprocessableEntityError
	return errors.As(target, &other)
}

func (u *unprocessableEntityError) Error() string                { return u.message }
func (u *unprocessableEntityError) Description() string          { return u.description }
func (u *unprocessableEntityError) StatusCode() int              { return u.statusCode }
func (u *unprocessableEntityError) Code() string                 { return u.code }
func (u *unprocessableEntityError) Retries() int                 { return 0 }
func (u *unprocessableEntityError) InternalNotify() bool         { return true }
func (u *unprocessableEntityError) Abort() bool                  { return true }
func (u *unprocessableEntityError) ErrorDetails() map[string]any { return map[string]any{} }
