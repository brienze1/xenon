package errs

import (
	"errors"
	"net/http"
)

type panicError struct {
	message     string
	description string
	code        string
	statusCode  int
}

func PanicError(message string) BaseError {
	return &panicError{
		statusCode:  http.StatusInternalServerError,
		message:     "panic",
		description: message,
		code:        "PNC-00500",
	}
}

func (p *panicError) ToDto(transactionId string) Error {
	return toDto(transactionId, p.description, p.code, nil)
}

func (p *panicError) SetStatusCode(statusCode int, description string) {
	p.statusCode = statusCode
	p.description = description
}

func (p *panicError) Is(target error) bool {
	var other *panicError
	return errors.As(target, &other)
}

func (p *panicError) Error() string                { return p.message }
func (p *panicError) Description() string          { return p.description }
func (p *panicError) StatusCode() int              { return p.statusCode }
func (p *panicError) Code() string                 { return p.code }
func (p *panicError) Retries() int                 { return 0 }
func (p *panicError) InternalNotify() bool         { return true }
func (p *panicError) Abort() bool                  { return false }
func (p *panicError) ErrorDetails() map[string]any { return map[string]any{} }
