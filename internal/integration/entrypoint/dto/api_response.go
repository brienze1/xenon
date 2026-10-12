package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ApiResponse struct {
	Description string       `json:"description"`
	Examples    []ApiExample `json:"examples"`
	Code        int          `json:"code"`
}

func NewApiResponse(response entity.ApiResponse) ApiResponse {
	documented := ApiResponse{Code: response.Code, Description: response.Description, Examples: make([]ApiExample, 0, len(response.Examples))}
	for _, example := range response.Examples {
		documented.Examples = append(documented.Examples, NewApiExample(example))
	}
	return documented
}
