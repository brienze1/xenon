package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type Pair struct {
	Key     string
	Value   string
	Line    int
	Enabled bool
}

func NewPair(key, value string, enabled bool) Pair {
	return Pair{Key: key, Value: value, Enabled: enabled}
}

func (p Pair) ToKeyValue() entity.KeyValue {
	row := entity.NewKeyValue(p.Key, p.Value)
	row.Enabled = p.Enabled
	return row
}

func (p Pair) ToVariable() entity.BrunoVariable {
	return entity.BrunoVariable{Key: p.Key, Value: p.Value, Line: p.Line, Enabled: p.Enabled}
}

func keyValues(pairs []Pair) []entity.KeyValue {
	rows := make([]entity.KeyValue, 0, len(pairs))
	for _, pair := range pairs {
		rows = append(rows, pair.ToKeyValue())
	}
	return rows
}

func variables(pairs []Pair) []entity.BrunoVariable {
	rows := make([]entity.BrunoVariable, 0, len(pairs))
	for _, pair := range pairs {
		rows = append(rows, pair.ToVariable())
	}
	return rows
}
