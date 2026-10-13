package properties

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
)

const (
	defaultVersion            = "0.1.0"
	defaultScryptLogN         = 18
	defaultImportArchiveLimit = 1 << 30
	defaultImportFileLimit    = 200 << 20
	defaultOAuthCallbackMs    = 5 * 60 * 1000
	defaultGitHubApiURL       = "https://api.github.com"
	defaultGitHubLoginURL     = "https://github.com"
)

// gitHubClientID is the GitHub OAuth App client id, set at build time with -ldflags -X.
var gitHubClientID = ""

type properties struct {
	Application *application
	Vault       *vault
	Import      *imports
	OAuth       *oauth
	Git         *git
}

type git struct {
	GitHubClientID string
	GitHubApiURL   string
	GitHubLoginURL string
	SshAuthSock    string
}

type oauth struct {
	CallbackTimeout time.Duration
}

type application struct {
	DataDir  string
	LogLevel string
	Env      string
	Version  string
	TempDir  string
}

type imports struct {
	ArchiveLimitBytes int64
	FileLimitBytes    int64
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
		Import: &imports{
			ArchiveLimitBytes: int64(positiveIntOrDefault(os.Getenv("XENON_IMPORT_ARCHIVE_LIMIT"), defaultImportArchiveLimit)),
			FileLimitBytes:    int64(positiveIntOrDefault(os.Getenv("XENON_IMPORT_FILE_LIMIT"), defaultImportFileLimit)),
		},
		OAuth: &oauth{
			CallbackTimeout: time.Duration(positiveIntOrDefault(os.Getenv("XENON_OAUTH_CALLBACK_TIMEOUT_MS"), defaultOAuthCallbackMs)) * time.Millisecond,
		},
		Git: &git{
			GitHubClientID: withDefault(os.Getenv("XENON_GITHUB_CLIENT_ID"), gitHubClientID),
			GitHubApiURL:   withDefault(os.Getenv("XENON_GITHUB_API_URL"), defaultGitHubApiURL),
			GitHubLoginURL: withDefault(os.Getenv("XENON_GITHUB_LOGIN_URL"), defaultGitHubLoginURL),
			SshAuthSock:    os.Getenv("SSH_AUTH_SOCK"),
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
