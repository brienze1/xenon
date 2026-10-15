package persistence

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/updatefailure"
	"github.com/brienze1/xenon/internal/infra/properties"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/pkg/logger"
)

const (
	updatesFolder       = "updates"
	zipPackageExtension = ".zip"
	bundleSuffix        = ".app"
	extractedFolderMode = 0o755
	extractedFileMask   = 0o777
	probeFileMode       = 0o600
)

var plainExecutables = map[string]bool{"xenon": true, "xenon.exe": true}

type updateInstallPersistence struct {
	rename func(from, to string) error
}

func NewUpdateInstallPersistence() adapter.UpdateInstallPersistence {
	return &updateInstallPersistence{rename: os.Rename}
}

func (u *updateInstallPersistence) InstallUpdate(ctx context.Context, installation entity.Installation, packagePath string) error {
	logger.Debug(ctx, "Started", installation, map[string]any{"package": packagePath})

	if err := probeFolder(installation); err != nil {
		return entity.NewUnwritableFolderFailure(installation.Folder(), err)
	}
	extracted, err := extractPackage(installation, packagePath)
	if err != nil {
		_ = os.RemoveAll(installation.ExtractionPath())
		return entity.NewUpdateFailure(updatefailure.InstallFailed, err)
	}
	if err := u.swap(installation, extracted); err != nil {
		_ = os.RemoveAll(installation.ExtractionPath())
		return entity.NewUpdateFailure(updatefailure.InstallFailed, err)
	}
	removeLeftover(ctx, installation.PreviousPath())
	removeLeftover(ctx, installation.ExtractionPath())
	removeLeftover(ctx, packagePath)

	logger.Debug(ctx, "Finished", installation)
	return nil
}

func (u *updateInstallPersistence) CleanUpdateLeftovers(ctx context.Context, installation entity.Installation) {
	logger.Debug(ctx, "Started", installation)

	if installation.HasPath() {
		removeLeftover(ctx, installation.PreviousPath())
		removeLeftover(ctx, installation.ExtractionPath())
		removeLeftover(ctx, installation.ProbePath())
	}
	removeLeftover(ctx, filepath.Join(properties.Properties().Application.DataDir, updatesFolder))

	logger.Debug(ctx, "Finished", installation)
}

func (u *updateInstallPersistence) swap(installation entity.Installation, extracted string) error {
	_ = os.RemoveAll(installation.PreviousPath())
	_, err := os.Lstat(installation.Path)
	hadPrevious := err == nil
	if hadPrevious {
		if err := u.rename(installation.Path, installation.PreviousPath()); err != nil {
			return err
		}
	}
	if err := u.rename(extracted, installation.Path); err != nil {
		if hadPrevious {
			return errors.Join(err, u.rename(installation.PreviousPath(), installation.Path))
		}
		return err
	}
	return nil
}

func probeFolder(installation entity.Installation) error {
	probe, err := os.OpenFile(installation.ProbePath(), os.O_CREATE|os.O_EXCL|os.O_WRONLY, probeFileMode)
	if err != nil {
		return err
	}
	_ = probe.Close()
	return os.Remove(installation.ProbePath())
}

func extractPackage(installation entity.Installation, packagePath string) (string, error) {
	root := installation.ExtractionPath()
	if err := os.RemoveAll(root); err != nil {
		return "", err
	}
	if err := os.MkdirAll(root, extractedFolderMode); err != nil {
		return "", err
	}
	var err error
	if strings.HasSuffix(packagePath, zipPackageExtension) {
		err = extractZip(packagePath, root)
	} else {
		err = extractTarGz(packagePath, root)
	}
	if err != nil {
		return "", err
	}
	return singleEntry(root, installation.Platform.IsBundle())
}

func extractTarGz(packagePath, root string) error {
	file, err := os.Open(packagePath)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	compressed, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer func() { _ = compressed.Close() }()
	archive := tar.NewReader(compressed)
	for {
		header, err := archive.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		target, err := safeTarget(root, header.Name)
		if err != nil {
			return err
		}
		if err := writeTarEntry(root, target, header, archive); err != nil {
			return err
		}
	}
}

func writeTarEntry(root, target string, header *tar.Header, content io.Reader) error {
	switch header.Typeflag {
	case tar.TypeDir:
		return os.MkdirAll(target, extractedFolderMode)
	case tar.TypeReg:
		return writeFile(target, fs.FileMode(header.Mode)&extractedFileMask, content)
	case tar.TypeSymlink:
		if !linkStaysInside(header.Name, header.Linkname) {
			return errors.New("the update package links outside itself: " + header.Name)
		}
		if err := os.MkdirAll(filepath.Dir(target), extractedFolderMode); err != nil {
			return err
		}
		return os.Symlink(header.Linkname, target)
	default:
		return errors.New("the update package holds an unsupported entry: " + header.Name)
	}
}

func extractZip(packagePath, root string) error {
	archive, err := zip.OpenReader(packagePath)
	if err != nil {
		return err
	}
	defer func() { _ = archive.Close() }()
	for _, file := range archive.File {
		target, err := safeTarget(root, file.Name)
		if err != nil {
			return err
		}
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(target, extractedFolderMode); err != nil {
				return err
			}
			continue
		}
		if !file.Mode().IsRegular() {
			return errors.New("the update package holds an unsupported entry: " + file.Name)
		}
		if err := writeZipEntry(file, target); err != nil {
			return err
		}
	}
	return nil
}

func writeZipEntry(file *zip.File, target string) error {
	content, err := file.Open()
	if err != nil {
		return err
	}
	defer func() { _ = content.Close() }()
	return writeFile(target, file.Mode()&extractedFileMask|probeFileMode, content)
}

func writeFile(target string, mode fs.FileMode, content io.Reader) error {
	if err := os.MkdirAll(filepath.Dir(target), extractedFolderMode); err != nil {
		return err
	}
	file, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(file, content); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

func safeTarget(root, name string) (string, error) {
	slashed := strings.ReplaceAll(name, "\\", "/")
	if slashed == "" || strings.HasPrefix(slashed, "/") || filepath.IsAbs(name) || filepath.VolumeName(name) != "" {
		return "", errors.New("the update package holds the unsafe entry " + name)
	}
	for _, part := range strings.Split(slashed, "/") {
		if part == ".." {
			return "", errors.New("the update package holds the unsafe entry " + name)
		}
	}
	return filepath.Join(root, filepath.FromSlash(slashed)), nil
}

func linkStaysInside(name, link string) bool {
	if link == "" || filepath.IsAbs(link) || strings.HasPrefix(link, "/") {
		return false
	}
	resolved := filepath.Clean(filepath.Join(filepath.Dir(filepath.FromSlash(name)), filepath.FromSlash(link)))
	return resolved != ".." && !strings.HasPrefix(resolved, ".."+string(filepath.Separator))
}

func singleEntry(root string, bundle bool) (string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return "", err
	}
	if len(entries) != 1 {
		return "", errors.New("the update package must hold exactly one entry")
	}
	entry := entries[0]
	switch {
	case bundle && entry.IsDir() && strings.HasSuffix(entry.Name(), bundleSuffix):
	case !bundle && entry.Type().IsRegular() && plainExecutables[entry.Name()]:
	default:
		return "", errors.New("the update package holds the unexpected entry " + entry.Name())
	}
	return filepath.Join(root, entry.Name()), nil
}

func removeLeftover(ctx context.Context, path string) {
	if err := os.RemoveAll(path); err != nil {
		logger.Warn(ctx, err, "Update leftover not removed yet, retrying at the next start", map[string]any{"path": path})
	}
}
