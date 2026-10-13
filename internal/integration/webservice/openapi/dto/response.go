package dto

import (
	"strconv"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type Response struct {
	Code        string  `yaml:"-"`
	Description string  `yaml:"description"`
	Content     Content `yaml:"content,omitempty"`
}

func NewResponse(response entity.ApiResponse) Response {
	content := Content{}
	for _, example := range response.Examples {
		index := len(content) - 1
		if index < 0 || content[index].Name != example.MediaType {
			content = append(content, MediaType{Name: example.MediaType})
			index++
		}
		content[index].Examples = append(content[index].Examples, NewExample(example))
	}
	return Response{Code: strconv.Itoa(response.Code), Description: response.Description, Content: content}
}

// ToEntity reads a response with a numeric status; "default" and ranges like "2XX" have no status to save.
func (r *Response) ToEntity() (entity.ApiResponse, bool, error) {
	code, err := strconv.Atoi(r.Code)
	if err != nil {
		return entity.ApiResponse{}, false, nil
	}
	response := entity.ApiResponse{Code: code, Description: r.Description}
	for index := range r.Content {
		examples, err := r.Content[index].examples(r.Description)
		if err != nil {
			return entity.ApiResponse{}, false, err
		}
		response.Examples = append(response.Examples, examples...)
	}
	return response, true, nil
}

func (r Response) MarshalYAML() (any, error) {
	return mappingOfFields(always("description", r.Description), when(len(r.Content) > 0, "content", r.Content))
}
