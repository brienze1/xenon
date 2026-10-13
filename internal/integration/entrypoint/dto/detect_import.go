package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type DetectImport struct {
	Text  string   `json:"text" validate:"import_input"`
	Paths []string `json:"paths" validate:"dive,absolute_path"`
}

func (d *DetectImport) ToEntity() entity.ImportRequest {
	return entity.ImportRequest{Paths: append([]string{}, d.Paths...), Text: d.Text}
}

func (d DetectImport) LogValue() any {
	return map[string]any{"paths": d.Paths, "hasText": d.Text != ""}
}
