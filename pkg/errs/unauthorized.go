package errs

import (
	"errors"
	"net/http"
)

type unauthorizedError struct {
	message     string
	description string
	code        string
	statusCode  int
}

func UnauthorizedError(description, code string) BaseError {
	return &unauthorizedError{
		statusCode:  http.StatusUnauthorized,
		message:     "unauthorized",
		description: description,
		code:        code,
	}
}

func (u *unauthorizedError) ToDto(transactionId string) Error {
	return toDto(transactionId, u.description, u.code, nil)
}

func (u *unauthorizedError) SetStatusCode(statusCode int, description string) {
	u.statusCode = statusCode
	u.description = description
}

func (u *unauthorizedError) Is(target error) bool {
	var other *unauthorizedError
	return errors.As(target, &other)
}

func (u *unauthorizedError) Error() string                { return u.message }
func (u *unauthorizedError) Description() string          { return u.description }
func (u *unauthorizedError) StatusCode() int              { return u.statusCode }
func (u *unauthorizedError) Code() string                 { return u.code }
func (u *unauthorizedError) Retries() int                 { return 0 }
func (u *unauthorizedError) InternalNotify() bool         { return true }
func (u *unauthorizedError) Abort() bool                  { return true }
func (u *unauthorizedError) ErrorDetails() map[string]any { return map[string]any{} }
