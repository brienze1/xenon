package persistence

import (
	"archive/zip"
	"context"
	"io"
	"strings"
	"time"

	"golang.org/x/text/unicode/norm"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/fileutils"
	"github.com/brienze1/xenon/pkg/logger"
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

func (e *exportFilePersistence) SaveArchive(ctx context.Context, target string, entries []entity.ExportEntry) error {
	logger.Debug(ctx, "Started", map[string]any{"path": target, "entries": len(entries)})

	modified := time.Now()
	err := fileutils.WriteStreamAtomic(target, exportedFileMode, func(writer io.Writer) error {
		archive := zip.NewWriter(writer)
		for _, entry := range entries {
			if err := ctx.Err(); err != nil {
				_ = archive.Close()
				return err
			}
			if err := writeArchiveEntry(archive, entry, modified); err != nil {
				_ = archive.Close()
				return err
			}
		}
		return archive.Close()
	})
	if err != nil {
		return exportError(ctx, err)
	}

	logger.Debug(ctx, "Finished", map[string]any{"path": target, "entries": len(entries)})
	return nil
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
