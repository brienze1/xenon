package model

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/yamlnode"
)

const (
	postmanKey = "postman"
	brunoKey   = "bruno"
)

func sourceFieldsToEntity(fields map[string]any) entity.SourceFields {
	if len(fields) == 0 {
		return nil
	}
	return fields
}

func sourceFieldsReserved(key string, fields map[string]any) yamlnode.Field {
	return yamlnode.Reserved(key, yamlnode.Any(fields))
}

func optionalSourceFields(key string, fields map[string]any) []yamlnode.Field {
	if len(fields) == 0 {
		return []yamlnode.Field{}
	}
	return []yamlnode.Field{sourceFieldsReserved(key, fields)}
}
