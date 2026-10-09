package persistence

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence/model"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/fileutils"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/yamlnode"
)

type treePersistence struct {
	index adapter.WorkspaceIndexCache
}

func NewTreePersistence(index adapter.WorkspaceIndexCache) adapter.TreePersistence {
	return &treePersistence{index: index}
}

func (t *treePersistence) Load(ctx context.Context, root string) (*entity.Tree, error) {
	logger.Debug(ctx, "Started", map[string]any{"root": root})
	t.index.Lock(ctx)
	defer t.index.Unlock(ctx)

	index, err := t.index.Refresh(ctx, root)
	if err != nil {
		return nil, err
	}

	logger.Debug(ctx, "Finished", map[string]any{"root": root, "items": index.Size()})
	return &index.Tree, nil
}

func (t *treePersistence) FindContainer(ctx context.Context, root, id string) (*entity.Collection, error) {
	logger.Debug(ctx, "Started", map[string]any{"root": root, "id": id})
	t.index.Lock(ctx)
	defer t.index.Unlock(ctx)

	_, item, found, err := t.index.Resolve(ctx, root, id)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, errs.NotFoundError("the item does not exist", "TRE-30404")
	}
	if item.IsRequest() {
		return nil, errs.NotFoundError("requests have no collection or folder settings", "TRE-03404")
	}
	container, err := readContainerEntity(id, item)
	if err != nil {
		return nil, err
	}

	logger.Debug(ctx, "Finished", map[string]any{"root": root, "id": id})
	return container, nil
}

func (t *treePersistence) FindNode(ctx context.Context, root, id string) (*entity.TreeNode, error) {
	logger.Debug(ctx, "Started", map[string]any{"root": root, "id": id})
	t.index.Lock(ctx)
	defer t.index.Unlock(ctx)

	_, node, _, err := t.item(ctx, root, id)
	if err != nil {
		return nil, err
	}

	logger.Debug(ctx, "Finished", *node)
	return node, nil
}

func (t *treePersistence) FindAncestors(ctx context.Context, root, id string) ([]entity.Collection, error) {
	logger.Debug(ctx, "Started", map[string]any{"root": root, "id": id})
	t.index.Lock(ctx)
	defer t.index.Unlock(ctx)

	ancestors := make([]entity.Collection, 0)
	index, _, _, err := t.index.Resolve(ctx, root, id)
	if err != nil {
		return nil, err
	}
	for _, node := range index.Tree.Ancestors(id) {
		item, _ := index.Item(node.ID)
		container, err := readContainerEntity(node.ID, item)
		if err != nil {
			return nil, err
		}
		ancestors = append(ancestors, *container)
	}

	logger.Debug(ctx, "Finished", map[string]any{"root": root, "id": id, "ancestors": len(ancestors)})
	return ancestors, nil
}

func (t *treePersistence) CreateContainer(ctx context.Context, root, parentId, fileName string, container entity.Collection) (*entity.TreeNode, error) {
	logger.Debug(ctx, "Started", map[string]any{"root": root, "parentId": parentId, "fileName": fileName, "id": container.ID})
	t.index.Lock(ctx)
	defer t.index.Unlock(ctx)

	index, parent, parentDir, err := t.parent(ctx, root, parentId)
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(parentDir, fileName)
	if err := os.MkdirAll(parentDir, workspaceDirMode); err != nil {
		return nil, errs.FileError(err, "TRE-20500")
	}
	if err := os.Mkdir(dir, workspaceDirMode); err != nil {
		return nil, errs.FileError(err, "TRE-32500")
	}
	stored := model.NewCollection(container)
	if err := t.writeFields(ctx, filepath.Join(dir, markerFile(container.Kind)), stored.Fields()); err != nil {
		return nil, err
	}
	if err := t.writeOrder(ctx, root, index, parentId, parent.OrderAppending(fileName)); err != nil {
		return nil, err
	}
	node, err := t.reloaded(ctx, root, container.ID)
	if err != nil {
		return nil, err
	}

	logger.Debug(ctx, "Finished", node)
	return node, nil
}

