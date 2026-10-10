package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/idutils"
	"github.com/brienze1/xenon/pkg/logger"
)

type manageExample struct {
	findWorkspaceRegistry usecase.FindWorkspaceRegistry
	findRequest           usecase.FindRequest
	saveRequestExamples   usecase.SaveRequestExamples
	findResponse          usecase.FindResponse
	saveResponse          usecase.SaveResponse
}

func NewManageExampleService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	findRequest usecase.FindRequest,
	saveRequestExamples usecase.SaveRequestExamples,
	findResponse usecase.FindResponse,
	saveResponse usecase.SaveResponse,
) adapter.ManageExampleService {
	return &manageExample{
		findWorkspaceRegistry: findWorkspaceRegistry,
		findRequest:           findRequest,
		saveRequestExamples:   saveRequestExamples,
		findResponse:          findResponse,
		saveResponse:          saveResponse,
	}
}

func (m *manageExample) Get(ctx context.Context, localId, requestId, exampleId string) (*entity.ExampleDetails, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "requestId": requestId, "exampleId": exampleId})

	loaded, err := m.load(ctx, localId, requestId)
	if err != nil {
		return nil, err
	}
	example, err := usecase.ResolveExample(loaded.request, exampleId)
	if err != nil {
		return nil, err
	}
	details, err := m.served(ctx, loaded.request, *example)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "requestId": requestId, "exampleId": exampleId, "result": *details})
	return details, nil
}

func (m *manageExample) Create(ctx context.Context, localId, requestId, name string) (*entity.ExampleDetails, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "requestId": requestId, "name": name})

	loaded, err := m.load(ctx, localId, requestId)
	if err != nil {
		return nil, err
	}
	created := entity.NewExampleOf(idutils.NewUlid(), name, loaded.request)
	details, err := m.saveNamed(ctx, *loaded, created, loaded.request.WithExampleAppended(created))
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "requestId": requestId, "name": name, "result": *details})
	return details, nil
}

func (m *manageExample) CreateFromResponse(ctx context.Context, localId, requestId, tabId, name string) (*entity.ExampleDetails, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "requestId": requestId, "tabId": tabId, "name": name})

	loaded, err := m.load(ctx, localId, requestId)
	if err != nil {
		return nil, err
	}
	response, body, err := m.received(ctx, tabId)
	if err != nil {
		return nil, err
	}
	created := entity.NewExampleFromExchange(idutils.NewUlid(), name, response.SentRequestOr(loaded.request), *response, body)
	details, err := m.saveNamed(ctx, *loaded, created, loaded.request.WithExampleAppended(created))
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "requestId": requestId, "tabId": tabId, "name": name, "result": *details})
	return details, nil
}

func (m *manageExample) Save(ctx context.Context, localId, requestId string, example entity.SavedExample, keepBody bool, baseRevision string, force bool) (*entity.ExampleDetails, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "requestId": requestId, "example": example, "keepBody": keepBody, "baseRevision": baseRevision, "force": force})

	workspace, err := m.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return nil, err
	}
	current, err := m.findRequest.FindRevision(ctx, workspace.Path, requestId)
	if err != nil {
		return nil, err
	}
	if current.BlocksSave(baseRevision, force) {
		return nil, usecase.RejectStaleExampleSave()
	}
	loaded, err := m.load(ctx, localId, requestId)
	if err != nil {
		return nil, err
	}
	stored, err := usecase.ResolveExample(loaded.request, example.ID)
	if err != nil {
		return nil, err
	}
	edited := stored.Edited(example, keepBody)
	if !edited.Response.HasValidCode() {
		return nil, usecase.RejectInvalidExampleCode()
	}
	details, err := m.saveNamed(ctx, *loaded, edited, loaded.request.WithExample(edited))
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "requestId": requestId, "example": example, "keepBody": keepBody, "baseRevision": baseRevision, "force": force, "result": *details})
	return details, nil
}

