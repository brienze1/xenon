package errs

import (
	"errors"
	"net/http"
)

type internalServerError struct {
	err          error
	message      string
	description  string
	code         string
	errorDetails []ErrorDetails
	statusCode   int
}

func InternalServerError(err error, code, message string) BaseError {
	return &internalServerError{
		err:         err,
		statusCode:  http.StatusInternalServerError,
		message:     "internal server error",
		description: message,
		code:        code,
		errorDetails: []ErrorDetails{{
			Attribute: "UNKNOWN_ERROR",
			Messages:  []string{errorMessage(err)},
		}},
	}
}

func (i *internalServerError) ToDto(transactionId string) Error {
	return toDto(transactionId, i.description, i.code, i.errorDetails)
}

func (i *internalServerError) SetStatusCode(statusCode int, description string) {
	i.statusCode = statusCode
	i.description = description
}

func (i *internalServerError) Is(target error) bool {
	var other *internalServerError
	return errors.As(target, &other)
}

func (i *internalServerError) Unwrap() error                { return i.err }
func (i *internalServerError) Error() string                { return i.message }
func (i *internalServerError) Description() string          { return i.description }
func (i *internalServerError) StatusCode() int              { return i.statusCode }
func (i *internalServerError) Code() string                 { return i.code }
func (i *internalServerError) Retries() int                 { return 5 }
func (i *internalServerError) InternalNotify() bool         { return true }
func (i *internalServerError) Abort() bool                  { return false }
func (i *internalServerError) ErrorDetails() map[string]any { return detailsToMap(i.errorDetails) }

func errorMessage(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
