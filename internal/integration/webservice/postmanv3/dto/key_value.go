package dto

type KeyValue struct {
	Extra       map[string]any `yaml:",inline"`
	Value       any            `yaml:"value"`
	Key         string         `yaml:"key"`
	Description string         `yaml:"description"`
	Disabled    bool           `yaml:"disabled"`
}

func NewKeyValue(key string, value any) KeyValue {
	return KeyValue{Key: key, Value: value}
}

func (k *KeyValue) toPostmanV21() map[string]any {
	row := copied(k.Extra)
	row["key"] = k.Key
	if _, hasSrc := k.Extra["src"]; k.Value != nil || !hasSrc {
		row["value"] = text(k.Value)
	}
	if k.Disabled {
		row["disabled"] = true
	}
	if k.Description != "" {
		row["description"] = k.Description
	}
	return row
}

func (k *KeyValue) toCredential() map[string]any {
	switch k.Value.(type) {
	case bool:
		return map[string]any{"key": k.Key, "value": k.Value, "type": "boolean"}
	case int, int64, float64:
		return map[string]any{"key": k.Key, "value": k.Value, "type": "number"}
	default:
		return map[string]any{"key": k.Key, "value": text(k.Value), "type": "string"}
	}
}
