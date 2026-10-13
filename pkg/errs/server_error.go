package errs

import "errors"

type serverError struct {
	message     string
	description string
	code        string
	statusCode  int
}

func ServerError(description, code string, status int) BaseError {
	return &serverError{
		statusCode:  status,
		message:     "server error",
		description: description,
		code:        code,
	}
}

func (s *serverError) ToDto(transactionId string) Error {
	return toDto(transactionId, s.description, s.code, nil)
}

func (s *serverError) SetStatusCode(statusCode int, description string) {
	s.statusCode = statusCode
	s.description = description
}

func (s *serverError) Is(target error) bool {
	var other *serverError
	return errors.As(target, &other)
}

func (s *serverError) Error() string                { return s.message }
func (s *serverError) Description() string          { return s.description }
func (s *serverError) StatusCode() int              { return s.statusCode }
func (s *serverError) Code() string                 { return s.code }
func (s *serverError) Retries() int                 { return 5 }
func (s *serverError) InternalNotify() bool         { return true }
func (s *serverError) Abort() bool                  { return false }
func (s *serverError) ErrorDetails() map[string]any { return map[string]any{} }
