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
	for _, pair := range mappingPairs(attributes) {
		if !hasYamlField(typed, pair[0].Value) {
			fields = append(fields, yamlnode.Field{Key: pair[0].Value, Value: pair[1]})
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
	for _, pair := range mappingPairs(node) {
		if pair[0].Value == key {
			return pair[1]
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

func mappingPairs(node *yaml.Node) [][2]*yaml.Node {
	pairs := make([][2]*yaml.Node, 0)
	for index := range len(node.Content) / 2 {
		pairs = append(pairs, [2]*yaml.Node{node.Content[2*index], node.Content[2*index+1]})
	}
	return pairs
}
