package entity

import (
	"net/http"
	"slices"
)

const defaultResponseDescription = "Response"

type ApiResponse struct {
	Description string
	Examples    []ApiExample
	Code        int
}

func NewApiResponses(examples []SavedExample) []ApiResponse {
	responses := make([]ApiResponse, 0)
	for _, example := range examples {
		code := example.Response.Code
		index := slices.IndexFunc(responses, func(response ApiResponse) bool { return response.Code == code })
		if index < 0 {
			responses = append(responses, ApiResponse{Code: code, Description: ResponseDescription(code)})
			index = len(responses) - 1
		}
		responses[index].Examples = append(responses[index].Examples, NewApiExample(example))
	}
	slices.SortStableFunc(responses, func(left, right ApiResponse) int { return left.Code - right.Code })
	return responses
}

func ResponseDescription(code int) string {
	if text := http.StatusText(code); text != "" {
		return text
	}
	return defaultResponseDescription
}

func (a ApiResponse) Equal(other ApiResponse) bool {
	return a.Code == other.Code && a.Description == other.Description && slices.Equal(a.Examples, other.Examples)
}
