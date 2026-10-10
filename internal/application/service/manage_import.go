package service

import (
	"context"
	"strconv"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/importitemkind"
	"github.com/brienze1/xenon/pkg/idutils"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/timeutils"
)

const importPickerTitle = "Choose files, folders or zips to import"

type manageImport struct {
	readImportSource      usecase.ReadImportSource
	saveImportSource      usecase.SaveImportSource
	findImportSource      usecase.FindImportSource
	discardImportSource   usecase.DiscardImportSource
	readPostmanCollection usecase.ReadPostmanCollection
	pickPath              usecase.PickPath
	targets               importTargets
	writer                importWriter
}

func NewManageImportService(
	readImportSource usecase.ReadImportSource,
	saveImportSource usecase.SaveImportSource,
	findImportSource usecase.FindImportSource,
	discardImportSource usecase.DiscardImportSource,
	readPostmanCollection usecase.ReadPostmanCollection,
	pickPath usecase.PickPath,
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	loadTree usecase.LoadTree,
	findTreeItem usecase.FindTreeItem,
	stageImportItem usecase.StageImportItem,
	commitImportItem usecase.CommitImportItem,
	rollbackImportItem usecase.RollbackImportItem,
	publishImportProgress usecase.PublishImportProgress,
	publishWorkspaceEvent usecase.PublishWorkspaceEvent,
) adapter.ManageImportService {
	return &manageImport{
		readImportSource:      readImportSource,
		saveImportSource:      saveImportSource,
		findImportSource:      findImportSource,
		discardImportSource:   discardImportSource,
		readPostmanCollection: readPostmanCollection,
		pickPath:              pickPath,
		targets:               importTargets{findWorkspaceRegistry: findWorkspaceRegistry, loadTree: loadTree, findTreeItem: findTreeItem},
		writer: importWriter{
			stageImportItem:       stageImportItem,
			commitImportItem:      commitImportItem,
			rollbackImportItem:    rollbackImportItem,
			publishImportProgress: publishImportProgress,
			publishWorkspaceEvent: publishWorkspaceEvent,
		},
	}
}

func (m *manageImport) Detect(ctx context.Context, request entity.ImportRequest) (*entity.ImportSource, error) {
	logger.Info(ctx, "Started", map[string]any{"paths": request.Paths, "hasText": request.Text != ""})

	inputs, tempDirs, err := m.readImportSource.Read(ctx, request)
	if err != nil {
		return nil, err
	}
	source, err := m.detected(ctx, inputs, tempDirs)
	if err != nil {
		_ = m.discardImportSource.DiscardFiles(ctx, tempDirs)
		return nil, err
	}
	if err := m.saveImportSource.Save(ctx, *source); err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"result": *source})
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

	source, err := m.findImportSource.Find(ctx, plan.SourceID)
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
	result := m.writer.write(ctx, source.ID, plan, targets)

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

func (m *manageImport) detected(ctx context.Context, inputs []entity.ImportInput, tempDirs []string) (*entity.ImportSource, error) {
	detected := make([]entity.ImportInput, 0, len(inputs))
	for _, input := range inputs {
		kind, version, err := usecase.DetectImportFormat(input.Probe)
		if err != nil {
			return nil, err
		}
		detected = append(detected, input.Detected(kind, version))
	}
	items := make([]entity.ImportItem, 0, len(detected))
	for _, input := range detected {
		read, err := m.items(ctx, input, len(items))
		if err != nil {
			return nil, err
		}
		items = append(items, read...)
	}
	source := entity.NewImportSource(idutils.NewUlid(), detected, tempDirs, timeutils.Now()).WithItems(items)
	return &source, nil
}

func (m *manageImport) items(ctx context.Context, input entity.ImportInput, offset int) ([]entity.ImportItem, error) {
	if !input.IsPostmanCollection() {
		return []entity.ImportItem{}, nil
	}
	root, warnings, err := m.readPostmanCollection.ReadCollection(ctx, input.Document())
	if err != nil {
		return nil, err
	}
	item := entity.NewImportItem(strconv.Itoa(offset), importitemkind.Collection, *root, warnings)
	usages := usecase.ScanPmUsage(item.Scripts())
	return []entity.ImportItem{item.WithPmUsages(usages).WithWarnings(usecase.PmUsageWarnings(usages))}, nil
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
