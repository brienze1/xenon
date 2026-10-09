package properties

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestPropertiesReadsEnvironmentOnEveryCall(t *testing.T) {
	t.Setenv("XENON_DATA_DIR", "/tmp/first")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("XENON_ENV", "test")
	first := Properties().Application
	if first.DataDir != "/tmp/first" || first.LogLevel != "debug" || first.Env != "test" || first.Version != appVersion {
		t.Fatalf("unexpected properties %+v", first)
	}
	t.Setenv("XENON_DATA_DIR", "/tmp/second")
	if got := Properties().Application.DataDir; got != "/tmp/second" {
		t.Fatalf("DataDir = %q, want /tmp/second", got)
	}
}

func TestDataDirDefaultsPerOperatingSystem(t *testing.T) {
	configDir := func() (string, error) { return "/config", nil }
	cases := map[string]string{
		"linux":   filepath.Join("/config", "xenon"),
		"darwin":  filepath.Join("/config", "Xenon"),
		"windows": filepath.Join("/config", "Xenon"),
	}
	for goos, want := range cases {
		if got := dataDir("", goos, configDir); got != want {
			t.Errorf("dataDir(%s) = %q, want %q", goos, got, want)
		}
	}
}

func TestDataDirFallsBackToTempDirWhenConfigDirIsUnknown(t *testing.T) {
	failing := func() (string, error) { return "", errors.New("no home") }
	want := filepath.Join(os.TempDir(), "Xenon")
	if got := dataDir("", "darwin", failing); got != want {
		t.Fatalf("dataDir = %q, want %q", got, want)
	}
}
