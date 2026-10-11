package dto

import (
	"bytes"
	"encoding/json"
)

type object struct {
	values map[string]any
	keys   []string
}

func newObject() *object {
	return &object{values: map[string]any{}}
}

func (o *object) set(key string, value any) *object {
	if _, found := o.values[key]; !found {
		o.keys = append(o.keys, key)
	}
	o.values[key] = value
	return o
}

func (o *object) setIf(condition bool, key string, value any) *object {
	if condition {
		o.set(key, value)
	}
	return o
}

func (o *object) setPresent(key string, value any) *object {
	return o.setIf(value != nil, key, value)
}

func (o *object) merged(passthrough map[string]any) *object {
	for _, key := range sortedKeys(passthrough) {
		value := passthrough[key]
		existing, found := o.values[key]
		if !found {
			o.set(key, value)
			continue
		}
		if nested, isObject := existing.(*object); isObject {
			nested.merged(asMap(value))
		}
	}
	return o
}

func (o *object) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteByte('{')
	for index, key := range o.keys {
		if index > 0 {
			buffer.WriteByte(',')
		}
		name, err := encodeJson(key)
		if err != nil {
			return nil, err
		}
		value, err := encodeJson(o.values[key])
		if err != nil {
			return nil, err
		}
		buffer.Write(name)
		buffer.WriteByte(':')
		buffer.Write(value)
	}
	buffer.WriteByte('}')
	return buffer.Bytes(), nil
}

func encodeJson(value any) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return bytes.TrimRight(buffer.Bytes(), "\n"), nil
}

// EncodeDocument writes a Postman document tab-indented like Postman's own exports, without HTML escaping.
func EncodeDocument(document any) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "\t")
	if err := encoder.Encode(document); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
