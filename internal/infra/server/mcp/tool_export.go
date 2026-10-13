package mcp

import (
	"context"
	"slices"
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums/exportformat"
	"github.com/brienze1/xenon/internal/domain/enums/exportshape"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

const (
	xenonExportFormat = "xenon"
	zipExtension      = ".zip"
)

type exportTarget struct {
	Kind string `json:"kind"`
	Id   string `json:"id"`
}

type exportArguments struct {
	Target           exportTarget `json:"target"`
	Format           string       `json:"format"`
	Path             string       `json:"path"`
	CurrentAsInitial bool         `json:"current_as_initial"`
	IncludeSecrets   bool         `json:"include_secrets"`
}

func (t *tools) export(ctx context.Context, call Call) (Result, error) {
	switch call.Action {
	case "options":
		return t.exportOptions(ctx, call)
	case "run":
		return t.runExport(ctx, call)
	default:
		return t.saveResponseBody(ctx, call)
	}
}

func (t *tools) exportOptions(ctx context.Context, call Call) (Result, error) {
	var arguments exportArguments
	if err := call.Decode(&arguments); err != nil {
		return Result{}, err
	}
	options, err := t.controllers.Export.Options(ctx, arguments.reference(call.LocalId))
	if err != nil {
		return Result{}, err
	}
	return Result{Payload: options, Target: arguments.activityTarget()}, nil
}

func (t *tools) runExport(ctx context.Context, call Call) (Result, error) {
	var arguments exportArguments
	if err := call.Decode(&arguments); err != nil {
		return Result{}, err
	}
	if arguments.IncludeSecrets {
		return Result{}, humanOnly("Exporting secrets")
	}
	options, err := t.controllers.Export.Options(ctx, arguments.reference(call.LocalId))
	if err != nil {
		return Result{}, err
	}
	format := arguments.exportFormat()
	summary, err := t.controllers.Export.Export(ctx, dto.Export{
		LocalId:                call.LocalId,
		ItemKind:               arguments.Target.Kind,
		ItemId:                 arguments.Target.Id,
		Format:                 format,
		Shape:                  exportShape(*options, format, arguments.Path),
		TargetPath:             arguments.Path,
		CurrentValuesAsInitial: arguments.CurrentAsInitial,
	})
	if err != nil {
		return Result{}, err
	}
	return Result{Payload: summary, Target: arguments.activityTarget()}, nil
}

func (e exportArguments) reference(localId string) dto.ExportItemReference {
	return dto.ExportItemReference{LocalId: localId, ItemKind: e.Target.Kind, ItemId: e.Target.Id}
}

func (e exportArguments) activityTarget() dto.AIActivityTarget {
	return dto.AIActivityTarget{Id: e.Target.Id, Path: e.Path}
}

func (e exportArguments) exportFormat() string {
	if e.Format == xenonExportFormat {
		return exportformat.XenonYaml.String()
	}
	return e.Format
}

func exportShape(options dto.ExportOptions, format, path string) string {
	index := slices.IndexFunc(options.Formats, func(option dto.ExportFormatOption) bool { return option.Format == format })
	if index < 0 || len(options.Formats[index].Shapes) == 0 {
		return exportshape.File.String()
	}
	shapes := options.Formats[index].Shapes
	wantsZip := strings.HasSuffix(strings.ToLower(path), zipExtension)
	for _, shape := range shapes {
		if (shape == exportshape.Zip.String()) == wantsZip {
			return shape
		}
	}
	return shapes[0]
}
