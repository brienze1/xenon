package entity

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/exportformat"
	"github.com/brienze1/xenon/internal/domain/enums/exportitemkind"
	"github.com/brienze1/xenon/internal/domain/enums/exportshape"
)

const (
	postmanZipSuffix         = "-postman.zip"
	postmanCollectionSuffix  = ".postman_collection.json"
	postmanEnvironmentSuffix = ".postman_environment.json"
	postmanGlobalsSuffix     = ".postman_globals.json"
	xenonZipSuffix           = ".zip"
	xenonYamlSuffix          = ".yaml"
)

type ExportFormatOption struct {
	Format                enums.ExportFormat
	SuggestedFileName     string
	Shapes                []enums.ExportShape
	IncludeSecretsAllowed bool
}

func ExportFormatOptions(kind enums.ExportItemKind, baseName string) []ExportFormatOption {
	options := make([]ExportFormatOption, 0, 2)
	for _, format := range []enums.ExportFormat{exportformat.XenonYaml, exportformat.Postman} {
		if option, found := ExportFormatOptionOf(kind, format); found {
			option.SuggestedFileName = ExportFileName(baseName, kind, format, option.Shapes[0])
			options = append(options, option)
		}
	}
	return options
}

func ExportFormatOptionOf(kind enums.ExportItemKind, format enums.ExportFormat) (ExportFormatOption, bool) {
	shapes := exportShapes(kind, format)
	if len(shapes) == 0 {
		return ExportFormatOption{}, false
	}
	return ExportFormatOption{Format: format, Shapes: shapes, IncludeSecretsAllowed: format == exportformat.XenonYaml}, true
}

func ExportFileName(baseName string, kind enums.ExportItemKind, format enums.ExportFormat, shape enums.ExportShape) string {
	if format == exportformat.XenonYaml {
		if shape == exportshape.Zip {
			return baseName + xenonZipSuffix
		}
		return baseName + xenonYamlSuffix
	}
	switch kind {
	case exportitemkind.Workspace:
		return baseName + postmanZipSuffix
	case exportitemkind.Environment:
		return baseName + postmanEnvironmentSuffix
	case exportitemkind.Globals:
		return baseName + postmanGlobalsSuffix
	default:
		return baseName + postmanCollectionSuffix
	}
}

func ExportFileNameForShape(name string, format enums.ExportFormat, shape enums.ExportShape) string {
	if format != exportformat.XenonYaml || (shape != exportshape.Zip && shape != exportshape.Bundle) {
		return name
	}
	base := strings.TrimSuffix(strings.TrimSuffix(name, xenonZipSuffix), xenonYamlSuffix)
	return ExportFileName(base, exportitemkind.Unknown, format, shape)
}

func exportShapes(kind enums.ExportItemKind, format enums.ExportFormat) []enums.ExportShape {
	switch {
	case kind == exportitemkind.Workspace && (format == exportformat.XenonYaml || format == exportformat.Postman):
		return []enums.ExportShape{exportshape.Zip}
	case (kind == exportitemkind.Collection || kind == exportitemkind.Folder) && format == exportformat.XenonYaml:
		return []enums.ExportShape{exportshape.Zip, exportshape.Bundle}
	case kind == exportitemkind.Globals && format == exportformat.Postman:
		return []enums.ExportShape{exportshape.File}
	case kind == exportitemkind.Globals || kind == exportitemkind.Unknown:
		return nil
	case format == exportformat.XenonYaml || format == exportformat.Postman:
		return []enums.ExportShape{exportshape.File}
	default:
		return nil
	}
}
