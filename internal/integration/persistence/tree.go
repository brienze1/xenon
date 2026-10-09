package persistence

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence/model"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/fileutils"
	"github.com/brienze1/xenon/pkg/idutils"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/yamlnode"
)

type treePersistence struct {
	index adapter.WorkspaceIndex
}

func NewTreePersistence(index adapter.WorkspaceIndex) adapter.TreePersistence {
	return &treePersistence{index: index}
}

func (t *treePersistence) Load(ctx context.Context, root string) (*entity.Tree, error) {
	logger.Debug(ctx, "Started", map[string]any{"root": root})
	t.index.Lock()
	defer t.index.Unlock()

	index, err := t.index.Refresh(root)
	if err != nil {
		return nil, err
	}

	logger.Debug(ctx, "Finished", map[string]any{"root": root, "items": len(index.Items)})
	return &index.Tree, nil
}

func (t *treePersistence) FindContainer(ctx context.Context, root, id string) (*entity.Collection, error) {
	logger.Debug(ctx, "Started", map[string]any{"root": root, "id": id})
	t.index.Lock()
	defer t.index.Unlock()

	_, item, found, err := t.index.Resolve(root, id)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, errs.NotFoundError("the item does not exist", "TRE-30404")
	}
	if item.Kind == itemkind.Request {
		return nil, errs.NotFoundError("requests have no collection or folder settings", "TRE-03404")
	}
	container, err := readContainerEntity(id, item)
	if err != nil {
		return nil, err
	}

	logger.Debug(ctx, "Finished", map[string]any{"root": root, "id": id})
	return container, nil
}

