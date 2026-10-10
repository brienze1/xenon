package properties

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
)

const (
	defaultVersion    = "0.1.0"
	defaultScryptLogN = 18
)

type properties struct {
	Application *application
	Vault       *vault
}

type application struct {
	DataDir  string
	LogLevel string
	Env      string
	Version  string
	TempDir  string
}

type vault struct {
	ForcePassphrase bool
	ScryptLogN      int
}

func Properties() *properties {
	return &properties{
		Application: &application{
			DataDir:  dataDir(os.Getenv("XENON_DATA_DIR"), runtime.GOOS, os.UserConfigDir),
			LogLevel: os.Getenv("LOG_LEVEL"),
			Env:      os.Getenv("XENON_ENV"),
			Version:  withDefault(os.Getenv("XENON_VERSION"), defaultVersion),
			TempDir:  os.TempDir(),
		},
		Vault: &vault{
			ForcePassphrase: os.Getenv("XENON_FORCE_PASSPHRASE") == "true",
			ScryptLogN:      positiveIntOrDefault(os.Getenv("XENON_SCRYPT_LOG_N"), defaultScryptLogN),
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

func withDefault(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

func positiveIntOrDefault(value string, defaultValue int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return defaultValue
	}
	return parsed
}
