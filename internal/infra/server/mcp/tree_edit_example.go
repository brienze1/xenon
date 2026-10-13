package mcp

import (
	"context"
	"encoding/json"
	"io"

	"github.com/brienze1/xenon/internal/domain/enums/bodyencoding"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

func (a treeEditArguments) mergedExample(example dto.Example) dto.ExampleEdit {
	edit := dto.ExampleEdit{Id: example.Id, Name: example.Name, Request: example.Request, Response: storedExampleResponse(example.Response)}
	if a.Name != nil {
		edit.Name = *a.Name
	}
	if a.Method != nil {
		edit.Request.Method = *a.Method
	}
	if a.Url != nil {
		edit.Request.Url = *a.Url
	}
	if a.Params != nil {
		edit.Request.Params = *a.Params
	}
	if a.Headers != nil {
		edit.Request.Headers = *a.Headers
	}
	if a.Body != nil {
		edit.Request.Body = *a.Body
	}
	if a.Auth != nil {
		edit.Request.Auth = *a.Auth
	}
	if a.Response != nil {
		edit.Response = *a.Response
	}
	return edit
}

// storedExampleResponse leaves the body out, so a save keeps the stored body.
func storedExampleResponse(response dto.ExampleResponse) dto.ExampleResponseEdit {
	return dto.ExampleResponseEdit{Status: response.Status, ContentType: response.ContentType, Headers: response.Headers, Code: response.Code}
}

func (t *tools) createExample(ctx context.Context, call Call, arguments treeEditArguments) (Result, error) {
	request, err := t.findWorkspaceItem(ctx, call.LocalId, arguments.RequestId)
	if err != nil {
		return Result{}, err
	}
	if !request.isRequest() {
		return Result{}, invalidArguments(call.Tool, "request_id must be a request")
	}
	if !arguments.FromLastResponse {
		created, err := t.controllers.ExampleManage.Create(ctx, dto.CreateExample{LocalId: call.LocalId, RequestId: request.Id, Name: arguments.name()})
		if err != nil {
			return Result{}, err
		}
		return t.treeItemResult(ctx, call, created.Id)
	}
	created, err := t.controllers.ExampleManage.CreateFromResponse(ctx, dto.CreateExampleFromResponse{LocalId: call.LocalId, RequestId: request.Id, TabId: call.TabId(), Name: arguments.name()})
	if err != nil {
		return Result{}, err
	}
	if err := t.maskExample(ctx, call, *created); err != nil {
		return Result{}, err
	}
	return t.treeItemResult(ctx, call, created.Id)
}

// maskExample rewrites an example made from a response with the leak guard applied, because it is AI output (FR-021).
func (t *tools) maskExample(ctx context.Context, call Call, example dto.Example) error {
	edit, err := maskedExampleEdit(ctx, call, dto.ExampleEdit{Id: example.Id, Name: example.Name, Request: example.Request, Response: storedExampleResponse(example.Response)})
	if err != nil {
		return err
	}
	body, err := t.exampleBody(ctx, call, example.Response)
	if err != nil {
		return err
	}
	if body != nil {
		masked := call.Masked(ctx, *body)
		edit.Response.Body, edit.Response.BodyEncoding = &masked, bodyencoding.Text.String()
	}
	_, err = t.controllers.ExampleManage.Save(ctx, dto.SaveExample{LocalId: call.LocalId, RequestId: example.RequestId, BaseRevision: example.Revision, Example: edit})
	return err
}

func (t *tools) exampleBody(ctx context.Context, call Call, response dto.ExampleResponse) (*string, error) {
	if response.Body != nil || response.IsBinary {
		return response.Body, nil
	}
	opened, err := t.controllers.RequestResponse.OpenBody(ctx, dto.NewTabReference(call.TabId()))
	if err != nil {
		return nil, err
	}
	defer func() { _ = opened.Content.Close() }()
	content, err := io.ReadAll(opened.Content)
	if err != nil {
		return nil, err
	}
	body := string(content)
	return &body, nil
}

func maskedExampleEdit(ctx context.Context, call Call, edit dto.ExampleEdit) (dto.ExampleEdit, error) {
	encoded, err := json.Marshal(edit)
	if err != nil {
		return dto.ExampleEdit{}, err
	}
	var masked dto.ExampleEdit
	if err := json.Unmarshal([]byte(call.Masked(ctx, string(encoded))), &masked); err != nil {
		return dto.ExampleEdit{}, err
	}
	return masked, nil
}

func (t *tools) updateExample(ctx context.Context, call Call, item workspaceItem, arguments treeEditArguments) (Result, error) {
	stored, err := t.controllers.ExampleManage.Get(ctx, dto.ExampleReference{LocalId: call.LocalId, RequestId: item.RequestId, ExampleId: item.Id})
	if err != nil {
		return Result{}, err
	}
	if _, err := t.controllers.ExampleManage.Save(ctx, dto.SaveExample{LocalId: call.LocalId, RequestId: item.RequestId, BaseRevision: stored.Revision, Example: arguments.mergedExample(*stored)}); err != nil {
		return Result{}, err
	}
	return t.treeItemResult(ctx, call, item.Id)
}
