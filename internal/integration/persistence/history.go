package persistence

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"sync"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/historyfilter"
	"github.com/brienze1/xenon/internal/domain/enums/historysource"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence/model"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/fileutils"
	"github.com/brienze1/xenon/pkg/logger"
)

const historyFile = "history.jsonl"

var lineBreak = []byte{'\n'}

type historyPersistence struct {
	mu sync.Mutex
}

func NewHistoryPersistence() adapter.HistoryPersistence {
	return &historyPersistence{}
}

func (h *historyPersistence) Save(ctx context.Context, localId string, entry entity.HistoryEntry, caps entity.HistoryCaps) error {
	logger.Debug(ctx, "Started", entry, map[string]any{"localId": localId, "caps": caps})

	h.mu.Lock()
	defer h.mu.Unlock()
	if err := os.MkdirAll(appDataPath(workspacesDataDir, localId), appDataDirMode); err != nil {
		return errs.FileError(err, "HIS-05500")
	}
	line, _ := json.Marshal(model.NewHistoryEntry(entry))
	if err := appendHistoryLine(historyPath(localId), line, caps); err != nil {
		return err
	}

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "id": entry.ID})
	return nil
}

func (h *historyPersistence) Find(ctx context.Context, localId string, caps entity.HistoryCaps) (*entity.History, error) {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId, "caps": caps})

	h.mu.Lock()
	lines, err := compactedHistoryLines(historyPath(localId), caps)
	h.mu.Unlock()
	if err != nil {
		return nil, err
	}
	entries := make([]entity.HistoryEntry, 0, len(lines))
	for _, line := range lines {
		var decoded model.HistoryEntry
		if err := json.Unmarshal(line, &decoded); err != nil {
			logger.Warn(ctx, err, "Skipping an unreadable history line", map[string]any{"localId": localId})
			continue
		}
		entries = append(entries, decoded.ToEntity())
	}
	history := entity.NewHistory(entries, caps)

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "entries": len(entries)})
	return &history, nil
}

func (h *historyPersistence) FindEntry(ctx context.Context, localId, id string, caps entity.HistoryCaps) (*entity.HistoryEntry, error) {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId, "id": id})

	history, err := h.Find(ctx, localId, caps)
	if err != nil {
		return nil, err
	}
	entry, found := history.Find(id)
	if !found {
		return nil, errs.NotFoundError("history entry not found", "HIS-03404")
	}

	logger.Debug(ctx, "Finished", entry)
	return &entry, nil
}

func (h *historyPersistence) Clear(ctx context.Context, localId string, source enums.HistoryFilter) error {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId, "source": source})

	h.mu.Lock()
	defer h.mu.Unlock()
	if err := os.MkdirAll(appDataPath(workspacesDataDir, localId), appDataDirMode); err != nil {
		return errs.FileError(err, "HIS-09500")
	}
	kept, err := historyLinesKeptOnClear(historyPath(localId), source)
	if err != nil {
		return err
	}
	if err := fileutils.WriteFileAtomic(historyPath(localId), joinHistoryLines(kept), appDataFileMode); err != nil {
		return errs.FileError(err, "HIS-08500")
	}

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "source": source})
	return nil
}

func historyPath(localId string) string {
	return appDataPath(workspacesDataDir, localId, historyFile)
}

func readHistoryFile(path string) ([]byte, error) {
	stored, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return []byte{}, nil
	}
	if err != nil {
		return nil, errs.FileError(err, "HIS-04500")
	}
	return stored, nil
}

func historyLines(stored []byte) [][]byte {
	lines := [][]byte{}
	for _, line := range bytes.Split(stored, lineBreak) {
		if len(bytes.TrimSpace(line)) > 0 {
			lines = append(lines, line)
		}
	}
	return lines
}

func historyLineSource(line []byte) enums.HistorySource {
	var decoded model.HistorySource
	if err := json.Unmarshal(line, &decoded); err != nil {
		return historysource.User
	}
	return decoded.ToEntity()
}

func historyLinesKeptOnClear(path string, source enums.HistoryFilter) ([][]byte, error) {
	if source == historyfilter.All {
		return [][]byte{}, nil
	}
	stored, err := readHistoryFile(path)
	if err != nil {
		return nil, err
	}
	return entity.WithoutHistorySource(historyLines(stored), historyLineSource, historysource.GetHistorySource(source.String())), nil
}

func compactedHistoryLines(path string, caps entity.HistoryCaps) ([][]byte, error) {
	stored, err := readHistoryFile(path)
	if err != nil {
		return nil, err
	}
	kept, compact := entity.CompactHistory(historyLines(stored), historyLineSource, caps)
	if compact {
		if err := compactHistory(path, kept); err != nil {
			return nil, err
		}
	}
	return kept, nil
}

func appendHistoryLine(path string, line []byte, caps entity.HistoryCaps) error {
	stored, err := readHistoryFile(path)
	if err != nil {
		return err
	}
	if kept, compact := entity.CompactHistory(append(historyLines(stored), line), historyLineSource, caps); compact {
		return compactHistory(path, kept)
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, appDataFileMode)
	if err != nil {
		return errs.FileError(err, "HIS-06500")
	}
	_, writeErr := file.Write(append(separatorAfter(stored), append(line, lineBreak...)...))
	if err := errors.Join(writeErr, file.Close()); err != nil {
		return errs.FileError(err, "HIS-10500")
	}
	return nil
}

func separatorAfter(stored []byte) []byte {
	if len(stored) == 0 || bytes.HasSuffix(stored, lineBreak) {
		return []byte{}
	}
	return lineBreak
}

func joinHistoryLines(lines [][]byte) []byte {
	if len(lines) == 0 {
		return []byte{}
	}
	return append(bytes.Join(lines, lineBreak), lineBreak...)
}

func compactHistory(path string, kept [][]byte) error {
	if err := fileutils.WriteFileAtomic(path, joinHistoryLines(kept), appDataFileMode); err != nil {
		return errs.FileError(err, "HIS-07500")
	}
	return nil
}
