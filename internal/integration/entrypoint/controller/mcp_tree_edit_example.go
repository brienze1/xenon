package controller

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	mcpErrors "github.com/brienze1/xenon/internal/integration/entrypoint/errors"
)

func (m *mcpTreeEdit) createExample(ctx context.Context, call dto.MCPToolCall, arguments dto.MCPTreeEdit) (dto.MCPToolResult, error) {
	entry, err := findMCPTreeEntry(ctx, m.searchTreeService, call.LocalId, arguments.RequestItemId())
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	if !entry.IsRequest() {
		return dto.MCPToolResult{}, mcpErrors.RejectInvalidArguments(call.Tool, mcpTreeEditRequestReason)
	}
	if arguments.CopiesLastResponse() {
		return m.createExampleFromResponse(ctx, call, entry, arguments)
	}

	request := arguments.CreateExample(call.LocalId, entry.ID())
	err = validateCreateExample(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	created, err := m.manageExampleService.Create(ctx, request.LocalId, request.RequestId, request.Name)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return m.edited(ctx, call, dto.NewExample(*created).Id, "")
}

func (m *mcpTreeEdit) createExampleFromResponse(ctx context.Context, call dto.MCPToolCall, entry entity.TreeEntry, arguments dto.MCPTreeEdit) (dto.MCPToolResult, error) {
	request := arguments.CreateExampleFromResponse(call.LocalId, entry.ID(), call.TabId())
	err := validateCreateExampleFromResponse(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	details, err := m.manageExampleService.CreateFromResponse(ctx, request.LocalId, request.RequestId, request.TabId, request.Name)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	err = m.maskExample(ctx, call, *details)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return m.edited(ctx, call, dto.NewExample(*details).Id, "")
}

func (m *mcpTreeEdit) maskExample(ctx context.Context, call dto.MCPToolCall, details entity.ExampleDetails) error {
	mask := func(text string) string {
		return guardMCPText(ctx, m.errorHandler, m.guardAIOutputService, call.LocalId, text)
	}
	example := dto.NewExample(details)
	edit, err := dto.NewExampleEdit(details).MaskedWith(mask)
	if err != nil {
		return err
	}
	body, found, err := m.exampleBody(ctx, call, example.Response)
	if err != nil {
		return err
	}
	if found {
		edit = edit.WithTextBody(mask(body))
	}
	return m.saveExample(ctx, dto.NewSaveExample(call.LocalId, example.RequestId, example.Revision, edit))
}

func (m *mcpTreeEdit) exampleBody(ctx context.Context, call dto.MCPToolCall, response dto.ExampleResponse) (string, bool, error) {
	if text, inline := response.InlineText(); inline {
		return text, true, nil
	}
	if !response.StreamsText() {
		return "", false, nil
	}
	body, err := readMCPResponseText(ctx, m.errorHandler, m.manageResponseService, call.TabId())
	if err != nil {
		return "", false, err
	}
	return body, true, nil
}

func (m *mcpTreeEdit) updateExample(ctx context.Context, call dto.MCPToolCall, entry entity.TreeEntry, arguments dto.MCPTreeEdit) (dto.MCPToolResult, error) {
	request := dto.NewExampleReference(call.LocalId, entry.RequestID, entry.ID())
	err := validateGetExample(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	stored, err := m.manageExampleService.Get(ctx, request.LocalId, request.RequestId, request.ExampleId)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	err = m.saveExample(ctx, arguments.SaveExample(call.LocalId, entry.RequestID, *stored))
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return m.edited(ctx, call, entry.ID(), "")
}

func (m *mcpTreeEdit) renameExample(ctx context.Context, call dto.MCPToolCall, entry entity.TreeEntry, arguments dto.MCPTreeEdit) (dto.MCPToolResult, error) {
	request := arguments.RenameExample(call.LocalId, entry.RequestID, entry.ID())
	err := validateRenameExample(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	_, err = m.manageExampleService.Rename(ctx, request.LocalId, request.RequestId, request.ExampleId, request.Name)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return m.edited(ctx, call, entry.ID(), "")
}

func (m *mcpTreeEdit) duplicateExample(ctx context.Context, call dto.MCPToolCall, entry entity.TreeEntry) (dto.MCPToolResult, error) {
	request := dto.NewExampleReference(call.LocalId, entry.RequestID, entry.ID())
	err := validateDuplicateExample(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	duplicated, err := m.manageExampleService.Duplicate(ctx, request.LocalId, request.RequestId, request.ExampleId)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return m.edited(ctx, call, dto.NewExample(*duplicated).Id, "")
}

func (m *mcpTreeEdit) saveExample(ctx context.Context, request dto.SaveExample) error {
	err := validateSaveExample(&request)
	if err != nil {
		return err
	}

	_, err = m.manageExampleService.Save(ctx, request.LocalId, request.RequestId, request.ToEntity(), request.KeepsBody(), request.BaseRevision, request.Force)
	return err
}
