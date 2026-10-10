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
	t.Setenv("XENON_VERSION", "")
	first := Properties().Application
	if first.DataDir != "/tmp/first" || first.LogLevel != "debug" || first.Env != "test" || first.Version != defaultVersion || first.TempDir != os.TempDir() {
		t.Fatalf("unexpected properties %+v", first)
	}
	t.Setenv("XENON_DATA_DIR", "/tmp/second")
	if got := Properties().Application.DataDir; got != "/tmp/second" {
		t.Fatalf("DataDir = %q, want /tmp/second", got)
	}
	t.Setenv("XENON_VERSION", "9.9.9")
	if got := Properties().Application.Version; got != "9.9.9" {
		t.Fatalf("Version = %q, want 9.9.9", got)
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

func TestVaultPropertiesDefaultAndOverride(t *testing.T) {
	t.Setenv("XENON_FORCE_PASSPHRASE", "")
	t.Setenv("XENON_SCRYPT_LOG_N", "")
	defaults := Properties().Vault
	if defaults.ForcePassphrase || defaults.ScryptLogN != defaultScryptLogN {
		t.Fatalf("unexpected defaults %+v", defaults)
	}
	t.Setenv("XENON_FORCE_PASSPHRASE", "true")
	t.Setenv("XENON_SCRYPT_LOG_N", "10")
	overridden := Properties().Vault
	if !overridden.ForcePassphrase || overridden.ScryptLogN != 10 {
		t.Fatalf("unexpected overrides %+v", overridden)
	}
}

func TestPositiveIntOrDefault(t *testing.T) {
	cases := map[string]int{"": 18, "abc": 18, "0": 18, "-3": 18, "1": 1, "12": 12}
	for raw, want := range cases {
		if got := positiveIntOrDefault(raw, 18); got != want {
			t.Errorf("positiveIntOrDefault(%q) = %d, want %d", raw, got, want)
		}
	}
}
