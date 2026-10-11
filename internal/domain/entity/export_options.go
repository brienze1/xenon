package entity

import "github.com/brienze1/xenon/internal/domain/enums"

type ExportOptions struct {
	SuggestedFileName string
	Formats           []ExportFormatOption
	HasSecrets        bool
	HasCurrentValues  bool
}

func NewExportOptions(kind enums.ExportItemKind, baseName string, scopes []ExportScope, local LocalVariables) ExportOptions {
	options := ExportOptions{SuggestedFileName: baseName, Formats: ExportFormatOptions(kind, baseName)}
	for _, scope := range scopes {
		options.HasSecrets = options.HasSecrets || scope.HasSecrets()
		options.HasCurrentValues = options.HasCurrentValues || scope.HasCurrentValues(local)
	}
	return options
}
