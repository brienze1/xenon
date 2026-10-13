package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ExportFormatOption struct {
	Format                string   `json:"format"`
	SuggestedFileName     string   `json:"suggested_file_name"`
	Shapes                []string `json:"shapes"`
	IncludeSecretsAllowed bool     `json:"include_secrets_allowed"`
}

func NewExportFormatOption(option entity.ExportFormatOption) ExportFormatOption {
	shapes := make([]string, 0, len(option.Shapes))
	for _, shape := range option.Shapes {
		shapes = append(shapes, shape.String())
	}
	return ExportFormatOption{Format: option.Format.String(), SuggestedFileName: option.SuggestedFileName, Shapes: shapes, IncludeSecretsAllowed: option.IncludeSecretsAllowed}
}
