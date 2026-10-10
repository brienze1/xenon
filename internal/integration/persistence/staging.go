package persistence

import (
	"context"
	"os"
	"path/filepath"
	"strconv"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence/model"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/fileutils"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/yamlnode"
)

const (
	stagingDirPrefix = ".xenon-import-"
	replacedDir      = ".replaced"
)

type importStagingPersistence struct {
	index adapter.WorkspaceIndexCache
}

type commitJournal struct {
	writes map[string][]byte
	moves  [][2]string
}

func NewImportStagingPersistence(index adapter.WorkspaceIndexCache) adapter.ImportStagingPersistence {
	return &importStagingPersistence{index: index}
}

func (i *importStagingPersistence) Stage(ctx context.Context, staged entity.StagedImport) (*entity.StagedImport, error) {
	logger.Debug(ctx, "Started", staged)

	dir := filepath.Join(staged.Root, stagingDirPrefix+staged.SourceID, staged.ItemKey)
	_ = os.RemoveAll(dir)
	if err := os.MkdirAll(dir, workspaceDirMode); err != nil {
		return nil, errs.FileError(err, "IMP-23500")
	}
	for _, node := range staged.Nodes {
		if err := writeImportNode(dir, node); err != nil {
			removeStaging(dir)
			return nil, err
		}
	}
	written := staged.WithDir(dir)

	logger.Debug(ctx, "Finished", written)
	return &written, nil
}

func (i *importStagingPersistence) Commit(ctx context.Context, staged entity.StagedImport) error {
	logger.Debug(ctx, "Started", staged)
	i.index.Lock(ctx)
	defer i.index.Unlock(ctx)

	index, err := i.index.Current(ctx, staged.Root)
	if err != nil {
		return err
	}
	journal := &commitJournal{writes: map[string][]byte{}}
	if err := i.apply(ctx, index, staged, journal); err != nil {
		journal.undo()
		return err
	}
	_ = os.RemoveAll(filepath.Join(staged.Dir, replacedDir))
	removeStaging(staged.Dir)
	if _, err := i.index.Refresh(ctx, staged.Root); err != nil {
		return err
	}

	logger.Debug(ctx, "Finished", staged)
	return nil
}

func (i *importStagingPersistence) Rollback(ctx context.Context, staged entity.StagedImport) error {
	logger.Debug(ctx, "Started", staged)

	removeStaging(staged.Dir)

	logger.Debug(ctx, "Finished", staged)
	return nil
}

func (i *importStagingPersistence) apply(ctx context.Context, index *entity.TreeIndex, staged entity.StagedImport, journal *commitJournal) error {
	targetDir, err := importTargetDir(index, staged)
	if err != nil {
		return err
	}
	for position, id := range staged.Replaces {
		item, _ := index.Item(id)
		if err := journal.move(item.Path, filepath.Join(staged.Dir, replacedDir, strconv.Itoa(position))); err != nil {
			return err
		}
	}
	for _, node := range staged.Nodes {
		if err := journal.move(filepath.Join(staged.Dir, node.FileName), filepath.Join(targetDir, node.FileName)); err != nil {
			return err
		}
	}
	if err := i.writeOrder(ctx, index, staged, journal); err != nil {
		return err
	}
	if staged.Collection == nil {
		return nil
	}
	item, _ := index.Item(staged.Collection.ID)
	stored := model.NewCollection(*staged.Collection)
	return i.write(ctx, filepath.Join(item.Path, collectionMarker), stored.Fields(), journal)
}

func (i *importStagingPersistence) writeOrder(ctx context.Context, index *entity.TreeIndex, staged entity.StagedImport, journal *commitJournal) error {
	if staged.ParentID == "" {
		_, stored, err := readWorkspace(staged.Root)
		if err != nil {
			return err
		}
		workspace := stored.ToEntity()
		workspace.Order = staged.Order
		updated := model.NewWorkspace(workspace)
		return i.write(ctx, filepath.Join(staged.Root, workspaceFile), updated.Fields(), journal)
	}
	item, _ := index.Item(staged.ParentID)
	parent := model.NewCollection(*staged.Parent)
	fields := parent.Fields()
	if staged.Merge {
		fields = parent.ImportedFields()
	}
	return i.write(ctx, filepath.Join(item.Path, markerFile(item.Kind)), fields, journal)
}

func (i *importStagingPersistence) write(ctx context.Context, path string, fields []yamlnode.Field, journal *commitJournal) error {
	original, err := os.ReadFile(path)
	if err != nil {
		return errs.FileError(err, "IMP-24500")
	}
	document, err := yamlnode.Parse(original)
	if err != nil {
		document = yamlnode.NewDocument()
	}
	yamlnode.Patch(yamlnode.Root(document), fields)
	data, err := yamlnode.Encode(document)
	if err != nil {
		return errs.InternalServerError(err, "IMP-25500", "the file could not be encoded")
	}
	if err := fileutils.WriteFileAtomic(path, data, workspaceFileMode); err != nil {
		return errs.FileError(err, "IMP-26500")
	}
	journal.writes[path] = original
	i.index.RecordWrite(ctx, path, data)
	return nil
}

func (c *commitJournal) move(source, target string) error {
	if err := os.MkdirAll(filepath.Dir(target), workspaceDirMode); err != nil {
		return errs.FileError(err, "IMP-27500")
	}
	if exists(target) {
		return errs.ConflictError("a file with this name already exists", "IMP-28409")
	}
	if err := os.Rename(source, target); err != nil {
		return errs.FileError(err, "IMP-29500")
	}
	c.moves = append(c.moves, [2]string{source, target})
	return nil
}

func (c *commitJournal) undo() {
	for path, original := range c.writes {
		_ = fileutils.WriteFileAtomic(path, original, workspaceFileMode)
	}
	for position := len(c.moves) - 1; position >= 0; position-- {
		_ = os.Rename(c.moves[position][1], c.moves[position][0])
	}
}

func importTargetDir(index *entity.TreeIndex, staged entity.StagedImport) (string, error) {
	if staged.ParentID == "" {
		return filepath.Join(staged.Root, collectionsDir), nil
	}
	item, found := index.Item(staged.ParentID)
	if !found || item.IsRequest() {
		return "", errs.NotFoundError("the target collection or folder does not exist", "IMP-15404")
	}
	return item.Path, nil
}

func writeImportNode(dir string, node entity.ImportNode) error {
	path := filepath.Join(dir, node.FileName)
	if node.IsRequest() {
		stored := model.NewRequest(*node.Request)
		return writeNewDocument(path, stored.Fields())
	}
	if err := os.Mkdir(path, workspaceDirMode); err != nil {
		return errs.FileError(err, "IMP-30500")
	}
	stored := model.NewCollection(node.Container)
	if err := writeNewDocument(filepath.Join(path, markerFile(node.Container.Kind)), stored.Fields()); err != nil {
		return err
	}
	for _, child := range node.Children {
		if err := writeImportNode(path, child); err != nil {
			return err
		}
	}
	return nil
}

func writeNewDocument(path string, fields []yamlnode.Field) error {
	document := yamlnode.NewDocument()
	yamlnode.Patch(yamlnode.Root(document), fields)
	data, err := yamlnode.Encode(document)
	if err != nil {
		return errs.InternalServerError(err, "IMP-31500", "the imported item could not be encoded")
	}
	if err := os.WriteFile(path, data, workspaceFileMode); err != nil {
		return errs.FileError(err, "IMP-32500")
	}
	return nil
}

func removeStaging(dir string) {
	_ = os.RemoveAll(dir)
	_ = os.Remove(filepath.Dir(dir))
}
