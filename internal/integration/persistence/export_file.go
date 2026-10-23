package persistence

import (
	"archive/zip"
	"context"
	"encoding/json"
	"encoding/xml"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/text/unicode/norm"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/runexportformat"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence/model"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/fileutils"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/timeutils"
)

const (
	exportedFileMode = 0o644
	cancelledStatus  = 499
	zipUtf8NamesFlag = 0x800
)

type exportFilePersistence struct{}

func NewExportFilePersistence() adapter.ExportFilePersistence {
	return &exportFilePersistence{}
}

func (e *exportFilePersistence) SaveFile(ctx context.Context, target string, data []byte) error {
	logger.Debug(ctx, "Started", map[string]any{"path": target, "bytes": len(data)})

	err := fileutils.WriteStreamAtomic(target, exportedFileMode, func(writer io.Writer) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		_, err := writer.Write(data)
		return err
	})
	if err != nil {
		return exportError(ctx, err)
	}

	logger.Debug(ctx, "Finished", map[string]any{"path": target})
	return nil
}

func (e *exportFilePersistence) SaveRunExport(ctx context.Context, target string, run entity.Run, format enums.RunExportFormat) error {
	logger.Debug(ctx, "Started", map[string]any{"path": target, "runId": run.ID, "format": format})

	data, err := encodeRunExport(run, format)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(target), exportDirMode); err != nil {
		return errs.FileError(err, "EXP-24500")
	}
	if err := e.SaveFile(ctx, target, data); err != nil {
		return err
	}

	logger.Debug(ctx, "Finished", map[string]any{"path": target, "runId": run.ID, "format": format, "bytes": len(data)})
	return nil
}

func (e *exportFilePersistence) SaveArchive(ctx context.Context, target string, entries []entity.ExportEntry, written func(done int)) error {
	logger.Debug(ctx, "Started", map[string]any{"path": target, "entries": len(entries)})

	modified := timeutils.Now()
	err := fileutils.WriteStreamAtomic(target, exportedFileMode, func(writer io.Writer) error {
		archive := zip.NewWriter(writer)
		for index, entry := range entries {
			if err := ctx.Err(); err != nil {
				_ = archive.Close()
				return err
			}
			if err := writeArchiveEntry(archive, entry, modified); err != nil {
				_ = archive.Close()
				return err
			}
			written(index + 1)
		}
		return archive.Close()
	})
	if err != nil {
		return exportError(ctx, err)
	}

	logger.Debug(ctx, "Finished", map[string]any{"path": target, "entries": len(entries)})
	return nil
}

func encodeRunExport(run entity.Run, format enums.RunExportFormat) ([]byte, error) {
	if format == runexportformat.Junit {
		return encodeRunJunit(run)
	}
	data, err := json.MarshalIndent(model.NewRun(run), "", "  ")
	if err != nil {
		return nil, errs.FileError(err, "EXP-25500")
	}
	return append(data, '\n'), nil
}

func encodeRunJunit(run entity.Run) ([]byte, error) {
	encoded, err := xml.MarshalIndent(model.NewRunJunitSuites(run), "", "  ")
	if err != nil {
		return nil, errs.FileError(err, "EXP-26500")
	}
	return append(append([]byte(xml.Header), encoded...), '\n'), nil
}

func writeArchiveEntry(archive *zip.Writer, entry entity.ExportEntry, modified time.Time) error {
	header := &zip.FileHeader{Name: archiveName(entry.Path), Method: zip.Deflate, Modified: modified, Flags: zipUtf8NamesFlag}
	header.SetMode(exportedFileMode)
	file, err := archive.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = file.Write(entry.Data)
	return err
}

func archiveName(name string) string {
	return norm.NFC.String(strings.ReplaceAll(name, `\`, "/"))
}

func exportError(ctx context.Context, err error) error {
	if ctx.Err() != nil {
		return errs.ClientError("the export was cancelled; no file was written", "EXP-05499", cancelledStatus)
	}
	return errs.FileError(err, "EXP-02500")
}
