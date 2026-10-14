package dto

import (
	"bytes"
	"encoding/json"
	"path"
	"strings"

	"go.yaml.in/yaml/v3"

	"github.com/brienze1/xenon/internal/domain/entity"
)

const jsonExtension = ".json"

func Decode(document entity.ImportDocument, target any) error {
	data := document.Data
	isJson := isJsonDocument(document)
	if isJson {
		var value any
		if err := json.Unmarshal(data, &value); err != nil {
			return NewDecodeError(document.Path, err)
		}
		converted, err := yaml.Marshal(value)
		if err != nil {
			return NewDecodeError(document.Path, err)
		}
		data = converted
	}
	if err := yaml.Unmarshal(data, target); err != nil {
		if isJson {
			return NewWrongTypeDecodeError(document.Path)
		}
		return NewDecodeError(document.Path, err)
	}
	return nil
}

func isJsonDocument(document entity.ImportDocument) bool {
	trimmed := bytes.TrimSpace(document.Data)
	return strings.EqualFold(path.Ext(document.Path), jsonExtension) || bytes.HasPrefix(trimmed, []byte("{"))
}
