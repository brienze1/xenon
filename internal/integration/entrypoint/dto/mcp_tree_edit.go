package dto

import (
	"math"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
)

const mcpTreeEditLastPosition = math.MaxInt32

type MCPTreeEdit struct {
	Name             *string              `json:"name" description:"Item name; required to create a collection, folder or request." used_by:"create,update,rename" required_for:"rename"`
	Method           *string              `json:"method" description:"HTTP method of the request." used_by:"create,update"`
	Url              *string              `json:"url" description:"URL of the request." used_by:"create,update"`
	Description      *string              `json:"description" description:"Item description." used_by:"create,update,save_settings"`
	Params           *[]KeyValue          `json:"params" description:"Query parameters of the request." used_by:"create,update"`
	Headers          *[]KeyValue          `json:"headers" description:"Headers of the request, collection or folder." used_by:"create,update,save_settings"`
	Body             *Body                `json:"body" description:"Request body." used_by:"create,update" validated_as:"request.body,example.request.body"`
	Auth             *Auth                `json:"auth" description:"Auth of the request, collection or folder." used_by:"create,update,save_settings" validated_as:"request.auth,example.request.auth"`
	Scripts          *Scripts             `json:"scripts" description:"Pre-request and post-response scripts." used_by:"create,update,save_settings" validated_as:"request.scripts,example.request.scripts"`
	Response         *ExampleResponseEdit `json:"response" description:"Saved response of an example." used_by:"update" validated_as:"example.response"`
	Variables        *[]MCPVariableRow    `json:"variables" description:"Collection variables; collections only." used_by:"save_settings"`
	Index            *int                 `json:"index" description:"New 0-based position among the siblings." used_by:"move"`
	Id               string               `json:"id" description:"Id of the item to change." used_by:"update,rename,move,duplicate,save_settings" required_for:"update,rename,move,duplicate,save_settings" validated_as:"example_id"`
	Kind             string               `json:"kind" description:"Kind of item to create." used_by:"create" required_for:"create" enum:"collection,folder,request,example"`
	ParentId         string               `json:"parent_id" description:"Parent collection or folder; required to create a folder or request." used_by:"create,move"`
	RequestId        string               `json:"request_id" description:"Request the example belongs to; required to create an example." used_by:"create"`
	OwnerId          string               `json:"owner_id" description:"Id of the item that owns the OAuth token." used_by:"get_token,refresh_token,use_token" required_for:"get_token,refresh_token,use_token"`
	Use              string               `json:"use" description:"Which token to use." used_by:"use_token" required_for:"use_token" enum:"access_token,id_token" validated_as:"use_token"`
	FromLastResponse bool                 `json:"from_last_response" description:"Create the example from the last response of the request." used_by:"create"`
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
