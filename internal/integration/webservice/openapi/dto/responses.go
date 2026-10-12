package dto

import (
	"go.yaml.in/yaml/v3"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type Responses []Response

func NewResponses(responses []entity.ApiResponse) Responses {
	documented := make(Responses, 0, len(responses))
	for _, response := range responses {
		documented = append(documented, NewResponse(response))
	}
	return documented
}

func (r Responses) MarshalYAML() (any, error) {
	return mappingOf(r, func(response Response) string { return response.Code })
}

func (r *Responses) UnmarshalYAML(node *yaml.Node) error {
	entries, err := entriesOf(node, func(response *Response, code string) { response.Code = code })
	*r = entries
	return err
}

func (r Responses) ToEntity() ([]entity.ApiResponse, error) {
	responses := make([]entity.ApiResponse, 0, len(r))
	for index := range r {
		response, numeric, err := r[index].ToEntity()
		if err != nil {
			return nil, err
		}
		if numeric {
			responses = append(responses, response)
		}
	}
	return responses, nil
}
