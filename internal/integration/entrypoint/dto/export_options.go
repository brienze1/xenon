package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ExportOptions struct {
	SuggestedFileName string               `json:"suggested_file_name"`
	Formats           []ExportFormatOption `json:"formats"`
	HasSecrets        bool                 `json:"has_secrets"`
	HasCurrentValues  bool                 `json:"has_current_values"`
}

func NewExportOptions(options entity.ExportOptions) ExportOptions {
	formats := make([]ExportFormatOption, 0, len(options.Formats))
	for _, option := range options.Formats {
		formats = append(formats, NewExportFormatOption(option))
	}
	return ExportOptions{SuggestedFileName: options.SuggestedFileName, Formats: formats, HasSecrets: options.HasSecrets, HasCurrentValues: options.HasCurrentValues}
}
