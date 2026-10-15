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
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/logger"
)

const (
	updatesFolder       = "updates"
	zipPackageExtension = ".zip"
	extractedFolderMode = 0o755
	extractedFileMask   = 0o777
	probeFileMode       = 0o600
)

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

	logger.Debug(ctx, "Finished", installation, map[string]any{"package": packagePath})
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
			return errs.InternalServerError(err, "UPD-25500", "The installed app could not be set aside")
		}
	}
	if err := u.rename(extracted, installation.Path); err != nil {
		if hadPrevious {
			return errs.InternalServerError(errors.Join(err, u.rename(installation.PreviousPath(), installation.Path)), "UPD-26500", "The update could not replace the installed app")
		}
		return errs.InternalServerError(err, "UPD-27500", "The update could not be moved into the installation folder")
	}
	return nil
}

func probeFolder(installation entity.Installation) error {
	probe, err := os.OpenFile(installation.ProbePath(), os.O_CREATE|os.O_EXCL|os.O_WRONLY, probeFileMode)
	if err != nil {
		return errs.InternalServerError(err, "UPD-28500", "The installation folder is not writable")
	}
	_ = probe.Close()
	if err := os.Remove(installation.ProbePath()); err != nil {
		return errs.InternalServerError(err, "UPD-29500", "The installation folder probe could not be removed")
	}
	return nil
}

func extractPackage(installation entity.Installation, packagePath string) (string, error) {
	root := installation.ExtractionPath()
	if err := os.RemoveAll(root); err != nil {
		return "", errs.InternalServerError(err, "UPD-30500", "The previous extraction folder could not be removed")
	}
	if err := os.MkdirAll(root, extractedFolderMode); err != nil {
		return "", errs.InternalServerError(err, "UPD-31500", "The extraction folder could not be created")
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
	return singleEntry(root, installation)
}

func extractTarGz(packagePath, root string) error {
	file, err := os.Open(packagePath)
	if err != nil {
		return errs.InternalServerError(err, "UPD-32500", "The update package could not be opened")
	}
	defer func() { _ = file.Close() }()
	compressed, err := gzip.NewReader(file)
	if err != nil {
		return errs.InternalServerError(err, "UPD-33500", "The update package could not be decompressed")
	}
	defer func() { _ = compressed.Close() }()
	archive := tar.NewReader(compressed)
	for {
		header, err := archive.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return errs.InternalServerError(err, "UPD-34500", "The update package entries could not be read")
		}
		target, err := safeTarget(root, header.Name)
		if err != nil {
			return err
		}
		if err := writeTarEntry(target, header, archive); err != nil {
			return err
		}
	}
}

func writeTarEntry(target string, header *tar.Header, content io.Reader) error {
	switch header.Typeflag {
	case tar.TypeDir:
		if err := os.MkdirAll(target, extractedFolderMode); err != nil {
			return errs.InternalServerError(err, "UPD-35500", "The update package folder "+header.Name+" could not be created")
		}
		return nil
	case tar.TypeReg:
		return writeFile(target, fs.FileMode(header.Mode)&extractedFileMask, content)
	case tar.TypeSymlink:
		if !linkStaysInside(header.Name, header.Linkname) {
			return errs.UnprocessableEntityError("The update package links outside itself: "+header.Name, "UPD-36422")
		}
		if err := os.MkdirAll(filepath.Dir(target), extractedFolderMode); err != nil {
			return errs.InternalServerError(err, "UPD-37500", "The folder of the update package link "+header.Name+" could not be created")
		}
		if err := os.Symlink(header.Linkname, target); err != nil {
			return errs.InternalServerError(err, "UPD-38500", "The update package link "+header.Name+" could not be created")
		}
		return nil
	default:
		return errs.UnprocessableEntityError("The update package holds an unsupported entry: "+header.Name, "UPD-39422")
	}
}

func extractZip(packagePath, root string) error {
	archive, err := zip.OpenReader(packagePath)
	if err != nil {
		return errs.InternalServerError(err, "UPD-40500", "The update package could not be opened")
	}
	defer func() { _ = archive.Close() }()
	for _, file := range archive.File {
		target, err := safeTarget(root, file.Name)
		if err != nil {
			return err
		}
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(target, extractedFolderMode); err != nil {
				return errs.InternalServerError(err, "UPD-41500", "The update package folder "+file.Name+" could not be created")
			}
			continue
		}
		if !file.Mode().IsRegular() {
			return errs.UnprocessableEntityError("The update package holds an unsupported entry: "+file.Name, "UPD-42422")
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
		return errs.InternalServerError(err, "UPD-43500", "The update package entry "+file.Name+" could not be read")
	}
	defer func() { _ = content.Close() }()
	return writeFile(target, file.Mode()&extractedFileMask|probeFileMode, content)
}

func writeFile(target string, mode fs.FileMode, content io.Reader) error {
	if err := os.MkdirAll(filepath.Dir(target), extractedFolderMode); err != nil {
		return errs.InternalServerError(err, "UPD-44500", "The folder of the update file "+filepath.Base(target)+" could not be created")
	}
	file, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
	if err != nil {
		return errs.InternalServerError(err, "UPD-45500", "The update file "+filepath.Base(target)+" could not be created")
	}
	if _, err := io.Copy(file, content); err != nil {
		_ = file.Close()
		return errs.InternalServerError(err, "UPD-46500", "The update file "+filepath.Base(target)+" could not be written")
	}
	if err := file.Close(); err != nil {
		return errs.InternalServerError(err, "UPD-47500", "The update file "+filepath.Base(target)+" could not be saved")
	}
	return nil
}

func safeTarget(root, name string) (string, error) {
	slashed := strings.ReplaceAll(name, "\\", "/")
	if slashed == "" || strings.HasPrefix(slashed, "/") || filepath.IsAbs(name) || filepath.VolumeName(name) != "" {
		return "", errs.UnprocessableEntityError("The update package holds the absolute entry "+name, "UPD-48422")
	}
	for _, part := range strings.Split(slashed, "/") {
		if part == ".." {
			return "", errs.UnprocessableEntityError("The update package holds the entry "+name+" that leaves the package", "UPD-49422")
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

func singleEntry(root string, installation entity.Installation) (string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return "", errs.InternalServerError(err, "UPD-50500", "The extracted update package could not be listed")
	}
	if len(entries) != 1 {
		return "", errs.UnprocessableEntityError("The update package must hold exactly one entry", "UPD-51422")
	}
	entry := entries[0]
	if !installation.AcceptsPackageEntry(entry.Name(), entry.IsDir(), entry.Type().IsRegular()) {
		return "", errs.UnprocessableEntityError("The update package holds the unexpected entry "+entry.Name(), "UPD-52422")
	}
	return filepath.Join(root, entry.Name()), nil
}

func removeLeftover(ctx context.Context, path string) {
	if err := os.RemoveAll(path); err != nil {
		logger.Warn(ctx, err, "Update leftover not removed yet, retrying at the next start", map[string]any{"path": path})
	}
}
