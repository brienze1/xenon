package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/clashresolution"
)

type ImportPlan struct {
	SourceID    string
	Passphrase  string
	ApplyToAll  enums.ClashResolution
	Items       []ImportPlanItem
	SkipSecrets bool
}

func (i ImportPlan) NeedsPassphrase(source ImportSource) bool {
	return source.HasEncryptedSecrets() && !i.SkipSecrets
}

func (i ImportPlan) Included() []ImportPlanItem {
	included := make([]ImportPlanItem, 0, len(i.Items))
	for _, item := range i.Items {
		if item.Include {
			included = append(included, item)
		}
	}
	return included
}

func (i ImportPlan) ResolutionFor(item ImportPlanItem) enums.ClashResolution {
	if item.Resolution != clashresolution.Unresolved {
		return item.Resolution
	}
	return i.ApplyToAll
}

func (i ImportPlan) Resolves(item ImportPlanItem) bool {
	return i.ResolutionFor(item) != clashresolution.Unresolved
}

func (i ImportPlan) LogValue() any {
	return map[string]any{"sourceId": i.SourceID, "items": len(i.Items), "applyToAll": i.ApplyToAll, "skipSecrets": i.SkipSecrets}
}
