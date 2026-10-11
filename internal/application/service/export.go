package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/exportformat"
	"github.com/brienze1/xenon/pkg/fileutils"
	"github.com/brienze1/xenon/pkg/logger"
)

const (
	exportPickerTitle        = "Export as Xenon YAML"
	postmanExportPickerTitle = "Export as Postman"
)

type exportDocuments struct {
	entries []entity.ExportEntry
	names   []string
	summary entity.ExportSummary
}

type export struct {
	findWorkspaceRegistry usecase.FindWorkspaceRegistry
	pickPath              usecase.PickPath
	showInFolder          usecase.ShowInFolder
	saveExportFile        usecase.SaveExportFile
	publishExportProgress usecase.PublishExportProgress
	items                 exportItems
	postman               postmanExport
	xenon                 xenonExport
}

func NewExportService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	loadTree usecase.LoadTree,
	findTreeItem usecase.FindTreeItem,
	findRequest usecase.FindRequest,
	listEnvironments usecase.ListEnvironments,
	findEnvironment usecase.FindEnvironment,
	loadWorkspace usecase.LoadWorkspace,
	findLocalVariables usecase.FindLocalVariables,
	writePostmanExport usecase.WritePostmanExport,
	writeXenonExport usecase.WriteXenonExport,
	encryptExportSecrets usecase.EncryptExportSecrets,
	findVault usecase.FindVault,
	findVaultSettings usecase.FindVaultSettings,
	saveExportFile usecase.SaveExportFile,
	pickPath usecase.PickPath,
	showInFolder usecase.ShowInFolder,
	publishExportProgress usecase.PublishExportProgress,
) adapter.ExportService {
	items := exportItems{
		loadTree:           loadTree,
		findTreeItem:       findTreeItem,
		findRequest:        findRequest,
		listEnvironments:   listEnvironments,
		findEnvironment:    findEnvironment,
		loadWorkspace:      loadWorkspace,
		findLocalVariables: findLocalVariables,
	}
	return &export{
		findWorkspaceRegistry: findWorkspaceRegistry,
		pickPath:              pickPath,
		showInFolder:          showInFolder,
		saveExportFile:        saveExportFile,
		publishExportProgress: publishExportProgress,
		items:                 items,
		postman:               postmanExport{writePostmanExport: writePostmanExport, items: items},
		xenon: xenonExport{
			writeXenonExport:     writeXenonExport,
			encryptExportSecrets: encryptExportSecrets,
			findVault:            findVault,
			findVaultSettings:    findVaultSettings,
			items:                items,
		},
	}
}

func (e *export) Options(ctx context.Context, localId string, kind enums.ExportItemKind, itemId string) (*entity.ExportOptions, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "kind": kind, "itemId": itemId})

	workspace, err := e.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return nil, err
	}
	name, scopes, err := e.items.scopes(ctx, workspace.Path, kind, itemId)
	if err != nil {
		return nil, err
	}
	local, err := e.items.local(ctx, localId)
	if err != nil {
		return nil, err
	}
	options := entity.NewExportOptions(kind, fileutils.SafeName(name), scopes, *local)

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "kind": kind, "itemId": itemId, "result": options})
	return &options, nil
}

func (e *export) PickTarget(ctx context.Context, suggestedFileName string, format enums.ExportFormat, shape enums.ExportShape) (string, error) {
	logger.Info(ctx, "Started", map[string]any{"suggestedFileName": suggestedFileName, "format": format, "shape": shape})

	title := exportPickerTitle
	if format == exportformat.Postman {
		title = postmanExportPickerTitle
	}
	path, err := e.pickPath.PickSaveFile(ctx, title, entity.ExportFileNameForShape(suggestedFileName, format, shape))
	if err != nil {
		return "", err
	}

	logger.Info(ctx, "Finished", map[string]any{"suggestedFileName": suggestedFileName, "format": format, "shape": shape, "path": path})
	return path, nil
}

func (e *export) Export(ctx context.Context, request entity.ExportRequest) (*entity.ExportSummary, error) {
	logger.Info(ctx, "Started", request)

	if !request.HasValidCombination() {
		return nil, usecase.RejectInvalidExportCombination()
	}
	if !request.HasValidPassphrase() {
		return nil, usecase.RejectInvalidExportPassphrase()
	}
	workspace, err := e.findWorkspaceRegistry.FindWorkspace(ctx, request.LocalID)
	if err != nil {
		return nil, err
	}
	documents, err := e.documents(ctx, workspace.Path, request)
	if err != nil {
		return nil, err
	}
	if err := e.save(ctx, request, documents.entries); err != nil {
		return nil, err
	}
	summary := documents.summary

	logger.Info(ctx, "Finished", request, map[string]any{"result": summary})
	return &summary, nil
}

func (e *export) ShowInFolder(ctx context.Context, path string) error {
	logger.Info(ctx, "Started", map[string]any{"path": path})

	if err := e.showInFolder.ShowInFolder(ctx, path); err != nil {
		return err
	}

	logger.Info(ctx, "Finished", map[string]any{"path": path})
	return nil
}

func (e *export) documents(ctx context.Context, root string, request entity.ExportRequest) (*exportDocuments, error) {
	if request.IsPostman() {
		return e.postman.documents(ctx, root, request)
	}
	return e.xenon.documents(ctx, root, request)
}

func (e *export) save(ctx context.Context, request entity.ExportRequest, entries []entity.ExportEntry) error {
	total := len(entries)
	if err := e.publishExportProgress.Progress(ctx, entity.ExportProgress{Done: 0, Total: total}); err != nil {
		return err
	}
	var err error
	if request.IsArchive() {
		err = e.saveExportFile.SaveArchive(ctx, request.TargetPath, entries, func(done int) {
			e.publishMilestone(ctx, entity.ExportProgress{Done: done, Total: total})
		})
	} else {
		err = e.saveExportFile.SaveFile(ctx, request.TargetPath, entries[0].Data)
	}
	if err != nil {
		return err
	}
	return e.publishExportProgress.Progress(ctx, entity.ExportProgress{Done: total, Total: total})
}

func (e *export) publishMilestone(ctx context.Context, progress entity.ExportProgress) {
	if !progress.IsMilestone() {
		return
	}
	if err := e.publishExportProgress.Progress(ctx, progress); err != nil {
		logger.Error(ctx, err, "The export progress could not be announced", map[string]any{"progress": progress})
	}
}