func (t *treePersistence) CreateRequest(ctx context.Context, root, parentId, fileName string, request entity.Request) (*entity.Request, error) {
	logger.Debug(ctx, "Started", request, map[string]any{"root": root, "parentId": parentId, "fileName": fileName})
	t.index.Lock(ctx)
	defer t.index.Unlock(ctx)

	index, parent, parentDir, err := t.parent(ctx, root, parentId)
	if err != nil {
		return nil, err
	}
	path := filepath.Join(parentDir, fileName)
	if exists(path) {
		return nil, errs.ConflictError("a file with this name already exists", "TRE-26409")
	}
	stored := model.NewRequest(request)
	if err := t.writeFields(ctx, path, stored.Fields()); err != nil {
		return nil, err
	}
	if err := t.writeOrder(ctx, root, index, parentId, parent.OrderAppending(fileName)); err != nil {
		return nil, err
	}
	if _, err := t.index.Refresh(ctx, root); err != nil {
		return nil, err
	}
	saved, err := readRequestEntity(request.ID, path)
	if err != nil {
		return nil, err
	}

	logger.Debug(ctx, "Finished", saved)
	return saved, nil
}

func (t *treePersistence) Rename(ctx context.Context, root, id, name, fileName string) (*entity.TreeNode, error) {
	logger.Debug(ctx, "Started", map[string]any{"root": root, "id": id, "name": name, "fileName": fileName})
	t.index.Lock(ctx)
	defer t.index.Unlock(ctx)

	index, node, item, err := t.item(ctx, root, id)
	if err != nil {
		return nil, err
	}
	path := filepath.Join(filepath.Dir(item.Path), fileName)
	if err := moveItem(item.Path, path); err != nil {
		return nil, err
	}
	if err := t.rewriteName(ctx, id, item.MovedTo(path), name, node.ChildOrder()); err != nil {
		return nil, err
	}
	parent, _ := index.Tree.Parent(id)
	if err := t.writeOrder(ctx, root, index, parent.ID, parent.OrderRenaming(node.FileName, fileName)); err != nil {
		return nil, err
	}
	renamed, err := t.reloaded(ctx, root, id)
	if err != nil {
		return nil, err
	}

	logger.Debug(ctx, "Finished", renamed)
	return renamed, nil
}

func (t *treePersistence) Duplicate(ctx context.Context, root, id string, copied entity.ItemCopy) (*entity.TreeNode, error) {
	logger.Debug(ctx, "Started", map[string]any{"root": root, "id": id, "copyId": copied.ID, "name": copied.Name, "fileName": copied.FileName})
	t.index.Lock(ctx)
	defer t.index.Unlock(ctx)

	index, node, item, err := t.item(ctx, root, id)
	if err != nil {
		return nil, err
	}
	if err := t.copyItem(ctx, copied, item.Path, filepath.Join(filepath.Dir(item.Path), copied.FileName)); err != nil {
		return nil, err
	}
	parent, _ := index.Tree.Parent(id)
	if err := t.writeOrder(ctx, root, index, parent.ID, parent.OrderInsertingAfter(node.FileName, copied.FileName)); err != nil {
		return nil, err
	}
	duplicate, err := t.reloaded(ctx, root, copied.ID)
	if err != nil {
		return nil, err
	}

	logger.Debug(ctx, "Finished", duplicate)
	return duplicate, nil
}

