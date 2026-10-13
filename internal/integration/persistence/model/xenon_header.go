package model

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
)

type XenonHeader struct {
	Kind    string `yaml:"kind"`
	Secrets string `yaml:"secrets"`
	Id      string `yaml:"id"`
	Root    string `yaml:"root"`
	Xenon   int    `yaml:"xenon"`
}

func (x *XenonHeader) VariablesScope() (entity.ScopeRef, bool) {
	switch {
	case itemkind.Environment.Equals(x.Kind) && x.Id != "":
		return entity.NewEnvironmentScopeRef(x.Id), true
	case itemkind.Collection.Equals(x.Kind) && x.Id != "":
		return entity.NewCollectionScopeRef(x.Id), true
	case itemkind.Workspace.Equals(x.Kind):
		return entity.NewGlobalsScopeRef(), true
	default:
		return entity.ScopeRef{}, false
	}
}

func (x *XenonHeader) Probed(probe entity.ImportProbe) entity.ImportProbe {
	probe.XenonFormat = x.Xenon
	probe.XenonKind = x.Kind
	probe.XenonRoot = x.Root
	return probe
}
