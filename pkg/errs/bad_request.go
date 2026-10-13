package errs

import (
	"errors"
	"net/http"
)

type badRequestError struct {
	message      string
	description  string
	code         string
	errorDetails []ErrorDetails
	statusCode   int
}

func BadRequestError(description, code string, details ...ErrorDetails) BaseError {
	return &badRequestError{
		statusCode:   http.StatusBadRequest,
		message:      "bad request",
		description:  description,
		code:         code,
		errorDetails: details,
	}
}

func (b *badRequestError) ToDto(transactionId string) Error {
	return toDto(transactionId, b.description, b.code, b.errorDetails)
}

func (b *badRequestError) SetStatusCode(statusCode int, description string) {
	b.statusCode = statusCode
	b.description = description
}

func (b *badRequestError) Is(target error) bool {
	var other *badRequestError
	return errors.As(target, &other)
}

func (b *badRequestError) Error() string                { return b.message }
func (b *badRequestError) Description() string          { return b.description }
func (b *badRequestError) StatusCode() int              { return b.statusCode }
func (b *badRequestError) Code() string                 { return b.code }
func (b *badRequestError) Retries() int                 { return 0 }
func (b *badRequestError) InternalNotify() bool         { return true }
func (b *badRequestError) Abort() bool                  { return true }
func (b *badRequestError) ErrorDetails() map[string]any { return detailsToMap(b.errorDetails) }
