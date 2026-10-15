package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

const importPickerTitle = "Choose files, folders or zips to import"

type manageImport struct {
	readImportSource    usecase.ReadImportSource
	saveImportSource    usecase.SaveImportSource
	findImportSource    usecase.FindImportSource
	discardImportSource usecase.DiscardImportSource
	pickPath            usecase.PickPath
	secrets             importSecrets
	readers             importReaders
	targets             importTargets
	writer              importWriter
}

func NewManageImportService(
	readImportSource usecase.ReadImportSource,
	saveImportSource usecase.SaveImportSource,
	findImportSource usecase.FindImportSource,
	discardImportSource usecase.DiscardImportSource,
	readPostmanCollection usecase.ReadPostmanCollection,
	readPostmanEnvironment usecase.ReadPostmanEnvironment,
	readPostmanDump usecase.ReadPostmanDump,
	readPostmanV3 usecase.ReadPostmanV3,
	readBrunoCollection usecase.ReadBrunoCollection,
	readCurlCommand usecase.ReadCurlCommand,
	readXenonExport usecase.ReadXenonExport,
	decryptExportSecrets usecase.DecryptExportSecrets,
	pickPath usecase.PickPath,
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	loadTree usecase.LoadTree,
	findTreeItem usecase.FindTreeItem,
	listEnvironments usecase.ListEnvironments,
	loadWorkspace usecase.LoadWorkspace,
	inspectWorkspaceFolder usecase.InspectWorkspaceFolder,
	copyWorkspaceFolder usecase.CopyWorkspaceFolder,
	removeWorkspaceFolder usecase.RemoveWorkspaceFolder,
	saveWorkspaceRegistry usecase.SaveWorkspaceRegistry,
	stageImportItem usecase.StageImportItem,
	commitImportItem usecase.CommitImportItem,
	rollbackImportItem usecase.RollbackImportItem,
	saveEnvironment usecase.SaveEnvironment,
	deleteEnvironment usecase.DeleteEnvironment,
	saveWorkspaceFile usecase.SaveWorkspaceFile,
	findLocalVariables usecase.FindLocalVariables,
	saveLocalVariables usecase.SaveLocalVariables,
	findVault usecase.FindVault,
	saveVault usecase.SaveVault,
	deleteVault usecase.DeleteVault,
	findVaultSettings usecase.FindVaultSettings,
	saveVaultSettings usecase.SaveVaultSettings,
	probeKeychain usecase.ProbeKeychain,
	publishImportProgress usecase.PublishImportProgress,
	publishWorkspaceEvent usecase.PublishWorkspaceEvent,
	publishVariableEvent usecase.PublishVariableEvent,
) adapter.ManageImportService {
	vaults := vaultWriter{
		findVault:            findVault,
		saveVault:            saveVault,
		findVaultSettings:    findVaultSettings,
		saveVaultSettings:    saveVaultSettings,
		probeKeychain:        probeKeychain,
		publishVariableEvent: publishVariableEvent,
	}
	return &manageImport{
		readImportSource:    readImportSource,
		saveImportSource:    saveImportSource,
		findImportSource:    findImportSource,
		discardImportSource: discardImportSource,
		pickPath:            pickPath,
		secrets:             importSecrets{decryptExportSecrets: decryptExportSecrets},
		readers: importReaders{
			readPostmanCollection:  readPostmanCollection,
			readPostmanEnvironment: readPostmanEnvironment,
			readPostmanDump:        readPostmanDump,
			readPostmanV3:          readPostmanV3,
			readBrunoCollection:    readBrunoCollection,
			readCurlCommand:        readCurlCommand,
			readXenonExport:        readXenonExport,
		},
		targets: importTargets{
			findWorkspaceRegistry:  findWorkspaceRegistry,
			loadTree:               loadTree,
			findTreeItem:           findTreeItem,
			listEnvironments:       listEnvironments,
			loadWorkspace:          loadWorkspace,
			inspectWorkspaceFolder: inspectWorkspaceFolder,
		},
		writer: importWriter{
			stageImportItem:       stageImportItem,
			commitImportItem:      commitImportItem,
			rollbackImportItem:    rollbackImportItem,
			publishImportProgress: publishImportProgress,
			publishWorkspaceEvent: publishWorkspaceEvent,
			workspaces: importWorkspaceWriter{
				copyWorkspaceFolder:   copyWorkspaceFolder,
				removeWorkspaceFolder: removeWorkspaceFolder,
				findWorkspaceRegistry: findWorkspaceRegistry,
				saveWorkspaceRegistry: saveWorkspaceRegistry,
				loadWorkspace:         loadWorkspace,
				deleteVault:           deleteVault,
				vaults:                vaults,
			},
			variables: importVariablesWriter{
				saveEnvironment:   saveEnvironment,
				deleteEnvironment: deleteEnvironment,
				saveWorkspaceFile: saveWorkspaceFile,
				loadWorkspace:     loadWorkspace,
				orders: environmentOrders{
					listEnvironments:  listEnvironments,
					loadWorkspace:     loadWorkspace,
					saveWorkspaceFile: saveWorkspaceFile,
				},
				values: scopeValues{
					findLocalVariables: findLocalVariables,
					saveLocalVariables: saveLocalVariables,
					findVault:          findVault,
					saveVault:          saveVault,
				},
				vaults:               vaults,
				publishVariableEvent: publishVariableEvent,
			},
		},
	}
}

