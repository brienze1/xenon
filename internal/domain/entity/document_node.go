package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/nodekind"
)

const rowKeyField = "key"

type DocumentNode struct {
	Fields map[string]*DocumentNode
	Kind   enums.NodeKind
	Tag    string
	Value  string
	Keys   []string
	Items  []*DocumentNode
}

func NewScalarNode(tag, value string) *DocumentNode {
	return &DocumentNode{Kind: nodekind.Scalar, Tag: tag, Value: value}
}

func NewSequenceNode(items []*DocumentNode) *DocumentNode {
	return &DocumentNode{Kind: nodekind.Sequence, Items: items}
}

func NewMappingNode() *DocumentNode {
	return &DocumentNode{Kind: nodekind.Mapping, Fields: map[string]*DocumentNode{}}
}

func (d *DocumentNode) Set(key string, value *DocumentNode) *DocumentNode {
	if _, exists := d.Fields[key]; !exists {
		d.Keys = append(d.Keys, key)
	}
	d.Fields[key] = value
	return d
}

func (d *DocumentNode) Field(key string) *DocumentNode {
	if d == nil {
		return nil
	}
	return d.Fields[key]
}

func (d *DocumentNode) IsMapping() bool {
	return d != nil && d.Kind == nodekind.Mapping
}

func (d *DocumentNode) Equals(other *DocumentNode) bool {
	if d == nil || other == nil {
		return d == other
	}
	if d.Kind != other.Kind {
		return false
	}
	switch d.Kind {
	case nodekind.Mapping:
		return d.sameFields(other)
	case nodekind.Sequence:
		return sameItems(d.Items, other.Items)
	default:
		return d.Tag == other.Tag && d.Value == other.Value
	}
}

func (d *DocumentNode) ScalarItems() ([]string, bool) {
	if d == nil || d.Kind != nodekind.Sequence {
		return nil, false
	}
	values := make([]string, 0, len(d.Items))
	for _, item := range d.Items {
		if item.Kind != nodekind.Scalar {
			return nil, false
		}
		values = append(values, item.Value)
	}
	return values, true
}

func (d *DocumentNode) KeyedRows() ([]string, map[string]*DocumentNode, bool) {
	if d == nil || d.Kind != nodekind.Sequence {
		return nil, nil, false
	}
	keys := make([]string, 0, len(d.Items))
	rows := map[string]*DocumentNode{}
	for _, item := range d.Items {
		key := item.Field(rowKeyField)
		if !item.IsMapping() || key == nil || key.Kind != nodekind.Scalar {
			return nil, nil, false
		}
		if _, duplicate := rows[key.Value]; duplicate {
			return nil, nil, false
		}
		keys = append(keys, key.Value)
		rows[key.Value] = item
	}
	return keys, rows, true
}

func (d *DocumentNode) sameFields(other *DocumentNode) bool {
	if len(d.Fields) != len(other.Fields) {
		return false
	}
	for key, value := range d.Fields {
		if !value.Equals(other.Fields[key]) {
			return false
		}
	}
	return true
}

func sameItems(items, others []*DocumentNode) bool {
	if len(items) != len(others) {
		return false
	}
	for index, item := range items {
		if !item.Equals(others[index]) {
			return false
		}
	}
	return true
}
