package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/referencestatus"
)

const MaskedValue = "•••••"

type ResolvedReference struct {
	Value     string
	Status    enums.ReferenceStatus
	Scope     enums.VariableScope
	Reference Reference
	Secret    bool
}

func (r ResolvedReference) IsResolved() bool {
	return r.Status == referencestatus.Resolved || r.Status == referencestatus.Dynamic
}

func (r ResolvedReference) DisplayValue() string {
	switch {
	case r.Status == referencestatus.Dynamic:
		return ""
	case r.Secret:
		return MaskedValue
	default:
		return r.Value
	}
}
