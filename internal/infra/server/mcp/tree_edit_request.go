package mcp

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/enums/aichangekind"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

func (a treeEditArguments) editsRequest() bool {
	return a.Method != nil || a.Url != nil || a.Description != nil || a.Params != nil || a.Headers != nil || a.Body != nil || a.Auth != nil || a.Scripts != nil
}

// mergedRequest keeps the stored scripts unless the call gives new ones.
func (a treeEditArguments) mergedRequest(request dto.Request) dto.Request {
	request.Scripts, request.InheritScripts = nil, nil
	if a.Name != nil {
		request.Name = *a.Name
	}
	if a.Method != nil {
		request.Method = *a.Method
	}
	if a.Url != nil {
		request.Url = *a.Url
	}
	if a.Description != nil {
		request.Description = *a.Description
	}
	if a.Params != nil {
		request.Params = *a.Params
	}
	if a.Headers != nil {
		request.Headers = *a.Headers
	}
	if a.Body != nil {
		request.Body = *a.Body
	}
	if a.Auth != nil {
		request.Auth = *a.Auth
	}
	if a.Scripts != nil {
		request.Scripts = a.Scripts
	}
	return request
}

func (t *tools) createRequest(ctx context.Context, call Call, arguments treeEditArguments) (Result, error) {
	node, err := t.controllers.TreeManage.CreateRequest(ctx, dto.CreateRequest{LocalId: call.LocalId, ParentId: arguments.ParentId, Name: arguments.name()})
	if err != nil {
		return Result{}, err
	}
	method := ""
	if node.Method != nil {
		method = *node.Method
	}
	if arguments.editsRequest() {
		stored, err := t.controllers.RequestGet.Get(ctx, dto.OpenRequest{LocalId: call.LocalId, Id: node.Id})
		if err != nil {
			return Result{}, err
		}
		saved, err := t.controllers.RequestSave.Save(ctx, dto.SaveRequest{LocalId: call.LocalId, BaseRevision: stored.Revision, Request: arguments.mergedRequest(dto.Request{Id: node.Id, Name: node.Name})})
		if err != nil {
			return Result{}, err
		}
		method = saved.Method
	}
	return t.treeRequestResult(ctx, call, node.Id, method)
}

func (t *tools) updateTreeItem(ctx context.Context, call Call, item workspaceItem, arguments treeEditArguments) (Result, error) {
	switch {
	case item.isExample():
		return t.updateExample(ctx, call, item, arguments)
	case item.isRequest():
		return t.updateRequest(ctx, call, item, arguments)
	default:
		return Result{}, invalidArguments(call.Tool, "update edits a request or an example; use rename or save_settings for a collection or folder")
	}
}

func (t *tools) updateRequest(ctx context.Context, call Call, item workspaceItem, arguments treeEditArguments) (Result, error) {
	stored, err := t.controllers.RequestGet.Get(ctx, dto.OpenRequest{LocalId: call.LocalId, Id: item.Id})
	if err != nil {
		return Result{}, err
	}
	saved, err := t.controllers.RequestSave.Save(ctx, dto.SaveRequest{LocalId: call.LocalId, BaseRevision: stored.Revision, Request: arguments.mergedRequest(*stored)})
	if err != nil {
		return Result{}, err
	}
	result, err := t.treeRequestResult(ctx, call, item.Id, saved.Method)
	if err != nil {
		return Result{}, err
	}
	result.Changes = []dto.AIChange{{Kind: aichangekind.Request.String(), RequestId: saved.Id, Revision: saved.Revision}}
	return result, nil
}
