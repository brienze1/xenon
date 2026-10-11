package model

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/yamlnode"
)

func authExtraFields(attributes []entity.AuthAttribute) []yamlnode.Field {
	fields := make([]yamlnode.Field, 0, len(attributes))
	for _, attribute := range attributes {
		fields = append(fields, yamlnode.Field{Key: attribute.Key, Value: yamlnode.Any(attribute.Value)})
	}
	return fields
}
