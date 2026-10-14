package dto

import (
	"encoding/json"

	"github.com/brienze1/xenon/internal/domain/entity"
	mcpErrors "github.com/brienze1/xenon/internal/integration/entrypoint/errors"
)

// NewExampleEdit leaves the response body out.
func NewExampleEdit(details entity.ExampleDetails) ExampleEdit {
	return ExampleEdit{
		Id:       details.Example.ID,
		Name:     details.Example.Name,
		Request:  NewExampleRequest(details.Example.Request),
		Response: NewStoredExampleResponseEdit(details.Example.Response),
	}
}

// MaskedWith applies mask to the JSON encoding of the whole edit.
func (e ExampleEdit) MaskedWith(mask func(text string) string) (ExampleEdit, error) {
	encoded, err := json.Marshal(e)
	if err != nil {
		return ExampleEdit{}, mcpErrors.FailUnknownAICall(err)
	}
	var masked ExampleEdit
	if err := json.Unmarshal([]byte(mask(string(encoded))), &masked); err != nil {
		return ExampleEdit{}, mcpErrors.FailUnknownAICall(err)
	}
	return masked, nil
}

// NewEditedExampleEdit keeps every stored field the edit leaves out; the response body stays out unless the edit passes one.
func NewEditedExampleEdit(details entity.ExampleDetails, edit *MCPTreeEdit) ExampleEdit {
	return ExampleEdit{
		Id:       details.Example.ID,
		Name:     overridden(edit.Name, details.Example.Name),
		Request:  NewEditedExampleRequest(NewExampleRequest(details.Example.Request), edit),
		Response: overridden(edit.Response, NewStoredExampleResponseEdit(details.Example.Response)),
	}
}

func (e ExampleEdit) WithTextBody(body string) ExampleEdit {
	e.Response = NewTextExampleResponseEdit(e.Response, body)
	return e
}