func (t *treePersistence) SaveContainer(ctx context.Context, root string, container entity.Collection) (*entity.Collection, error) {
	logger.Debug(ctx, "Started", map[string]any{"root": root, "id": container.ID})
	t.index.Lock(ctx)
	defer t.index.Unlock(ctx)

	_, _, item, err := t.item(ctx, root, container.ID)
	if err != nil {
		return nil, err
	}
	stored := model.NewCollection(container)
	if err := t.writeFields(ctx, filepath.Join(item.Path, markerFile(item.Kind)), stored.Fields()); err != nil {
		return nil, err
	}
	saved, err := readContainerEntity(container.ID, item)
	if err != nil {
		return nil, err
	}

	logger.Debug(ctx, "Finished", map[string]any{"root": root, "id": saved.ID})
	return saved, nil
}

func (t *treePersistence) Delete(ctx context.Context, root, id string) error {
	logger.Debug(ctx, "Started", map[string]any{"root": root, "id": id})
	t.index.Lock(ctx)
	defer t.index.Unlock(ctx)

	index, node, item, err := t.item(ctx, root, id)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(item.Path); err != nil {
		return errs.FileError(err, "TRE-23500")
	}
	parent, _ := index.Tree.Parent(id)
	if err := t.writeOrder(ctx, root, index, parent.ID, parent.OrderRemoving(node.FileName)); err != nil {
		return err
	}
	if _, err := t.index.Refresh(ctx, root); err != nil {
		return err
	}

	logger.Debug(ctx, "Finished", map[string]any{"root": root, "id": id})
	return nil
}

func (t *treePersistence) parent(ctx context.Context, root, parentId string) (*entity.TreeIndex, *entity.TreeNode, string, error) {
	index, err := t.index.Current(ctx, root)
	if err != nil {
		return nil, nil, "", err
	}
	if parentId == "" {
		return index, &index.Tree.Root, filepath.Join(root, collectionsDir), nil
	}
	parent, found := index.Tree.Container(parentId)
	if !found {
		return nil, nil, "", errs.NotFoundError("the parent collection or folder does not exist", "TRE-28404")
	}
	item, _ := index.Item(parentId)
	return index, parent, item.Path, nil
}

func (t *treePersistence) item(ctx context.Context, root, id string) (*entity.TreeIndex, *entity.TreeNode, entity.IndexedItem, error) {
	index, item, found, err := t.index.Resolve(ctx, root, id)
	if err != nil {
		return nil, nil, entity.IndexedItem{}, err
	}
	node, inTree := index.Tree.Item(id)
	if !found || !inTree {
		return nil, nil, entity.IndexedItem{}, errs.NotFoundError("the item does not exist", "TRE-29404")
	}
	return index, node, item, nil
}

func (t *treePersistence) reloaded(ctx context.Context, root, id string) (*entity.TreeNode, error) {
	index, err := t.index.Refresh(ctx, root)
	if err != nil {
		return nil, err
	}
	node, found := index.Tree.Item(id)
	if !found {
		return nil, errs.NotFoundError("the item could not be read back", "TRE-31404")
	}
	return node, nil
}

func (t *treePersistence) writeOrder(ctx context.Context, root string, index *entity.TreeIndex, parentId string, order []string) error {
	if parentId == "" {
		_, stored, err := readWorkspace(root)
		if err != nil {
			return err
		}
		workspace := stored.ToEntity()
		workspace.Order = order
		updated := model.NewWorkspace(workspace)
		return t.writeFields(ctx, filepath.Join(root, workspaceFile), updated.Fields())
	}
	item, _ := index.Item(parentId)
	container, err := readContainerEntity(parentId, item)
	if err != nil {
		return err
	}
	container.Order = order
	stored := model.NewCollection(*container)
	return t.writeFields(ctx, filepath.Join(item.Path, markerFile(item.Kind)), stored.Fields())
}

func (t *treePersistence) rewriteName(ctx context.Context, id string, item entity.IndexedItem, name string, order []string) error {
	if item.IsRequest() {
		request, err := readRequestEntity(id, item.Path)
		if err != nil {
			return err
		}
		request.Name = name
		stored := model.NewRequest(*request)
		return t.writeFields(ctx, item.Path, stored.Fields())
	}
	container, err := readContainerEntity(id, item)
	if err != nil {
		return err
	}
	container.Name = name
	container.Order = order
	stored := model.NewCollection(*container)
	return t.writeFields(ctx, filepath.Join(item.Path, markerFile(item.Kind)), stored.Fields())
}

