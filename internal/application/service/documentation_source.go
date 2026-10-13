package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
)

type documentationSource struct {
	findTreeItem     usecase.FindTreeItem
	findRequest      usecase.FindRequest
	writeApiDocument usecase.WriteApiDocument
	readApiDocument  usecase.ReadApiDocument
}

// collection gathers what the collection or folder documents; its ancestors above it only lend their auth.
func (d documentationSource) collection(ctx context.Context, root, id string) (*entity.DocumentedCollection, error) {
	node, err := d.findTreeItem.FindNode(ctx, root, id)
	if err != nil {
		return nil, err
	}
	if !node.IsContainer() {
		return nil, errs.BadRequestError("only collections and folders have documentation", "DOC-05400")
	}
	collection, err := d.findTreeItem.FindContainer(ctx, root, id)
	if err != nil {
		return nil, err
	}
	inherited, err := d.findTreeItem.FindAncestors(ctx, root, id)
	if err != nil {
		return nil, err
	}
	documented := entity.NewDocumentedCollection(*collection, inherited)
	return d.children(ctx, root, *node, []entity.Collection{*collection}, documented)
}

func (d documentationSource) children(ctx context.Context, root string, node entity.TreeNode, ancestors []entity.Collection, documented entity.DocumentedCollection) (*entity.DocumentedCollection, error) {
	for _, child := range entity.ExportableChildren(node) {
		if child.IsRequest() {
			request, err := d.findRequest.Find(ctx, root, child.ID)
			if err != nil {
				return nil, err
			}
			documented = documented.WithRequest(*request, ancestors)
			continue
		}
		folder, err := d.findTreeItem.FindContainer(ctx, root, child.ID)
		if err != nil {
			return nil, err
		}
		below := append(append([]entity.Collection{}, ancestors...), *folder)
		next, err := d.children(ctx, root, child, below, documented.WithFolder(below))
		if err != nil {
			return nil, err
		}
		documented = *next
	}
	return &documented, nil
}

// documentation writes the collection's or folder's OpenAPI text and reads it back, so the view shows exactly that text.
func (d documentationSource) documentation(ctx context.Context, documented entity.DocumentedCollection) (*entity.ApiDocumentation, error) {
	text, err := d.writeApiDocument.WriteDocument(ctx, documented.ApiDocument())
	if err != nil {
		return nil, err
	}
	return d.read(ctx, string(text))
}

func (d documentationSource) read(ctx context.Context, text string) (*entity.ApiDocumentation, error) {
	document, err := d.readApiDocument.ReadDocument(ctx, []byte(text))
	if err != nil {
		return nil, err
	}
	return &entity.ApiDocumentation{Text: text, Document: *document}, nil
}

// plan compares the edited text with the document generated from the collection or folder, both read the same way.
func (d documentationSource) plan(ctx context.Context, root, id, text string, newID func() string) (*entity.DocumentationPlan, error) {
	edited, err := d.read(ctx, text)
	if err != nil {
		return nil, err
	}
	documented, err := d.collection(ctx, root, id)
	if err != nil {
		return nil, err
	}
	baseline, err := d.documentation(ctx, *documented)
	if err != nil {
		return nil, err
	}
	plan := entity.NewDocumentationPlan(*documented, baseline.Document, edited.Document, newID)
	return &plan, nil
}
