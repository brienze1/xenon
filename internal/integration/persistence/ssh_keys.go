package persistence

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"errors"
	"io/fs"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/fileutils"
	"github.com/brienze1/xenon/pkg/logger"
)

const (
	generatedKeyComment = "xenon"
	publicKeySuffix     = ".pub"
	publicKeyFileMode   = 0o644
)

type sshKeyPersistence struct{}

func NewSshKeyPersistence() adapter.SshKeyPersistence {
	return &sshKeyPersistence{}
}

func (s *sshKeyPersistence) Generate(ctx context.Context, accountId string) (*entity.SshKey, error) {
	logger.Debug(ctx, "Started", map[string]any{"accountId": accountId})

	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, errs.InternalServerError(err, "ACC-37500", "the SSH key could not be generated")
	}
	block, err := ssh.MarshalPrivateKey(private, generatedKeyComment)
	if err != nil {
		return nil, errs.InternalServerError(err, "ACC-38500", "the SSH key could not be encoded")
	}
	sshPublic, _ := ssh.NewPublicKey(public)
	path := appDataPath(sshDataDir, "xenon_"+accountId+"_ed25519")
	authorized := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(sshPublic))) + " " + generatedKeyComment
	if err := os.MkdirAll(appDataPath(sshDataDir), appDataDirMode); err != nil {
		return nil, errs.FileError(err, "ACC-39500")
	}
	if err := fileutils.WriteFileAtomic(path, pem.EncodeToMemory(block), appDataFileMode); err != nil {
		return nil, errs.FileError(err, "ACC-40500")
	}
	if err := fileutils.WriteFileAtomic(path+publicKeySuffix, []byte(authorized+"\n"), publicKeyFileMode); err != nil {
		return nil, errs.FileError(err, "ACC-41500")
	}
	key := entity.SshKey{Path: path, PublicKey: authorized, Fingerprint: ssh.FingerprintSHA256(sshPublic)}

	logger.Debug(ctx, "Finished", key)
	return &key, nil
}

func (s *sshKeyPersistence) Read(ctx context.Context, path, passphrase string) (*entity.SshKey, error) {
	logger.Debug(ctx, "Started", map[string]any{"path": path, "hasPassphrase": passphrase != ""})

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errs.BadRequestError("the SSH key file "+path+" could not be read", "ACC-10400")
	}
	signer, err := ssh.ParsePrivateKey(data)
	var missing *ssh.PassphraseMissingError
	if errors.As(err, &missing) {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(data, []byte(passphrase))
	}
	if err != nil {
		return nil, errs.BadRequestError("the SSH key could not be opened (wrong passphrase or not a private key)", "ACC-42400")
	}
	key := entity.SshKey{
		Path:        path,
		PublicKey:   strings.TrimSpace(string(ssh.MarshalAuthorizedKey(signer.PublicKey()))),
		Fingerprint: ssh.FingerprintSHA256(signer.PublicKey()),
	}

	logger.Debug(ctx, "Finished", key)
	return &key, nil
}

func (s *sshKeyPersistence) DeleteKey(ctx context.Context, path string) error {
	logger.Debug(ctx, "Started", map[string]any{"path": path})

	for _, file := range []string{path, path + publicKeySuffix} {
		if err := os.Remove(file); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return errs.FileError(err, "ACC-43500")
		}
	}

	logger.Debug(ctx, "Finished", map[string]any{"path": path})
	return nil
}
