package dto

import "github.com/brienze1/xenon/internal/domain/entity"

func (m *MCPTreeEdit) CreateExample(localId, requestId string) CreateExample {
	return NewCreateExample(localId, requestId, m.ItemName())
}

func (m *MCPTreeEdit) CreateExampleFromResponse(localId, requestId, tabId string) CreateExampleFromResponse {
	return NewCreateExampleFromResponse(localId, requestId, tabId, m.ItemName())
}

func (m *MCPTreeEdit) RenameExample(localId, requestId, exampleId string) RenameExample {
	return NewRenameExample(localId, requestId, exampleId, m.ItemName())
}

func (m *MCPTreeEdit) SaveExample(localId, requestId string, stored entity.ExampleDetails) SaveExample {
	return NewSaveExample(localId, requestId, stored.Revision, NewEditedExampleEdit(stored, m))
}