func (t *treePersistence) copyItem(ctx context.Context, copied entity.ItemCopy, source, target string) error {
	if copied.IsRequest() {
		return t.copyDocument(ctx, copied, source, target)
	}
	if !copied.IsContainer() {
		return t.copyRaw(ctx, source, target)
	}
	if err := os.Mkdir(target, workspaceDirMode); err != nil {
		return errs.FileError(err, "TRE-33500")
	}
	marker := markerFile(copied.Kind)
	if err := t.copyDocument(ctx, copied, filepath.Join(source, marker), filepath.Join(target, marker)); err != nil {
		return err
	}
	for _, child := range copied.Children {
		if err := t.copyItem(ctx, child, filepath.Join(source, child.FileName), filepath.Join(target, child.FileName)); err != nil {
			return err
		}
	}
	return nil
}

func (t *treePersistence) copyDocument(ctx context.Context, copied entity.ItemCopy, source, target string) error {
	data, err := os.ReadFile(source)
	if err != nil {
		return errs.FileError(err, "TRE-25500")
	}
	document, err := yamlnode.Parse(data)
	if err != nil {
		return t.copyRaw(ctx, source, target)
	}
	stored := model.NewItemCopy(copied)
	yamlnode.Patch(yamlnode.Root(document), stored.Fields())
	return t.writeDocument(ctx, target, document)
}

func (t *treePersistence) copyRaw(ctx context.Context, source, target string) error {
	info, err := os.Lstat(source)
	if err != nil || !info.Mode().IsRegular() {
		return nil
	}
	data, err := os.ReadFile(source)
	if err != nil {
		return errs.FileError(err, "TRE-36500")
	}
	if err := fileutils.WriteFileAtomic(target, data, workspaceFileMode); err != nil {
		return errs.FileError(err, "TRE-34500")
	}
	t.index.RecordWrite(ctx, target, data)
	return nil
}

func (t *treePersistence) writeFields(ctx context.Context, path string, fields []yamlnode.Field) error {
	document := yamlnode.NewDocument()
	if data, err := os.ReadFile(path); err == nil {
		if parsed, err := yamlnode.Parse(data); err == nil {
			document = parsed
		}
	}
	yamlnode.Patch(yamlnode.Root(document), fields)
	return t.writeDocument(ctx, path, document)
}

func (t *treePersistence) writeDocument(ctx context.Context, path string, document *yaml.Node) error {
	data, err := yamlnode.Encode(document)
	if err != nil {
		return errs.InternalServerError(err, "TRE-24500", "the file could not be encoded")
	}
	if err := fileutils.WriteFileAtomic(path, data, workspaceFileMode); err != nil {
		return errs.FileError(err, "TRE-21500")
	}
	t.index.RecordWrite(ctx, path, data)
	return nil
}

func moveItem(source, target string) error {
	if source == target {
		return nil
	}
	if strings.EqualFold(source, target) {
		temporary := fileutils.TempPath(target)
		if err := os.Rename(source, temporary); err != nil {
			return errs.FileError(err, "TRE-22500")
		}
		source = temporary
	} else if exists(target) {
		return errs.ConflictError("a file with this name already exists", "TRE-37409")
	}
	if err := os.Rename(source, target); err != nil {
		return errs.FileError(err, "TRE-35500")
	}
	return nil
}

func readContainerEntity(id string, item entity.IndexedItem) (*entity.Collection, error) {
	stored, reason := readCollection(filepath.Join(item.Path, markerFile(item.Kind)), item.Kind)
	if reason != "" {
		return nil, errs.UnprocessableEntityError(reason, "TRE-27422")
	}
	container := stored.ToContainerEntity(id, item.Kind)
	return &container, nil
}
