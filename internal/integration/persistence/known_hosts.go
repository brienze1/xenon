package persistence

import (
	"context"
	_ "embed"
	"encoding/base64"
	"errors"
	"io/fs"
	"os"
	"sync"

	"golang.org/x/crypto/ssh"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/fileutils"
	"github.com/brienze1/xenon/pkg/logger"
)

const (
	sshDataDir     = "ssh"
	knownHostsFile = "known_hosts"
)

//go:embed known_hosts.seed
var preTrustedHostKeys string

type knownHostsPersistence struct {
	mu sync.Mutex
}

func NewKnownHostsPersistence() adapter.KnownHostsPersistence {
	return &knownHostsPersistence{}
}

func (k *knownHostsPersistence) FindKnownHosts(ctx context.Context) (*entity.KnownHosts, error) {
	logger.Debug(ctx, "Started")

	k.mu.Lock()
	stored, err := readKnownHosts()
	k.mu.Unlock()
	if err != nil {
		return nil, err
	}
	known := entity.KnownHosts{Keys: append(parseKnownHosts(preTrustedHostKeys), parseKnownHosts(stored)...)}

	logger.Debug(ctx, "Finished", map[string]any{"keys": len(known.Keys)})
	return &known, nil
}

func (k *knownHostsPersistence) Trust(ctx context.Context, key entity.HostKey) error {
	logger.Debug(ctx, "Started", key)

	k.mu.Lock()
	defer k.mu.Unlock()
	stored, err := readKnownHosts()
	if err != nil {
		return err
	}
	line := key.Host + " " + key.KeyType + " " + key.Key + "\n"
	if err := os.MkdirAll(appDataPath(sshDataDir), appDataDirMode); err != nil {
		return errs.FileError(err, "ACC-34500")
	}
	if err := fileutils.WriteFileAtomic(knownHostsPath(), []byte(stored+line), appDataFileMode); err != nil {
		return errs.FileError(err, "ACC-35500")
	}

	logger.Debug(ctx, "Finished", key)
	return nil
}

func readKnownHosts() (string, error) {
	data, err := os.ReadFile(knownHostsPath())
	if errors.Is(err, fs.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", errs.FileError(err, "ACC-36500")
	}
	return string(data), nil
}

func parseKnownHosts(content string) []entity.HostKey {
	keys := make([]entity.HostKey, 0)
	rest := []byte(content)
	for {
		_, hosts, key, _, remaining, err := ssh.ParseKnownHosts(rest)
		if err != nil {
			return keys
		}
		rest = remaining
		for _, host := range hosts {
			keys = append(keys, entity.HostKey{
				Host:        host,
				KeyType:     key.Type(),
				Fingerprint: ssh.FingerprintSHA256(key),
				Key:         base64.StdEncoding.EncodeToString(key.Marshal()),
			})
		}
	}
}

func knownHostsPath() string {
	return appDataPath(sshDataDir, knownHostsFile)
}
