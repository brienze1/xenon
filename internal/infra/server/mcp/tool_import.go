package mcp

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/enums/aichangekind"
	"github.com/brienze1/xenon/internal/domain/enums/clashresolution"
	"github.com/brienze1/xenon/internal/domain/enums/importitemkind"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/pkg/logger"
)

type importArguments struct {
	SourceId   string   `json:"source_id"`
	OnClash    string   `json:"on_clash"`
	Passphrase string   `json:"passphrase"`
	Paths      []string `json:"paths"`
}

func (t *tools) importFiles(ctx context.Context, call Call) (Result, error) {
	var arguments importArguments
	if err := call.Decode(&arguments); err != nil {
		return Result{}, err
	}
	switch call.Action {
	case "detect":
		return t.detectImport(ctx, arguments)
	case "clashes":
		return t.importClashes(ctx, call, arguments)
	case "run":
		return t.runImport(ctx, call, arguments)
	default:
		return t.discardImport(ctx, arguments)
	}
}

func (t *tools) detectImport(ctx context.Context, arguments importArguments) (Result, error) {
	preview, err := t.controllers.Import.Detect(ctx, dto.DetectImport{Paths: arguments.Paths})
	if err != nil {
		return Result{}, err
	}
	return Result{Payload: preview}, nil
}

func (t *tools) importClashes(ctx context.Context, call Call, arguments importArguments) (Result, error) {
	preview, err := t.controllers.Import.Detect(ctx, dto.DetectImport{Paths: arguments.Paths})
	if err != nil {
		return Result{}, err
	}
	defer t.discardImportSource(ctx, preview.SourceId)
	clashes, err := t.controllers.Import.Clashes(ctx, importPlan(call.LocalId, *preview, arguments))
	if err != nil {
		return Result{}, err
	}
	return Result{Payload: List[dto.ImportClash]{Items: clashes.Clashes}}, nil
}

func (t *tools) runImport(ctx context.Context, call Call, arguments importArguments) (Result, error) {
	if arguments.OnClash == "" {
		arguments.OnClash = clashresolution.KeepBoth.String()
	}
	preview, err := t.controllers.Import.Detect(ctx, dto.DetectImport{Paths: arguments.Paths})
	if err != nil {
		return Result{}, err
	}
	defer t.discardImportSource(ctx, preview.SourceId)
	imported, err := t.controllers.Import.Import(ctx, importPlan(call.LocalId, *preview, arguments))
	if err != nil {
		return Result{}, err
	}
	return Result{Payload: imported, Changes: []dto.AIChange{{Kind: aichangekind.Tree.String()}, {Kind: aichangekind.Variables.String()}}}, nil
}

func (t *tools) discardImport(ctx context.Context, arguments importArguments) (Result, error) {
	if arguments.SourceId == "" {
		return Result{Payload: struct{}{}}, nil
	}
	if err := t.controllers.Import.Discard(ctx, dto.ImportSourceReference{SourceId: arguments.SourceId}); err != nil {
		return Result{}, err
	}
	return Result{Payload: struct{}{}}, nil
}

func (t *tools) discardImportSource(ctx context.Context, sourceId string) {
	if err := t.controllers.Import.Discard(ctx, dto.ImportSourceReference{SourceId: sourceId}); err != nil {
		logger.Warn(ctx, err, "The import source could not be discarded", map[string]any{"sourceId": sourceId})
	}
}

func importPlan(localId string, preview dto.ImportPreview, arguments importArguments) dto.ImportPlan {
	items := make([]dto.ImportPlanItem, 0, len(preview.Items))
	for _, item := range preview.Items {
		items = append(items, dto.ImportPlanItem{Key: item.Key, TargetLocalId: localId, Resolution: arguments.OnClash, Include: item.Kind != importitemkind.Workspace.String()})
	}
	return dto.ImportPlan{SourceId: preview.SourceId, ApplyToAll: arguments.OnClash, Passphrase: arguments.Passphrase, Items: items}
}