func (m *manageImport) Detect(ctx context.Context, request entity.ImportRequest) (*entity.ImportSource, error) {
	logger.Info(ctx, "Started", map[string]any{"request": request})

	inputs, tempDirs, err := m.readImportSource.Read(ctx, request)
	if err != nil {
		return nil, err
	}
	source, err := m.readers.detected(ctx, inputs, tempDirs)
	if err != nil {
		if discardErr := m.discardImportSource.DiscardFiles(ctx, tempDirs); discardErr != nil {
			logger.Error(ctx, discardErr, "The unread import files could not be discarded", map[string]any{"tempDirs": tempDirs})
		}
		return nil, err
	}
	if err := m.saveImportSource.Save(ctx, *source); err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"request": request, "result": *source})
	return source, nil
}

func (m *manageImport) PickSources(ctx context.Context) ([]string, error) {
	logger.Info(ctx, "Started")

	paths, err := m.pickPath.PickSources(ctx, importPickerTitle)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"paths": paths})
	return paths, nil
}

func (m *manageImport) Clashes(ctx context.Context, plan entity.ImportPlan) ([]entity.Clash, error) {
	logger.Info(ctx, "Started", map[string]any{"plan": plan})

	source, err := m.findImportSource.Find(ctx, plan.SourceID)
	if err != nil {
		return nil, err
	}
	targets, err := m.targets.resolve(ctx, *source, plan)
	if err != nil {
		return nil, err
	}
	clashes := make([]entity.Clash, 0)
	for _, target := range targets {
		clashes = append(clashes, target.clashes()...)
	}

	logger.Info(ctx, "Finished", map[string]any{"plan": plan, "clashes": len(clashes)})
	return clashes, nil
}

func (m *manageImport) Import(ctx context.Context, plan entity.ImportPlan) (*entity.ImportResult, error) {
	logger.Info(ctx, "Started", map[string]any{"plan": plan})

	found, err := m.findImportSource.Find(ctx, plan.SourceID)
	if err != nil {
		return nil, err
	}
	source, err := m.secrets.unlocked(ctx, *found, plan)
	if err != nil {
		return nil, err
	}
	targets, err := m.targets.resolve(ctx, *source, plan)
	if err != nil {
		return nil, err
	}
	if clashes := unresolvedClashes(plan, targets); len(clashes) > 0 {
		return nil, usecase.RejectUnresolvedClashes(clashes)
	}
	if ctx.Err() != nil {
		return nil, usecase.RejectCancelledImport()
	}
	result := m.writer.write(ctx, source.ID, plan, targets).WithSourceWarnings(source.SourceWarnings)

	logger.Info(ctx, "Finished", map[string]any{"plan": plan, "result": result})
	return &result, nil
}

func (m *manageImport) Discard(ctx context.Context, sourceId string) error {
	logger.Info(ctx, "Started", map[string]any{"sourceId": sourceId})

	if err := m.discardImportSource.Discard(ctx, sourceId); err != nil {
		return err
	}

	logger.Info(ctx, "Finished", map[string]any{"sourceId": sourceId})
	return nil
}

func unresolvedClashes(plan entity.ImportPlan, targets []importTarget) []entity.Clash {
	clashes := make([]entity.Clash, 0)
	for _, target := range targets {
		if !plan.Resolves(target.planned) {
			clashes = append(clashes, target.clashes()...)
		}
	}
	return clashes
}
