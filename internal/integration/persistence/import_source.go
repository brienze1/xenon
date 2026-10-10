package persistence

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/infra/properties"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence/model"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/timeutils"
)

const (
	pastedTextName    = "pasted text"
	zipExtension      = ".zip"
	jsonExtension     = ".json"
	importTempPattern = "xenon-import-"
)

var (
	utf8ByteOrderMark        = []byte{0xEF, 0xBB, 0xBF}
	skippedImportDirs        = []string{"node_modules", ".git"}
	importDocumentExtensions = []string{jsonExtension, ".bru", ".yml", ".yaml"}
)

type importSourcePersistence struct {
	archive adapter.Archive
	sources map[string]entity.ImportSource
	mu      sync.Mutex
}

func NewImportSourcePersistence(archive adapter.Archive) adapter.ImportSourcePersistence {
	return &importSourcePersistence{archive: archive, sources: map[string]entity.ImportSource{}}
}

func (i *importSourcePersistence) Read(ctx context.Context, request entity.ImportRequest) ([]entity.ImportInput, []string, error) {
	logger.Debug(ctx, "Started", request)

	if request.IsText() {
		input, err := readPastedText(request.Text)
		if err != nil {
			return nil, nil, err
		}
		inputs := []entity.ImportInput{*input}
		logger.Debug(ctx, "Finished", map[string]any{"inputs": inputs, "tempDirs": []string{}})
		return inputs, []string{}, nil
	}
	inputs := make([]entity.ImportInput, 0, len(request.Paths))
	tempDirs := make([]string, 0)
	for _, path := range request.Paths {
		input, tempDir, err := i.readPath(path)
		if tempDir != "" {
			tempDirs = append(tempDirs, tempDir)
		}
		if err != nil {
			removeDirs(tempDirs)
			return nil, nil, err
		}
		inputs = append(inputs, *input)
	}

	logger.Debug(ctx, "Finished", map[string]any{"inputs": inputs, "tempDirs": tempDirs})
	return inputs, tempDirs, nil
}

func (i *importSourcePersistence) Save(ctx context.Context, source entity.ImportSource) error {
	logger.Debug(ctx, "Started", source)
	i.mu.Lock()
	defer i.mu.Unlock()

	i.sources[source.ID] = source

	logger.Debug(ctx, "Finished", source)
	return nil
}

func (i *importSourcePersistence) Find(ctx context.Context, id string) (*entity.ImportSource, error) {
	logger.Debug(ctx, "Started", map[string]any{"id": id})
	i.mu.Lock()
	defer i.mu.Unlock()

	source, found := i.sources[id]
	if found && source.IsExpired(timeutils.Now()) {
		i.remove(id)
		found = false
	}
	if !found {
		return nil, errs.NotFoundError("the import source does not exist or has expired; detect it again", "IMP-05404")
	}

	logger.Debug(ctx, "Finished", source)
	return &source, nil
}

func (i *importSourcePersistence) Discard(ctx context.Context, id string) error {
	logger.Debug(ctx, "Started", map[string]any{"id": id})
	i.mu.Lock()
	defer i.mu.Unlock()

	i.remove(id)

	logger.Debug(ctx, "Finished", map[string]any{"id": id})
	return nil
}

func (i *importSourcePersistence) DiscardFiles(ctx context.Context, dirs []string) error {
	logger.Debug(ctx, "Started", map[string]any{"dirs": dirs})

	removeDirs(dirs)

	logger.Debug(ctx, "Finished", map[string]any{"dirs": dirs})
	return nil
}

func (i *importSourcePersistence) remove(id string) {
	removeDirs(i.sources[id].TempDirs)
	delete(i.sources, id)
}

func (i *importSourcePersistence) readPath(path string) (*entity.ImportInput, string, error) {
	name := filepath.Base(path)
	info, err := os.Stat(path)
	if err != nil {
		return nil, "", errs.NotFoundError(name+" cannot be found", "IMP-08404")
	}
	if info.IsDir() {
		input, err := readImportFolder(path, name)
		return input, "", err
	}
	if !strings.EqualFold(filepath.Ext(path), zipExtension) {
		input, err := readImportFile(path, name, info.Size())
		return input, "", err
	}
	tempDir, err := os.MkdirTemp(properties.Properties().Application.TempDir, importTempPattern)
	if err != nil {
		return nil, "", errs.FileError(err, "IMP-20500")
	}
	if err := i.archive.Extract(path, tempDir, properties.Properties().Import.ArchiveLimitBytes); err != nil {
		return nil, tempDir, err
	}
	input, err := readImportFolder(singleFolderInside(tempDir), name)
	return input, tempDir, err
}

