package errs

type BaseError interface {
	Code() string
	ErrorDetails() map[string]any
	Error() string
	Description() string
	StatusCode() int
	SetStatusCode(statusCode int, description string)
	ToDto(transactionId string) Error
	Is(target error) bool
	Retries() int
	InternalNotify() bool
	Abort() bool
}

type Error struct {
	Error ErrorData `json:"error"`
}

type ErrorData struct {
	Id           string         `json:"id,omitempty"`
	Description  string         `json:"description,omitempty"`
	Code         string         `json:"code,omitempty"`
	ErrorDetails []ErrorDetails `json:"error_details,omitempty"`
}

type ErrorDetails struct {
	Attribute string   `json:"attribute"`
	Messages  []string `json:"messages"`
}

func toDto(transactionId, description, code string, details []ErrorDetails) Error {
	return Error{
		Error: ErrorData{
			Id:           transactionId,
			Description:  description,
			Code:         code,
			ErrorDetails: details,
		},
	}
}

func detailsToMap(details []ErrorDetails) map[string]any {
	result := make(map[string]any)
	for _, detail := range details {
		result["attribute"] = detail.Attribute
		result["messages"] = detail.Messages
	}
	return result
}