func (m *manageExample) Rename(ctx context.Context, localId, requestId, exampleId, name string) (*entity.ExampleDetails, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "requestId": requestId, "exampleId": exampleId, "name": name})

	loaded, err := m.load(ctx, localId, requestId)
	if err != nil {
		return nil, err
	}
	stored, err := usecase.ResolveExample(loaded.request, exampleId)
	if err != nil {
		return nil, err
	}
	renamed := stored.Renamed(name)
	details, err := m.saveNamed(ctx, *loaded, renamed, loaded.request.WithExample(renamed))
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "requestId": requestId, "exampleId": exampleId, "name": name, "result": *details})
	return details, nil
}

func (m *manageExample) Duplicate(ctx context.Context, localId, requestId, exampleId string) (*entity.ExampleDetails, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "requestId": requestId, "exampleId": exampleId})

	loaded, err := m.load(ctx, localId, requestId)
	if err != nil {
		return nil, err
	}
	stored, err := usecase.ResolveExample(loaded.request, exampleId)
	if err != nil {
		return nil, err
	}
	copied := stored.Duplicated(idutils.NewUlid())
	details, err := m.saveNamed(ctx, *loaded, copied, loaded.request.WithExampleAfter(exampleId, copied))
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "requestId": requestId, "exampleId": exampleId, "result": *details})
	return details, nil
}

func (m *manageExample) Delete(ctx context.Context, localId, requestId, exampleId string) error {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "requestId": requestId, "exampleId": exampleId})

	loaded, err := m.load(ctx, localId, requestId)
	if err != nil {
		return err
	}
	if _, err := usecase.ResolveExample(loaded.request, exampleId); err != nil {
		return err
	}
	if _, err := m.saveRequestExamples.SaveExamples(ctx, loaded.root, loaded.request.WithoutExample(exampleId)); err != nil {
		return err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "requestId": requestId, "exampleId": exampleId})
	return nil
}

func (m *manageExample) load(ctx context.Context, localId, requestId string) (*loadedExamples, error) {
	workspace, err := m.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return nil, err
	}
	request, err := m.findRequest.Find(ctx, workspace.Path, requestId)
	if err != nil {
		return nil, err
	}
	return &loadedExamples{root: workspace.Path, request: *request}, nil
}

func (m *manageExample) received(ctx context.Context, tabId string) (*entity.Response, []byte, error) {
	response, err := m.findResponse.Find(ctx, tabId)
	if errs.IsNotFoundError(err) {
		return nil, nil, usecase.RejectMissingTabResponse()
	}
	if err != nil {
		return nil, nil, err
	}
	if !response.IsOk() {
		return nil, nil, usecase.RejectMissingTabResponse()
	}
	content, err := m.findResponse.OpenBody(ctx, tabId)
	if err != nil {
		return nil, nil, err
	}
	body, err := usecase.ReadResponseContent(content)
	if err != nil {
		return nil, nil, err
	}
	return response, body, nil
}

func (m *manageExample) saveNamed(ctx context.Context, loaded loadedExamples, example entity.SavedExample, updated entity.Request) (*entity.ExampleDetails, error) {
	if !example.HasValidName() {
		return nil, usecase.RejectInvalidExampleName()
	}
	saved, err := m.saveRequestExamples.SaveExamples(ctx, loaded.root, updated)
	if err != nil {
		return nil, err
	}
	stored, err := usecase.ResolveExample(*saved, example.ID)
	if err != nil {
		return nil, err
	}
	return m.served(ctx, *saved, *stored)
}

func (m *manageExample) served(ctx context.Context, request entity.Request, example entity.SavedExample) (*entity.ExampleDetails, error) {
	details := entity.NewExampleDetails(request, example)
	if details.ServesBodyByRoute() {
		if err := m.saveResponse.Save(ctx, example.BodyKey(), example.Response.AsResponse()); err != nil {
			return nil, err
		}
	}
	return &details, nil
}
