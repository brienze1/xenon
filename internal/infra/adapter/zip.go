package adapter

import (
	"archive/zip"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/pkg/errs"
)

const (
	extractedDirMode  = 0o700
	extractedFileMode = 0o600
)

type zipArchive struct{}

func NewArchive() adapter.Archive {
	return &zipArchive{}
}

func (z *zipArchive) Extract(archivePath, targetDir string, limit int64) error {
	name := filepath.Base(archivePath)
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return errs.BadRequestError(name+" is not a valid zip archive", "IMP-18400")
	}
	defer func() { _ = reader.Close() }()
	for _, file := range reader.File {
		if isUnsafeEntry(file) {
			return errs.BadRequestError(name+" contains the unsafe entry "+file.Name, "IMP-03400")
		}
	}
	remaining := limit
	for _, file := range reader.File {
		written, err := extractEntry(file, targetDir, remaining)
		if err != nil {
			return errs.FileError(err, "IMP-19500")
		}
		remaining -= written
		if remaining < 0 {
			return errs.ClientError(name+" unpacks to more than "+strconv.FormatInt(limit, 10)+" bytes", "IMP-02413", http.StatusRequestEntityTooLarge)
		}
	}
	return nil
}

func isUnsafeEntry(file *zip.File) bool {
	name := strings.ReplaceAll(file.Name, `\`, "/")
	if strings.HasPrefix(name, "/") || filepath.VolumeName(file.Name) != "" || file.Mode()&os.ModeSymlink != 0 {
		return true
	}
	for _, segment := range strings.Split(name, "/") {
		if segment == ".." {
			return true
		}
	}
	return false
}

func extractEntry(file *zip.File, targetDir string, remaining int64) (int64, error) {
	path := filepath.Join(targetDir, filepath.FromSlash(strings.ReplaceAll(file.Name, `\`, "/")))
	if file.FileInfo().IsDir() {
		return 0, os.MkdirAll(path, extractedDirMode)
	}
	if err := os.MkdirAll(filepath.Dir(path), extractedDirMode); err != nil {
		return 0, err
	}
	source, err := file.Open()
	if err != nil {
		return 0, err
	}
	defer func() { _ = source.Close() }()
	target, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, extractedFileMode)
	if err != nil {
		return 0, err
	}
	defer func() { _ = target.Close() }()
	return io.Copy(target, io.LimitReader(source, remaining+1))
}
