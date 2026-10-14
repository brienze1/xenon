package dto

import (
	"math"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
)

const mcpTreeEditLastPosition = math.MaxInt32

type MCPTreeEdit struct {
	Name             *string              `json:"name"`
	Method           *string              `json:"method"`
	Url              *string              `json:"url"`
	Description      *string              `json:"description"`
	Params           *[]KeyValue          `json:"params"`
	Headers          *[]KeyValue          `json:"headers"`
	Body             *Body                `json:"body"`
	Auth             *Auth                `json:"auth"`
	Scripts          *Scripts             `json:"scripts"`
	Response         *ExampleResponseEdit `json:"response"`
	Variables        *[]MCPVariableRow    `json:"variables"`
	Index            *int                 `json:"index"`
	Id               string               `json:"id"`
	Kind             string               `json:"kind"`
	ParentId         string               `json:"parent_id"`
	RequestId        string               `json:"request_id"`
	OwnerId          string               `json:"owner_id"`
	Use              string               `json:"use"`
	FromLastResponse bool                 `json:"from_last_response"`
}

// ItemKind matches kind exactly (case-sensitive, untrimmed); anything else is unknown.
func (m *MCPTreeEdit) ItemKind() enums.ItemKind {
	kind := itemkind.GetItemKind(m.Kind)
	if kind.String() != m.Kind {
		return itemkind.Unknown
	}
	return kind
}

func (m *MCPTreeEdit) ItemName() string {
	if m.Name == nil {
		return ""
	}
	return *m.Name
}

// Position is the move target index; without one the item goes last.
func (m *MCPTreeEdit) Position() int {
	if m.Index == nil {
		return mcpTreeEditLastPosition
	}
	return *m.Index
}

func (m *MCPTreeEdit) CopiesLastResponse() bool {
	return m.FromLastResponse
}

func (m *MCPTreeEdit) EditsSettings() bool {
	return m.Auth != nil || m.Headers != nil || m.Scripts != nil || m.Description != nil
}

func (m *MCPTreeEdit) HasVariables() bool {
	return m.Variables != nil
}

func (m *MCPTreeEdit) VariableRows() []MCPVariableRow {
	if m.Variables == nil {
		return nil
	}
	return *m.Variables
}

func (m *MCPTreeEdit) CreateCollection(localId string) CreateCollection {
	return NewCreateCollection(localId, m.ItemName())
}

func (m *MCPTreeEdit) CreateFolder(localId string) CreateFolder {
	return NewCreateFolder(localId, m.ParentId, m.ItemName())
}

func (m *MCPTreeEdit) RenameItem(localId, id string) RenameItem {
	return NewRenameItem(localId, id, m.ItemName())
}

func (m *MCPTreeEdit) MoveItem(localId, id string) MoveItem {
	return NewMoveItem(localId, id, m.ParentId, m.Position())
}

func (m *MCPTreeEdit) SaveItemSettings(localId, id string, stored ItemSettings) SaveItemSettings {
	return NewSaveItemSettings(localId, id, stored, m)
}

func (m *MCPTreeEdit) OAuthTokenRequest(localId, ownerId string) OAuthTokenRequest {
	return NewOAuthTokenRequest(localId, ownerId)
}

func (m *MCPTreeEdit) OAuthTokenUse(localId, ownerId string) OAuthTokenUse {
	return NewOAuthTokenUse(localId, ownerId, m.Use)
}

func (m *MCPTreeEdit) ItemId() string {
	return m.Id
}

func (m *MCPTreeEdit) RequestItemId() string {
	return m.RequestId
}

func (m *MCPTreeEdit) OwnerItemId() string {
	return m.OwnerId
}
