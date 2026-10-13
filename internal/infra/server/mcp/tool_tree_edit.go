package mcp

import (
	"context"
	"math"
	"time"

	"github.com/brienze1/xenon/internal/domain/enums/aichangekind"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

const (
	lastPosition = math.MaxInt32
	tokenExpired = "expired"
	tokenValid   = "valid"
)

type treeEditArguments struct {
	Name             *string                  `json:"name"`
	Method           *string                  `json:"method"`
	Url              *string                  `json:"url"`
	Description      *string                  `json:"description"`
	Params           *[]dto.KeyValue          `json:"params"`
	Headers          *[]dto.KeyValue          `json:"headers"`
	Body             *dto.Body                `json:"body"`
	Auth             *dto.Auth                `json:"auth"`
	Scripts          *dto.Scripts             `json:"scripts"`
	Response         *dto.ExampleResponseEdit `json:"response"`
	Variables        *[]variableRowArguments  `json:"variables"`
	Index            *int                     `json:"index"`
	Id               string                   `json:"id"`
	Kind             string                   `json:"kind"`
	ParentId         string                   `json:"parent_id"`
	RequestId        string                   `json:"request_id"`
	OwnerId          string                   `json:"owner_id"`
	Use              string                   `json:"use"`
	FromLastResponse bool                     `json:"from_last_response"`
}

type treeEditedView struct {
	Id     string `json:"id"`
	Kind   string `json:"kind"`
	Name   string `json:"name"`
	Method string `json:"method,omitempty"`
	Path   string `json:"path,omitempty"`
}

type tokenMetadataLine struct {
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	Type      string     `json:"type"`
	Scopes    string     `json:"scopes,omitempty"`
	State     string     `json:"state"`
	Use       string     `json:"use,omitempty"`
}

func (a treeEditArguments) name() string {
	if a.Name == nil {
		return ""
	}
	return *a.Name
}

func (a treeEditArguments) position() int {
	if a.Index == nil {
		return lastPosition
	}
	return *a.Index
}

func (a treeEditArguments) editsSettings() bool {
	return a.Auth != nil || a.Headers != nil || a.Scripts != nil || a.Description != nil
}

func newTokenMetadataLine(token dto.TokenInfo) tokenMetadataLine {
	line := tokenMetadataLine{ExpiresAt: token.ExpiresAt, Type: token.TokenType, Scopes: token.Scope, State: tokenValid, Use: token.UseToken}
	if token.Expired {
		line.State = tokenExpired
	}
	return line
}

func (t *tools) treeEdit(ctx context.Context, call Call) (Result, error) {
	var arguments treeEditArguments
	if err := call.Decode(&arguments); err != nil {
		return Result{}, err
	}
	result, err := t.editTree(ctx, call, arguments)
	if err != nil {
		return Result{}, err
	}
	result.Changes = append([]dto.AIChange{{Kind: aichangekind.Tree.String()}}, result.Changes...)
	return result, nil
}

func (t *tools) editTree(ctx context.Context, call Call, arguments treeEditArguments) (Result, error) {
	switch call.Action {
	case "create":
		return t.createTreeItem(ctx, call, arguments)
	case "get_token", "refresh_token", "use_token":
		return t.manageToken(ctx, call, arguments)
	}
	item, err := t.findWorkspaceItem(ctx, call.LocalId, arguments.Id)
	if err != nil {
		return Result{}, err
	}
	switch call.Action {
	case "update":
		return t.updateTreeItem(ctx, call, item, arguments)
	case "rename":
		return t.renameTreeItem(ctx, call, item, arguments.name())
	case "move":
		return t.moveTreeItem(ctx, call, item, arguments)
	case "duplicate":
		return t.duplicateTreeItem(ctx, call, item)
	default:
		return t.saveTreeSettings(ctx, call, item, arguments)
	}
}

func (t *tools) createTreeItem(ctx context.Context, call Call, arguments treeEditArguments) (Result, error) {
	switch arguments.Kind {
	case itemkind.Collection.String():
		node, err := t.controllers.TreeManage.CreateCollection(ctx, dto.CreateCollection{LocalId: call.LocalId, Name: arguments.name()})
		return t.treeNodeResult(ctx, call, node, err)
	case itemkind.Folder.String():
		node, err := t.controllers.TreeManage.CreateFolder(ctx, dto.CreateFolder{LocalId: call.LocalId, ParentId: arguments.ParentId, Name: arguments.name()})
		return t.treeNodeResult(ctx, call, node, err)
	case itemkind.Request.String():
		return t.createRequest(ctx, call, arguments)
	case itemkind.Example.String():
		return t.createExample(ctx, call, arguments)
	default:
		return Result{}, invalidArguments(call.Tool, "kind must be collection, folder, request or example")
	}
}

func (t *tools) treeNodeResult(ctx context.Context, call Call, node *dto.TreeNode, err error) (Result, error) {
	if err != nil {
		return Result{}, err
	}
	return t.treeItemResult(ctx, call, node.Id)
}

func (t *tools) treeItemResult(ctx context.Context, call Call, id string) (Result, error) {
	return t.treeRequestResult(ctx, call, id, "")
}

func (t *tools) treeRequestResult(ctx context.Context, call Call, id, method string) (Result, error) {
	item, err := t.findWorkspaceItem(ctx, call.LocalId, id)
	if err != nil {
		return Result{}, err
	}
	return Result{Payload: treeEditedView{Id: item.Id, Kind: item.Kind, Name: item.Name, Method: method, Path: item.Path}, Target: item.target()}, nil
}

func (t *tools) renameTreeItem(ctx context.Context, call Call, item workspaceItem, name string) (Result, error) {
	if item.isExample() {
		_, err := t.controllers.ExampleManage.Rename(ctx, dto.RenameExample{LocalId: call.LocalId, RequestId: item.RequestId, ExampleId: item.Id, Name: name})
		if err != nil {
			return Result{}, err
		}
		return t.treeItemResult(ctx, call, item.Id)
	}
	node, err := t.controllers.TreeManage.Rename(ctx, dto.RenameItem{LocalId: call.LocalId, Id: item.Id, Name: name})
	return t.treeNodeResult(ctx, call, node, err)
}

func (t *tools) moveTreeItem(ctx context.Context, call Call, item workspaceItem, arguments treeEditArguments) (Result, error) {
	if item.isExample() {
		return Result{}, invalidArguments(call.Tool, "an example cannot be moved")
	}
	if _, err := t.controllers.TreeManage.Move(ctx, dto.MoveItem{LocalId: call.LocalId, Id: item.Id, NewParentId: arguments.ParentId, Index: arguments.position()}); err != nil {
		return Result{}, err
	}
	return t.treeItemResult(ctx, call, item.Id)
}

func (t *tools) duplicateTreeItem(ctx context.Context, call Call, item workspaceItem) (Result, error) {
	if item.isExample() {
		duplicated, err := t.controllers.ExampleManage.Duplicate(ctx, dto.ExampleReference{LocalId: call.LocalId, RequestId: item.RequestId, ExampleId: item.Id})
		if err != nil {
			return Result{}, err
		}
		return t.treeItemResult(ctx, call, duplicated.Id)
	}
	node, err := t.controllers.TreeManage.Duplicate(ctx, dto.TreeItemReference{LocalId: call.LocalId, Id: item.Id})
	return t.treeNodeResult(ctx, call, node, err)
}

func (t *tools) saveTreeSettings(ctx context.Context, call Call, item workspaceItem, arguments treeEditArguments) (Result, error) {
	if !item.isContainer() {
		return Result{}, invalidArguments(call.Tool, "settings belong to a collection or a folder")
	}
	if arguments.Variables != nil {
		if item.Kind != itemkind.Collection.String() {
			return Result{}, invalidArguments(call.Tool, "only a collection has variables")
		}
		if _, err := t.saveVariableRows(ctx, call, collectionScope+scopeSeparator+item.Id, *arguments.Variables); err != nil {
			return Result{}, err
		}
	}
	if arguments.editsSettings() {
		if err := t.saveItemSettings(ctx, call, item, arguments); err != nil {
			return Result{}, err
		}
	}
	result, err := t.treeItemResult(ctx, call, item.Id)
	if err != nil {
		return Result{}, err
	}
	if arguments.Variables != nil {
		result.Changes = []dto.AIChange{{Kind: aichangekind.Variables.String()}}
	}
	return result, nil
}

func (t *tools) saveItemSettings(ctx context.Context, call Call, item workspaceItem, arguments treeEditArguments) error {
	stored, err := t.controllers.TreeSettings.GetSettings(ctx, dto.TreeItemReference{LocalId: call.LocalId, Id: item.Id})
	if err != nil {
		return err
	}
	settings := dto.SaveItemSettings{LocalId: call.LocalId, Id: item.Id, Auth: stored.Auth, Headers: stored.Headers, Description: stored.Description}
	if arguments.Auth != nil {
		settings.Auth = *arguments.Auth
	}
	if arguments.Headers != nil {
		settings.Headers = *arguments.Headers
	}
	if arguments.Description != nil {
		settings.Description = *arguments.Description
	}
	if arguments.Scripts != nil {
		settings.Scripts, settings.InheritScripts = arguments.Scripts, stored.InheritScripts
	}
	_, err = t.controllers.TreeSettings.SaveSettings(ctx, settings)
	return err
}

func (t *tools) manageToken(ctx context.Context, call Call, arguments treeEditArguments) (Result, error) {
	owner, err := t.findWorkspaceItem(ctx, call.LocalId, arguments.OwnerId)
	if err != nil {
		return Result{}, err
	}
	token, err := t.tokenAction(ctx, call, owner.Id, arguments.Use)
	if err != nil {
		return Result{}, err
	}
	return Result{Payload: newTokenMetadataLine(*token), Target: owner.target()}, nil
}

func (t *tools) tokenAction(ctx context.Context, call Call, ownerId, use string) (*dto.TokenInfo, error) {
	switch call.Action {
	case "get_token":
		return t.controllers.OAuth.GetNewToken(ctx, dto.OAuthTokenRequest{LocalId: call.LocalId, OwnerId: ownerId})
	case "refresh_token":
		return t.controllers.OAuth.Refresh(ctx, dto.OAuthTokenRequest{LocalId: call.LocalId, OwnerId: ownerId})
	default:
		return t.controllers.OAuth.Use(ctx, dto.OAuthTokenUse{LocalId: call.LocalId, OwnerId: ownerId, UseToken: use})
	}
}