func readPastedText(text string) (*entity.ImportInput, error) {
	data := []byte(strings.TrimSpace(text))
	if err := checkImportSize(pastedTextName, int64(len(data))); err != nil {
		return nil, err
	}
	probe := probeDocument(pastedTextName, data, false).Pasted(string(data))
	return &entity.ImportInput{Root: "", Probe: probe, Documents: []entity.ImportDocument{{Path: pastedTextName, Data: data}}}, nil
}

func readImportFile(path, name string, size int64) (*entity.ImportInput, error) {
	data, err := readImportDocument(path, name, size)
	if err != nil {
		return nil, err
	}
	probe := probeDocument(name, data, strings.EqualFold(filepath.Ext(name), jsonExtension))
	return &entity.ImportInput{Root: path, Probe: probe, Documents: []entity.ImportDocument{{Path: name, Data: data}}}, nil
}

func readImportFolder(root, name string) (*entity.ImportInput, error) {
	input := entity.ImportInput{Root: root, Probe: entity.ImportProbe{Name: name, IsFolder: true, Entries: []string{}}, Documents: []entity.ImportDocument{}}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && path != root && isSkippedImportDir(entry.Name()) {
			return filepath.SkipDir
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		input.Probe.Entries = append(input.Probe.Entries, relative)
		if !isImportDocument(relative) {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		data, err := readImportDocument(path, relative, info.Size())
		if err != nil {
			return err
		}
		input.Documents = append(input.Documents, entity.ImportDocument{Path: relative, Data: data})
		return nil
	})
	var custom errs.BaseError
	if errors.As(err, &custom) {
		return nil, custom
	}
	if err != nil {
		return nil, errs.FileError(err, "IMP-21500")
	}
	return &input, nil
}

func probeDocument(name string, data []byte, isJsonFile bool) entity.ImportProbe {
	trimmed := bytes.TrimSpace(data)
	if !isJsonFile && !looksLikeJson(trimmed) {
		return entity.ImportProbe{Name: name}
	}
	var document model.ImportProbe
	if err := json.Unmarshal(trimmed, &document); err != nil {
		var syntax *json.SyntaxError
		return entity.ImportProbe{Name: name, IsMalformed: errors.As(err, &syntax)}
	}
	return document.ToEntity(name)
}

func looksLikeJson(data []byte) bool {
	return bytes.HasPrefix(data, []byte("{")) || bytes.HasPrefix(data, []byte("["))
}

func readImportDocument(path, name string, size int64) ([]byte, error) {
	if err := checkImportSize(name, size); err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errs.FileError(err, "IMP-22500")
	}
	data = bytes.TrimPrefix(data, utf8ByteOrderMark)
	if !utf8.Valid(data) {
		return nil, errs.BadRequestError(filepath.Base(name)+" is not UTF-8 text", "IMP-01400")
	}
	return data, nil
}

func checkImportSize(name string, size int64) error {
	limit := properties.Properties().Import.FileLimitBytes
	if size > limit {
		return errs.ClientError(filepath.Base(name)+" is larger than "+strconv.FormatInt(limit, 10)+" bytes", "IMP-02413", http.StatusRequestEntityTooLarge)
	}
	return nil
}

func isSkippedImportDir(name string) bool {
	return slices.Contains(skippedImportDirs, name)
}

func isImportDocument(relative string) bool {
	if slices.ContainsFunc(strings.Split(relative, "/"), func(segment string) bool { return strings.HasPrefix(segment, ".") }) {
		return false
	}
	return slices.Contains(importDocumentExtensions, strings.ToLower(filepath.Ext(relative)))
}

func singleFolderInside(dir string) string {
	listing := readEntries(dir)
	if len(listing) == 1 && listing[0].IsDir() {
		return filepath.Join(dir, listing[0].Name())
	}
	return dir
}

func removeDirs(dirs []string) {
	for _, dir := range dirs {
		_ = os.RemoveAll(dir)
	}
}
