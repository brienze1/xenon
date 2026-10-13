package errs

import (
	"encoding/json"
	"errors"
	"io"
)

type clientError struct {
	message      string
	description  string
	code         string
	errorDetails []ErrorDetails
	statusCode   int
}

func ClientError(description, code string, status int, responseBody ...io.ReadCloser) BaseError {
	return &clientError{
		statusCode:   status,
		message:      "client error",
		description:  description,
		code:         code,
		errorDetails: parseErrorDetails(responseBody),
	}
}

func (c *clientError) ToDto(transactionId string) Error {
	return toDto(transactionId, c.description, c.code, c.errorDetails)
}

func (c *clientError) SetStatusCode(statusCode int, description string) {
	c.statusCode = statusCode
	c.description = description
}

func (c *clientError) Is(target error) bool {
	var other *clientError
	return errors.As(target, &other)
}

func (c *clientError) Error() string                { return c.message }
func (c *clientError) Description() string          { return c.description }
func (c *clientError) StatusCode() int              { return c.statusCode }
func (c *clientError) Code() string                 { return c.code }
func (c *clientError) Retries() int                 { return 0 }
func (c *clientError) InternalNotify() bool         { return true }
func (c *clientError) Abort() bool                  { return true }
func (c *clientError) ErrorDetails() map[string]any { return detailsToMap(c.errorDetails) }

func parseErrorDetails(bodies []io.ReadCloser) []ErrorDetails {
	if len(bodies) == 0 || bodies[0] == nil {
		return nil
	}
	defer func() { _ = bodies[0].Close() }()
	var body Error
	if err := json.NewDecoder(bodies[0]).Decode(&body); err != nil {
		return nil
	}
	return body.Error.ErrorDetails
}