func (t *treePersistence) FindAncestors(ctx context.Context, root, id string) ([]entity.Collection, error) {
	logger.Debug(ctx, "Started", map[string]any{"root": root, "id": id})
	t.index.Lock()
	defer t.index.Unlock()

	ancestors := make([]entity.Collection, 0)
	index, _, _, err := t.index.Resolve(root, id)
	if err != nil {
		return nil, err
	}
	for _, node := range index.Tree.Ancestors(id) {
		container, err := readContainerEntity(node.ID, index.Items[node.ID])
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
	t.index.Lock()
	defer t.index.Unlock()

	index, parent, parentDir, err := t.parent(root, parentId)
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
	if err := t.writeFields(filepath.Join(dir, markerFile(container.Kind)), stored.Fields()); err != nil {
		return nil, err
	}
	if err := t.writeOrder(root, index, parentId, parent.OrderAppending(fileName)); err != nil {
		return nil, err
	}
	node, err := t.reloaded(root, container.ID)
	if err != nil {
		return nil, err
	}

	logger.Debug(ctx, "Finished", node)
	return node, nil
}

func (t *treePersistence) CreateRequest(ctx context.Context, root, parentId, fileName string, request entity.Request) (*entity.Request, error) {
	logger.Debug(ctx, "Started", request, map[string]any{"root": root, "parentId": parentId, "fileName": fileName})
	t.index.Lock()
	defer t.index.Unlock()

	index, parent, parentDir, err := t.parent(root, parentId)
	if err != nil {
		return nil, err
	}
	path := filepath.Join(parentDir, fileName)
	if exists(path) {
		return nil, errs.ConflictError("a file with this name already exists", "TRE-26409")
	}
	stored := model.NewRequest(request)
	if err := t.writeFields(path, stored.Fields()); err != nil {
		return nil, err
	}
	if err := t.writeOrder(root, index, parentId, parent.OrderAppending(fileName)); err != nil {
		return nil, err
	}
	if _, err := t.index.Refresh(root); err != nil {
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
	t.index.Lock()
	defer t.index.Unlock()

	index, node, item, err := t.item(root, id)
	if err != nil {
		return nil, err
	}
	path := filepath.Join(filepath.Dir(item.Path), fileName)
	if err := moveItem(item.Path, path); err != nil {
		return nil, err
	}
	if err := t.rewriteName(id, model.IndexedItem{Path: path, Kind: item.Kind}, name, node.ChildOrder()); err != nil {
		return nil, err
	}
	parent, _ := index.Tree.Parent(id)
	if err := t.writeOrder(root, index, parent.ID, parent.OrderRenaming(node.FileName, fileName)); err != nil {
		return nil, err
	}
	renamed, err := t.reloaded(root, id)
	if err != nil {
		return nil, err
	}

	logger.Debug(ctx, "Finished", renamed)
	return renamed, nil
}

func (t *treePersistence) Duplicate(ctx context.Context, root, id, name, fileName string) (*entity.TreeNode, error) {
	logger.Debug(ctx, "Started", map[string]any{"root": root, "id": id, "name": name, "fileName": fileName})
	t.index.Lock()
	defer t.index.Unlock()

	index, node, item, err := t.item(root, id)
	if err != nil {
		return nil, err
	}
	copyId, err := t.copyNode(*node, item.Path, filepath.Join(filepath.Dir(item.Path), fileName), name)
	if err != nil {
		return nil, err
	}
	parent, _ := index.Tree.Parent(id)
	if err := t.writeOrder(root, index, parent.ID, parent.OrderInsertingAfter(node.FileName, fileName)); err != nil {
		return nil, err
	}
	duplicate, err := t.reloaded(root, copyId)
	if err != nil {
		return nil, err
	}

	logger.Debug(ctx, "Finished", duplicate)
	return duplicate, nil
}

func (t *treePersistence) SaveContainer(ctx context.Context, root string, container entity.Collection) (*entity.Collection, error) {
	logger.Debug(ctx, "Started", map[string]any{"root": root, "id": container.ID})
	t.index.Lock()
	defer t.index.Unlock()

	_, node, item, err := t.item(root, container.ID)
	if err != nil {
		return nil, err
	}
	container.Order = node.ChildOrder()
	stored := model.NewCollection(container)
	if err := t.writeFields(filepath.Join(item.Path, markerFile(item.Kind)), stored.Fields()); err != nil {
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
	t.index.Lock()
	defer t.index.Unlock()

	index, node, item, err := t.item(root, id)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(item.Path); err != nil {
		return errs.FileError(err, "TRE-23500")
	}
	parent, _ := index.Tree.Parent(id)
	if err := t.writeOrder(root, index, parent.ID, parent.OrderRemoving(node.FileName)); err != nil {
		return err
	}
	if _, err := t.index.Refresh(root); err != nil {
		return err
	}

	logger.Debug(ctx, "Finished", map[string]any{"root": root, "id": id})
	return nil
}

func (t *treePersistence) parent(root, parentId string) (*model.TreeIndex, *entity.TreeNode, string, error) {
	index, err := t.index.Current(root)
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
	return index, parent, index.Items[parentId].Path, nil
}

func (t *treePersistence) item(root, id string) (*model.TreeIndex, *entity.TreeNode, model.IndexedItem, error) {
	index, item, found, err := t.index.Resolve(root, id)
	if err != nil {
		return nil, nil, model.IndexedItem{}, err
	}
	node, inTree := index.Tree.Item(id)
	if !found || !inTree {
		return nil, nil, model.IndexedItem{}, errs.NotFoundError("the item does not exist", "TRE-29404")
	}
	return index, node, item, nil
}

func (t *treePersistence) reloaded(root, id string) (*entity.TreeNode, error) {
	index, err := t.index.Refresh(root)
	if err != nil {
		return nil, err
	}
	node, found := index.Tree.Item(id)
	if !found {
		return nil, errs.NotFoundError("the item could not be read back", "TRE-31404")
	}
	return node, nil
}

func (t *treePersistence) writeOrder(root string, index *model.TreeIndex, parentId string, order []string) error {
	if parentId == "" {
		_, stored, err := readWorkspace(root)
		if err != nil {
			return err
		}
		workspace := stored.ToEntity()
		workspace.Order = order
		updated := model.NewWorkspace(workspace)
		return t.writeFields(filepath.Join(root, workspaceFile), updated.Fields())
	}
	item := index.Items[parentId]
	container, err := readContainerEntity(parentId, item)
	if err != nil {
		return err
	}
	container.Order = order
	stored := model.NewCollection(*container)
	return t.writeFields(filepath.Join(item.Path, markerFile(item.Kind)), stored.Fields())
}

func (t *treePersistence) rewriteName(id string, item model.IndexedItem, name string, order []string) error {
	if item.Kind == itemkind.Request {
		request, err := readRequestEntity(id, item.Path)
		if err != nil {
			return err
		}
		request.Name = name
		stored := model.NewRequest(*request)
		return t.writeFields(item.Path, stored.Fields())
	}
	container, err := readContainerEntity(id, item)
	if err != nil {
		return err
	}
	container.Name = name
	container.Order = order
	stored := model.NewCollection(*container)
	return t.writeFields(filepath.Join(item.Path, markerFile(item.Kind)), stored.Fields())
}

func (t *treePersistence) copyNode(node entity.TreeNode, source, target, name string) (string, error) {
	id := idutils.NewUlid()
	if node.IsRequest() {
		return id, t.copyDocument(source, target, id, name)
	}
	if !node.IsContainer() {
		return id, t.copyRaw(source, target)
	}
	if err := os.Mkdir(target, workspaceDirMode); err != nil {
		return "", errs.FileError(err, "TRE-33500")
	}
	marker := markerFile(node.Kind)
	if err := t.copyDocument(filepath.Join(source, marker), filepath.Join(target, marker), id, name); err != nil {
		return "", err
	}
	for _, child := range node.Children {
		if child.Ignored {
			continue
		}
		if _, err := t.copyNode(child, filepath.Join(source, child.FileName), filepath.Join(target, child.FileName), ""); err != nil {
			return "", err
		}
	}
	return id, nil
}

func (t *treePersistence) copyDocument(source, target, id, name string) error {
	data, err := os.ReadFile(source)
	if err != nil {
		return errs.FileError(err, "TRE-25500")
	}
	document, err := yamlnode.Parse(data)
	if err != nil {
		return t.copyRaw(source, target)
	}
	fields := []yamlnode.Field{yamlnode.String("id", id)}
	if name != "" {
		fields = append(fields, yamlnode.String("name", name))
	}
	yamlnode.Patch(yamlnode.Root(document), fields)
	return t.writeDocument(target, document)
}

func (t *treePersistence) copyRaw(source, target string) error {
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
	t.index.RecordWrite(target, data)
	return nil
}

func (t *treePersistence) writeFields(path string, fields []yamlnode.Field) error {
	document := yamlnode.NewDocument()
	if data, err := os.ReadFile(path); err == nil {
		if parsed, err := yamlnode.Parse(data); err == nil {
			document = parsed
		}
	}
	yamlnode.Patch(yamlnode.Root(document), fields)
	return t.writeDocument(path, document)
}

func (t *treePersistence) writeDocument(path string, document *yaml.Node) error {
	data, err := yamlnode.Encode(document)
	if err != nil {
		return errs.InternalServerError(err, "TRE-24500", "the file could not be encoded")
	}
	if err := fileutils.WriteFileAtomic(path, data, workspaceFileMode); err != nil {
		return errs.FileError(err, "TRE-21500")
	}
	t.index.RecordWrite(path, data)
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

func readContainerEntity(id string, item model.IndexedItem) (*entity.Collection, error) {
	stored, reason := readCollection(filepath.Join(item.Path, markerFile(item.Kind)), item.Kind)
	if reason != "" {
		return nil, errs.UnprocessableEntityError(reason, "TRE-27422")
	}
	container := stored.ToContainerEntity(id, item.Kind)
	return &container, nil
}
