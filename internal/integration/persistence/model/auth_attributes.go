package model

import (
	"reflect"
	"strings"

	"go.yaml.in/yaml/v3"

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

var typedAuthFields = map[string]reflect.Type{
	"bearer": reflect.TypeOf(AuthBearer{}),
	"basic":  reflect.TypeOf(AuthBasic{}),
	"apikey": reflect.TypeOf(AuthApiKey{}),
	"oauth2": reflect.TypeOf(AuthOAuth2{}),
	"awsv4":  reflect.TypeOf(AuthAwsV4{}),
	"digest": reflect.TypeOf(AuthDigest{}),
}

func (a *Auth) UnmarshalYAML(node *yaml.Node) error {
	type plain Auth
	if err := node.Decode((*plain)(a)); err != nil {
		return err
	}
	a.Extra = storedExtraFields(node, a.Type)
	return nil
}

func storedExtraFields(node *yaml.Node, authType string) []yamlnode.Field {
	fields := make([]yamlnode.Field, 0)
	typed, known := typedAuthFields[authType]
	attributes := mappingValue(node, authType)
	if !known || attributes == nil || attributes.Kind != yaml.MappingNode {
		return fields
	}
	for index := 0; index+1 < len(attributes.Content); index += 2 {
		key := attributes.Content[index].Value
		if !hasYamlField(typed, key) {
			fields = append(fields, yamlnode.Field{Key: key, Value: attributes.Content[index+1]})
		}
	}
	return fields
}

func authAttributesOf(fields []yamlnode.Field) []entity.AuthAttribute {
	attributes := make([]entity.AuthAttribute, 0, len(fields))
	for _, field := range fields {
		var value any
		if err := field.Value.Decode(&value); err == nil {
			attributes = append(attributes, entity.AuthAttribute{Key: field.Key, Value: value})
		}
	}
	return attributes
}

func mappingValue(node *yaml.Node, key string) *yaml.Node {
	if node.Kind != yaml.MappingNode {
		return nil
	}
	for index := 0; index+1 < len(node.Content); index += 2 {
		if node.Content[index].Value == key {
			return node.Content[index+1]
		}
	}
	return nil
}

func hasYamlField(typed reflect.Type, key string) bool {
	for index := 0; index < typed.NumField(); index++ {
		name, _, _ := strings.Cut(typed.Field(index).Tag.Get("yaml"), ",")
		if name == key {
			return true
		}
	}
	return false
}
