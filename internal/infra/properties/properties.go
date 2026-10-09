package properties

import (
	"os"
	"path/filepath"
	"runtime"
)

var appVersion = "0.1.0"

type properties struct {
	Application *application
}

type application struct {
	DataDir  string
	LogLevel string
	Env      string
	Version  string
}

func Properties() *properties {
	return &properties{
		Application: &application{
			DataDir:  dataDir(os.Getenv("XENON_DATA_DIR"), runtime.GOOS, os.UserConfigDir),
			LogLevel: os.Getenv("LOG_LEVEL"),
			Env:      os.Getenv("XENON_ENV"),
			Version:  appVersion,
		},
	}
}

func dataDir(configured, goos string, userConfigDir func() (string, error)) string {
	if configured != "" {
		return configured
	}
	base, err := userConfigDir()
	if err != nil {
		base = os.TempDir()
	}
	return filepath.Join(base, appFolderName(goos))
}

func appFolderName(goos string) string {
	if goos == "linux" {
		return "xenon"
	}
	return "Xenon"
}
