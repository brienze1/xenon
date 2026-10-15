package properties

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
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

func TestOAuthCallbackTimeoutDefaultAndOverride(t *testing.T) {
	t.Setenv("XENON_OAUTH_CALLBACK_TIMEOUT_MS", "")
	if got := Properties().OAuth.CallbackTimeout; got != 5*time.Minute {
		t.Fatalf("default CallbackTimeout = %s, want 5m", got)
	}
	t.Setenv("XENON_OAUTH_CALLBACK_TIMEOUT_MS", "250")
	if got := Properties().OAuth.CallbackTimeout; got != 250*time.Millisecond {
		t.Fatalf("CallbackTimeout = %s, want 250ms", got)
	}
}

func TestGitPropertiesDefaultAndOverride(t *testing.T) {
	t.Setenv("XENON_GITHUB_CLIENT_ID", "")
	t.Setenv("XENON_GITHUB_API_URL", "")
	t.Setenv("XENON_GITHUB_LOGIN_URL", "")
	t.Setenv("SSH_AUTH_SOCK", "")
	defaults := Properties().Git
	if defaults.GitHubClientID != gitHubClientID || defaults.GitHubApiURL != defaultGitHubApiURL || defaults.GitHubLoginURL != defaultGitHubLoginURL || defaults.SshAuthSock != "" {
		t.Fatalf("unexpected defaults %+v", defaults)
	}
	t.Setenv("XENON_GITHUB_CLIENT_ID", "Iv1.test")
	t.Setenv("XENON_GITHUB_API_URL", "http://127.0.0.1:1/api")
	t.Setenv("XENON_GITHUB_LOGIN_URL", "http://127.0.0.1:1")
	t.Setenv("SSH_AUTH_SOCK", "/tmp/agent.sock")
	overridden := Properties().Git
	if overridden.GitHubClientID != "Iv1.test" || overridden.GitHubApiURL != "http://127.0.0.1:1/api" || overridden.GitHubLoginURL != "http://127.0.0.1:1" || overridden.SshAuthSock != "/tmp/agent.sock" {
		t.Fatalf("unexpected overrides %+v", overridden)
	}
}

func TestMCPPortOverrideReplacesOnlyWhenValid(t *testing.T) {
	cases := map[string]*int{"": nil, "nope": nil, "0": nil, "-1": nil, "65536": nil, "47900": intPointer(47900)}
	for value, want := range cases {
		t.Setenv("XENON_MCP_PORT", value)
		got := Properties().Application.MCPPortOverride
		if (got == nil) != (want == nil) || got != nil && *got != *want {
			t.Errorf("XENON_MCP_PORT=%q: got %v, want %v", value, got, want)
		}
	}
}

func TestExecutablePathFallsBackToTheAppName(t *testing.T) {
	if got := executablePath(func() (string, error) { return "/Applications/Xenon.app/Contents/MacOS/xenon", nil }); got != "/Applications/Xenon.app/Contents/MacOS/xenon" {
		t.Errorf("executablePath() = %q", got)
	}
	if got := executablePath(func() (string, error) { return "", errors.New("unknown") }); got == "" {
		t.Error("executablePath() must never be empty")
	}
	if Properties().Application.Executable == "" {
		t.Error("the running executable must be known")
	}
}

func intPointer(value int) *int {
	return &value
}

func TestUpdatePropertiesDefaultAndOverride(t *testing.T) {
	t.Setenv("XENON_UPDATE_API_URL", "")
	t.Setenv("XENON_UPDATE_PLATFORM", "")
	t.Setenv("XENON_INSTALL_PATH", "")
	defaults := Properties().Update
	if defaults.ApiURL != defaultUpdateApiURL || defaults.Platform == "" || defaults.InstallPath == "" || defaults.Installable {
		t.Fatalf("unexpected defaults %+v", defaults)
	}
	t.Setenv("XENON_UPDATE_API_URL", "http://127.0.0.1:1")
	t.Setenv("XENON_UPDATE_PLATFORM", "linux-amd64")
	t.Setenv("XENON_INSTALL_PATH", "/tmp/apps/xenon")
	overridden := Properties().Update
	if overridden.ApiURL != "http://127.0.0.1:1" || overridden.Platform != "linux-amd64" || overridden.InstallPath != "/tmp/apps/xenon" || !overridden.Installable {
		t.Fatalf("unexpected overrides %+v", overridden)
	}
}

func TestUpdatePlatformPerOperatingSystem(t *testing.T) {
	cases := map[[2]string]string{{"darwin", "arm64"}: "darwin-universal", {"darwin", "amd64"}: "darwin-universal", {"windows", "amd64"}: "windows-amd64", {"linux", "arm64"}: "linux-arm64"}
	for input, want := range cases {
		if got := updatePlatform(input[0], input[1]); got != want {
			t.Errorf("updatePlatform(%v) = %q, want %q", input, got, want)
		}
	}
}

func TestInstallPathIsTheBundleOrTheExecutable(t *testing.T) {
	bundled := filepath.Join("/Applications", "Xenon.app", "Contents", "MacOS", "xenon")
	if got := installPath(bundled); got != filepath.Join("/Applications", "Xenon.app") {
		t.Errorf("installPath(bundle) = %q", got)
	}
	plain := filepath.Join("/opt", "xenon", "xenon")
	if got := installPath(plain); got != plain {
		t.Errorf("installPath(plain) = %q", got)
	}
}

func TestVersionComesFromTheBuildWhenNotOverridden(t *testing.T) {
	t.Setenv("XENON_VERSION", "")
	previous := version
	t.Cleanup(func() { version = previous })
	version = "0.3.2"
	application := Properties().Application
	if application.Version != "0.3.2" || !application.Packaged {
		t.Fatalf("unexpected application %+v", application)
	}
	if !Properties().Update.Installable {
		t.Fatal("a packaged build is installable")
	}
}
