package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/idutils"
	"github.com/brienze1/xenon/pkg/logger"
)

type manageDocumentation struct {
	findWorkspaceRegistry usecase.FindWorkspaceRegistry
	source                documentationSource
	writer                documentationWriter
}

func NewManageDocumentationService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	findTreeItem usecase.FindTreeItem,
	findRequest usecase.FindRequest,
	loadTree usecase.LoadTree,
	saveTreeItem usecase.SaveTreeItem,
	saveRequestExamples usecase.SaveRequestExamples,
	createTreeItem usecase.CreateTreeItem,
	deleteTreeItem usecase.DeleteTreeItem,
	renameTreeItem usecase.RenameTreeItem,
	writeApiDocument usecase.WriteApiDocument,
	readApiDocument usecase.ReadApiDocument,
) adapter.ManageDocumentationService {
	return &manageDocumentation{
		findWorkspaceRegistry: findWorkspaceRegistry,
		source: documentationSource{
			findTreeItem:     findTreeItem,
			findRequest:      findRequest,
			writeApiDocument: writeApiDocument,
			readApiDocument:  readApiDocument,
		},
		writer: documentationWriter{
			findTreeItem:        findTreeItem,
			loadTree:            loadTree,
			saveTreeItem:        saveTreeItem,
			saveRequestExamples: saveRequestExamples,
			createTreeItem:      createTreeItem,
			deleteTreeItem:      deleteTreeItem,
			renameTreeItem:      renameTreeItem,
		},
	}
}

func (m *manageDocumentation) Get(ctx context.Context, localId, id string) (*entity.ApiDocumentation, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "id": id})

	workspace, err := m.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return nil, err
	}
	documented, err := m.source.collection(ctx, workspace.Path, id)
	if err != nil {
		return nil, err
	}
	documentation, err := m.source.documentation(ctx, *documented)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "id": id, "result": *documentation})
	return documentation, nil
}

func (m *manageDocumentation) Preview(ctx context.Context, text string) (*entity.ApiDocumentation, error) {
	logger.Info(ctx, "Started", map[string]any{"text": text})

	documentation, err := m.source.read(ctx, text)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"text": text, "result": *documentation})
	return documentation, nil
}

func (m *manageDocumentation) Plan(ctx context.Context, localId, id, text string) (*entity.DocumentationPlan, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "id": id, "text": text})

	workspace, err := m.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return nil, err
	}
	plan, err := m.source.plan(ctx, workspace.Path, id, text, idutils.NewUlid)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "id": id, "text": text, "result": *plan})
	return plan, nil
}

func (m *manageDocumentation) Apply(ctx context.Context, localId, id, text string, confirmDeletions, keepDeleted bool) (*entity.DocumentationResult, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "id": id, "text": text, "confirmDeletions": confirmDeletions, "keepDeleted": keepDeleted})

	workspace, err := m.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return nil, err
	}
	planned, err := m.source.plan(ctx, workspace.Path, id, text, idutils.NewUlid)
	if err != nil {
		return nil, err
	}
	plan := planned.Applying(keepDeleted)
	if err := plan.ConfirmedBy(confirmDeletions); err != nil {
		return nil, err
	}
	written, err := m.writer.write(ctx, workspace.Path, plan)
	if err != nil {
		return nil, err
	}
	documented, err := m.source.collection(ctx, workspace.Path, id)
	if err != nil {
		return nil, err
	}
	documentation, err := m.source.documentation(ctx, *documented)
	if err != nil {
		return nil, err
	}
	result := entity.DocumentationResult{Plan: *written, Documentation: *documentation}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "id": id, "text": text, "confirmDeletions": confirmDeletions, "keepDeleted": keepDeleted, "result": result})
	return &result, nil
}
