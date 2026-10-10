package errs

import (
	"errors"
	"net/http"
)

type serviceUnavailableError struct {
	err          error
	message      string
	description  string
	code         string
	errorDetails []ErrorDetails
	statusCode   int
}

func ServiceUnavailableError(err error, code, description string) BaseError {
	return &serviceUnavailableError{
		err:         err,
		statusCode:  http.StatusServiceUnavailable,
		message:     "service unavailable",
		description: description,
		code:        code,
		errorDetails: []ErrorDetails{{
			Attribute: "UNAVAILABLE",
			Messages:  []string{errorMessage(err)},
		}},
	}
}

func IsServiceUnavailableError(err error) bool {
	var target *serviceUnavailableError
	return errors.As(err, &target)
}

func (s *serviceUnavailableError) ToDto(transactionId string) Error {
	return toDto(transactionId, s.description, s.code, s.errorDetails)
}

func (s *serviceUnavailableError) SetStatusCode(statusCode int, description string) {
	s.statusCode = statusCode
	s.description = description
}

func (s *serviceUnavailableError) Is(target error) bool {
	var other *serviceUnavailableError
	return errors.As(target, &other)
}

func (s *serviceUnavailableError) Unwrap() error                { return s.err }
func (s *serviceUnavailableError) Error() string                { return s.message }
func (s *serviceUnavailableError) Description() string          { return s.description }
func (s *serviceUnavailableError) StatusCode() int              { return s.statusCode }
func (s *serviceUnavailableError) Code() string                 { return s.code }
func (s *serviceUnavailableError) Retries() int                 { return 0 }
func (s *serviceUnavailableError) InternalNotify() bool         { return true }
func (s *serviceUnavailableError) Abort() bool                  { return false }
func (s *serviceUnavailableError) ErrorDetails() map[string]any { return detailsToMap(s.errorDetails) }
