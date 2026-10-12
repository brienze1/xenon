package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/idutils"
)

type documentationWriter struct {
	findTreeItem        usecase.FindTreeItem
	loadTree            usecase.LoadTree
	saveTreeItem        usecase.SaveTreeItem
	saveRequestExamples usecase.SaveRequestExamples
	createTreeItem      usecase.CreateTreeItem
	deleteTreeItem      usecase.DeleteTreeItem
	renameTreeItem      usecase.RenameTreeItem
}

// write saves every change of the plan, one atomic write per file: the collection, then the requests.
func (d documentationWriter) write(ctx context.Context, root string, plan entity.DocumentationPlan) (*entity.DocumentationPlan, error) {
	written := plan
	if err := d.describeCollection(ctx, root, plan); err != nil {
		return nil, err
	}
	written.Updated = make([]entity.Request, 0, len(plan.Updated))
	for _, request := range plan.Updated {
		saved, err := d.saveRequestExamples.SaveExamples(ctx, root, request)
		if err != nil {
			return nil, err
		}
		written.Updated = append(written.Updated, *saved)
	}
	written.Created = make([]entity.CreatedRequest, 0, len(plan.Created))
	for _, created := range plan.Created {
		saved, err := d.create(ctx, root, plan.Collection.ID, created)
		if err != nil {
			return nil, err
		}
		written.Created = append(written.Created, *saved)
	}
	for _, request := range plan.Deleted {
		if err := d.deleteTreeItem.Delete(ctx, root, request.ID); err != nil {
			return nil, err
		}
	}
	return &written, d.renameCollection(ctx, root, plan)
}

func (d documentationWriter) describeCollection(ctx context.Context, root string, plan entity.DocumentationPlan) error {
	if !plan.CollectionChanged {
		return nil
	}
	container, err := d.findTreeItem.FindContainer(ctx, root, plan.Collection.ID)
	if err != nil {
		return err
	}
	node, err := d.findTreeItem.FindNode(ctx, root, plan.Collection.ID)
	if err != nil {
		return err
	}
	_, err = d.saveTreeItem.SaveContainer(ctx, root, container.Described(plan.Collection.Description).WithOrder(node.ChildOrder()))
	return err
}

func (d documentationWriter) renameCollection(ctx context.Context, root string, plan entity.DocumentationPlan) error {
	tree, err := d.loadTree.Load(ctx, root)
	if err != nil {
		return err
	}
	node, found := tree.Item(plan.Collection.ID)
	if !found {
		return errs.NotFoundError("the collection does not exist", "DOC-06404")
	}
	if node.Name == plan.Collection.Name {
		return nil
	}
	fileName := usecase.DeriveFileName(plan.Collection.Name, node.Kind, tree.SiblingEntries(node.ID))
	_, err = d.renameTreeItem.Rename(ctx, root, node.ID, plan.Collection.Name, fileName)
	return err
}

func (d documentationWriter) create(ctx context.Context, root, collectionID string, created entity.CreatedRequest) (*entity.CreatedRequest, error) {
	tree, err := d.loadTree.Load(ctx, root)
	if err != nil {
		return nil, err
	}
	parent, found := tree.Item(collectionID)
	if !found {
		return nil, errs.NotFoundError("the collection does not exist", "DOC-07404")
	}
	for _, name := range created.Folders {
		if parent, err = d.folder(ctx, root, *parent, name); err != nil {
			return nil, err
		}
	}
	fileName := usecase.DeriveFileName(created.Request.Name, itemkind.Request, parent.Entries)
	saved, err := d.createTreeItem.CreateRequest(ctx, root, parent.ID, fileName, created.Request)
	if err != nil {
		return nil, err
	}
	created.Request = *saved
	return &created, nil
}

func (d documentationWriter) folder(ctx context.Context, root string, parent entity.TreeNode, name string) (*entity.TreeNode, error) {
	if existing, found := parent.ChildFolder(name); found {
		return &existing, nil
	}
	folder := entity.NewFolder(idutils.NewUlid(), name)
	return d.createTreeItem.CreateContainer(ctx, root, parent.ID, usecase.DeriveFileName(folder.Name, folder.Kind, parent.Entries), folder)
}
