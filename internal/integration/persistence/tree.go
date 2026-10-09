package persistence

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence/model"
	"github.com/brienze1/xenon/internal/integration/persistence/yamlnode"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/fileutils"
	"github.com/brienze1/xenon/pkg/idutils"
	"github.com/brienze1/xenon/pkg/logger"
)

type treePersistence struct {
	index *WorkspaceIndex
}

func NewTreePersistence(index *WorkspaceIndex) adapter.TreePersistence {
	return &treePersistence{index: index}
}

func (t *treePersistence) Load(ctx context.Context, root string) (*entity.Tree, error) {
	logger.Debug(ctx, "Started", map[string]any{"root": root})
	t.index.mu.Lock()
	defer t.index.mu.Unlock()

	index, err := t.index.refresh(root)
	if err != nil {
		return nil, err
	}

	logger.Debug(ctx, "Finished", map[string]any{"root": root, "items": len(index.items)})
	return &index.tree, nil
}

func (t *treePersistence) FindContainer(ctx context.Context, root, id string) (*entity.Collection, error) {
	logger.Debug(ctx, "Started", map[string]any{"root": root, "id": id})
	t.index.mu.Lock()
	defer t.index.mu.Unlock()

	_, item, found, err := t.index.resolve(root, id)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, errs.NotFoundError("the item does not exist", "TRE-02404")
	}
	if item.kind == itemkind.Request {
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
	t.index.mu.Lock()
	defer t.index.mu.Unlock()

	ancestors := make([]entity.Collection, 0)
	if strings.TrimSpace(id) == "" {
		logger.Debug(ctx, "Finished", map[string]any{"root": root, "id": id, "ancestors": 0})
		return ancestors, nil
	}
	index, _, _, err := t.index.resolve(root, id)
	if err != nil {
		return nil, err
	}
	for _, node := range index.tree.Ancestors(id) {
		container, err := readContainerEntity(node.ID, index.items[node.ID])
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
	t.index.mu.Lock()
	defer t.index.mu.Unlock()

	index, parent, parentDir, err := t.parent(root, parentId)
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(parentDir, fileName)
	if err := os.MkdirAll(parentDir, workspaceDirMode); err != nil {
		return nil, errs.FileError(err, "TRE-20500")
	}
	if err := os.Mkdir(dir, workspaceDirMode); err != nil {
		return nil, errs.FileError(err, "TRE-20500")
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
	t.index.mu.Lock()
	defer t.index.mu.Unlock()

	index, parent, parentDir, err := t.parent(root, parentId)
	if err != nil {
		return nil, err
	}
	path := filepath.Join(parentDir, fileName)
	if exists(path) {
		return nil, errs.ClientError("a file with this name already exists", "TRE-26409", http.StatusConflict)
	}
	stored := model.NewRequest(request)
	if err := t.writeFields(path, stored.Fields()); err != nil {
		return nil, err
	}
	if err := t.writeOrder(root, index, parentId, parent.OrderAppending(fileName)); err != nil {
		return nil, err
	}
	if _, err := t.index.refresh(root); err != nil {
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
	t.index.mu.Lock()
	defer t.index.mu.Unlock()

	index, node, item, err := t.item(root, id)
	if err != nil {
		return nil, err
	}
	path := filepath.Join(filepath.Dir(item.path), fileName)
	if err := moveItem(item.path, path); err != nil {
		return nil, err
	}
	if err := t.rewriteName(id, indexedItem{path: path, kind: item.kind}, strings.TrimSpace(name), node.ChildOrder()); err != nil {
		return nil, err
	}
	parent, _ := index.tree.Parent(id)
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
	t.index.mu.Lock()
	defer t.index.mu.Unlock()

	index, node, item, err := t.item(root, id)
	if err != nil {
		return nil, err
	}
	copyId, err := t.copyNode(*node, item.path, filepath.Join(filepath.Dir(item.path), fileName), strings.TrimSpace(name))
	if err != nil {
		return nil, err
	}
	parent, _ := index.tree.Parent(id)
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
	t.index.mu.Lock()
	defer t.index.mu.Unlock()

	_, node, item, err := t.item(root, container.ID)
	if err != nil {
		return nil, err
	}
	container.Order = node.ChildOrder()
	stored := model.NewCollection(container)
	if err := t.writeFields(filepath.Join(item.path, markerFile(item.kind)), stored.Fields()); err != nil {
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
	t.index.mu.Lock()
	defer t.index.mu.Unlock()

	index, node, item, err := t.item(root, id)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(item.path); err != nil {
		return errs.FileError(err, "TRE-23500")
	}
	parent, _ := index.tree.Parent(id)
	if err := t.writeOrder(root, index, parent.ID, parent.OrderRemoving(node.FileName)); err != nil {
		return err
	}
	if _, err := t.index.refresh(root); err != nil {
		return err
	}

	logger.Debug(ctx, "Finished", map[string]any{"root": root, "id": id})
	return nil
}

func (t *treePersistence) parent(root, parentId string) (*rootIndex, *entity.TreeNode, string, error) {
	index, err := t.index.current(root)
	if err != nil {
		return nil, nil, "", err
	}
	if parentId == "" {
		return index, &index.tree.Root, filepath.Join(root, collectionsDir), nil
	}
	parent, found := index.tree.Container(parentId)
	if !found {
		return nil, nil, "", errs.NotFoundError("the parent collection or folder does not exist", "TRE-01404")
	}
	return index, parent, index.items[parentId].path, nil
}

func (t *treePersistence) item(root, id string) (*rootIndex, *entity.TreeNode, indexedItem, error) {
	index, item, found, err := t.index.resolve(root, id)
	if err != nil {
		return nil, nil, indexedItem{}, err
	}
	node, inTree := index.tree.Item(id)
	if !found || !inTree {
		return nil, nil, indexedItem{}, errs.NotFoundError("the item does not exist", "TRE-02404")
	}
	return index, node, item, nil
}

func (t *treePersistence) reloaded(root, id string) (*entity.TreeNode, error) {
	index, err := t.index.refresh(root)
	if err != nil {
		return nil, err
	}
	node, found := index.tree.Item(id)
	if !found {
		return nil, errs.NotFoundError("the item could not be read back", "TRE-02404")
	}
	return node, nil
}

func (t *treePersistence) writeOrder(root string, index *rootIndex, parentId string, order []string) error {
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
	item := index.items[parentId]
	container, err := readContainerEntity(parentId, item)
	if err != nil {
		return err
	}
	container.Order = order
	stored := model.NewCollection(*container)
	return t.writeFields(filepath.Join(item.path, markerFile(item.kind)), stored.Fields())
}

func (t *treePersistence) rewriteName(id string, item indexedItem, name string, order []string) error {
	if item.kind == itemkind.Request {
		request, err := readRequestEntity(id, item.path)
		if err != nil {
			return err
		}
		request.Name = name
		stored := model.NewRequest(*request)
		return t.writeFields(item.path, stored.Fields())
	}
	container, err := readContainerEntity(id, item)
	if err != nil {
		return err
	}
	container.Name = name
	container.Order = order
	stored := model.NewCollection(*container)
	return t.writeFields(filepath.Join(item.path, markerFile(item.kind)), stored.Fields())
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
		return "", errs.FileError(err, "TRE-20500")
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
		return errs.FileError(err, "TRE-25500")
	}
	if err := fileutils.WriteFileAtomic(target, data, workspaceFileMode); err != nil {
		return errs.FileError(err, "TRE-21500")
	}
	t.index.recordWrite(target, data)
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
	t.index.recordWrite(path, data)
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
		return errs.ClientError("a file with this name already exists", "TRE-26409", http.StatusConflict)
	}
	if err := os.Rename(source, target); err != nil {
		return errs.FileError(err, "TRE-22500")
	}
	return nil
}

func readContainerEntity(id string, item indexedItem) (*entity.Collection, error) {
	stored, reason := readCollection(filepath.Join(item.path, markerFile(item.kind)), item.kind)
	if reason != "" {
		return nil, errs.ClientError(reason, "TRE-27422", http.StatusUnprocessableEntity)
	}
	container := stored.ToEntity()
	container.ID = id
	container.Kind = item.kind
	return &container, nil
}

func readRequestEntity(id, path string) (*entity.Request, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errs.FileError(err, "REQ-12500")
	}
	stored, reason := readRequest(path)
	if reason != "" {
		return nil, errs.ClientError(reason, "REQ-15422", http.StatusUnprocessableEntity)
	}
	request := stored.ToEntity()
	request.ID = id
	request.Revision = revisionOf(data)
	return &request, nil
}
