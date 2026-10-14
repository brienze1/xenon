package entity

import (
	"slices"

	"github.com/brienze1/xenon/pkg/errs"
)

// DocumentationPlan is what saving an edited OpenAPI document changes in its collection or folder.
type DocumentationPlan struct {
	Updated           []Request
	Created           []CreatedRequest
	Deleted           []Request
	Collection        Collection
	CollectionChanged bool
}

// NewDocumentationPlan compares the edited document with the one generated from the collection (the baseline)
// and maps every difference back onto the requests, matched by x-xenon-id.
func NewDocumentationPlan(documented DocumentedCollection, baseline, edited ApiDocument, newID func() string) DocumentationPlan {
	plan := DocumentationPlan{Collection: documented.Collection, Updated: []Request{}, Created: []CreatedRequest{}, Deleted: []Request{}}
	plan = plan.withInfo(baseline, edited)
	kept := make([]string, 0, len(edited.Operations))
	for _, operation := range edited.Operations {
		current, exists := documented.Request(operation.RequestID)
		if !exists || slices.Contains(kept, operation.RequestID) {
			plan.Created = append(plan.Created, newCreatedRequest(operation, edited.ServerOf(operation), newID))
			continue
		}
		kept = append(kept, operation.RequestID)
		previous, _ := baseline.Operation(operation.RequestID)
		if updated, changed := operation.ApplyTo(current.Request, previous, edited.ServerOf(operation), baseline.ServerOf(previous), newID); changed {
			plan.Updated = append(plan.Updated, updated)
		}
	}
	for _, current := range documented.Requests {
		if !slices.Contains(kept, current.Request.ID) {
			plan.Deleted = append(plan.Deleted, current.Request)
		}
	}
	return plan
}

func (d DocumentationPlan) HasDeletions() bool {
	return len(d.Deleted) > 0
}

func (d DocumentationPlan) withInfo(baseline, edited ApiDocument) DocumentationPlan {
	if edited.Title != baseline.Title && NormalizeItemName(edited.Title) != "" {
		d.Collection.Name, d.CollectionChanged = NormalizeItemName(edited.Title), true
	}
	if edited.Description != baseline.Description {
		d.Collection.Description, d.CollectionChanged = edited.Description, true
	}
	return d
}

func newCreatedRequest(operation ApiOperation, server string, newID func() string) CreatedRequest {
	request, _ := operation.ApplyTo(NewRequest(newID(), operation.Name()), ApiOperation{}, server, "", newID)
	return CreatedRequest{Request: request, Path: operation.Path, Folders: FolderNames(operation.Tag())}
}

// KeepingDeleted leaves in place the requests the edited document no longer lists.
func (d DocumentationPlan) KeepingDeleted() DocumentationPlan {
	d.Deleted = []Request{}
	return d
}

func (d DocumentationPlan) Applying(keepDeleted bool) DocumentationPlan {
	if keepDeleted {
		return d.KeepingDeleted()
	}
	return d
}

// NeedsConfirmation is true while the plan deletes requests the user has not agreed to delete.
func (d DocumentationPlan) NeedsConfirmation(deletionsConfirmed bool) bool {
	return d.HasDeletions() && !deletionsConfirmed
}

// ConfirmedBy refuses the plan while it deletes requests the user has not agreed to delete.
func (d DocumentationPlan) ConfirmedBy(deletionsConfirmed bool) error {
	if d.NeedsConfirmation(deletionsConfirmed) {
		return errs.BadRequestError("saving deletes requests; confirm the deletions first", "DOC-04400")
	}
	return nil
}

func (d DocumentationPlan) LogValue() any {
	return map[string]any{"collectionChanged": d.CollectionChanged, "updated": len(d.Updated), "created": len(d.Created), "deleted": len(d.Deleted)}
}
