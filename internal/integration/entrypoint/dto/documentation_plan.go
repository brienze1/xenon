package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type DocumentationPlan struct {
	Name              string              `json:"name"`
	Updated           []DocumentationItem `json:"updated"`
	Created           []DocumentationItem `json:"created"`
	Deleted           []DocumentationItem `json:"deleted"`
	CollectionChanged bool                `json:"collection_changed"`
}

func NewDocumentationPlan(plan entity.DocumentationPlan) DocumentationPlan {
	documented := DocumentationPlan{
		Name:              plan.Collection.Name,
		CollectionChanged: plan.CollectionChanged,
		Updated:           make([]DocumentationItem, 0, len(plan.Updated)),
		Created:           make([]DocumentationItem, 0, len(plan.Created)),
		Deleted:           make([]DocumentationItem, 0, len(plan.Deleted)),
	}
	for _, request := range plan.Updated {
		documented.Updated = append(documented.Updated, NewDocumentationItem(request))
	}
	for _, created := range plan.Created {
		documented.Created = append(documented.Created, NewCreatedDocumentationItem(created))
	}
	for _, request := range plan.Deleted {
		documented.Deleted = append(documented.Deleted, NewDocumentationItem(request))
	}
	return documented
}

func (d DocumentationPlan) LogValue() any {
	return map[string]any{"collection_changed": d.CollectionChanged, "updated": len(d.Updated), "created": len(d.Created), "deleted": len(d.Deleted)}
}
