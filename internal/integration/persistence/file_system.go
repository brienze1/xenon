package persistence

import (
	"os"
	"path/filepath"

	"github.com/brienze1/xenon/internal/infra/properties"
)

const (
	appDataDirMode  = 0o700
	appDataFileMode = 0o600
	registryFile    = "app.json"
)

func isDirectory(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}

func appDataPath(elements ...string) string {
	return filepath.Join(append([]string{properties.Properties().Application.DataDir}, elements...)...)
}
