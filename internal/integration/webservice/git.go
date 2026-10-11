package webservice

import (
	"encoding/base64"
	"errors"
	"io"
	"net"
	"os"
	"strconv"
	"sync"

	"github.com/go-git/go-git/v5/plumbing/transport"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/pkg/errs"
)

const (
	originRemote   = "origin"
	sshProtocol    = "ssh"
	defaultSshPort = 22
	defaultSshUser = "git"
)

var errUntrustedHostKey = errors.New("the SSH host key is not trusted")

type gitWebservice struct {
	sshAgent adapter.SshAgent
	locks    sync.Map
}

func NewGitWebservice(sshAgent adapter.SshAgent) adapter.GitWebservice {
	return &gitWebservice{sshAgent: sshAgent}
}

func (g *gitWebservice) lock(clonePath string) func() {
	value, _ := g.locks.LoadOrStore(clonePath, &sync.Mutex{})
	mutex := value.(*sync.Mutex)
	mutex.Lock()
	return mutex.Unlock
}

func (g *gitWebservice) auth(access entity.GitAccess) (transport.AuthMethod, io.Closer, error) {
	endpoint, err := transport.NewEndpoint(access.Repository)
	if err != nil {
		return nil, nil, errs.BadRequestError("the repository URL is not valid: "+access.Repository, "GIT-22400")
	}
	if endpoint.Protocol != sshProtocol {
		if access.Account.IsSsh() {
			return nil, nil, errs.BadRequestError("an SSH account needs an SSH repository URL (git@host:owner/repo.git)", "GIT-17400")
		}
		return &githttp.BasicAuth{Username: access.Account.BasicAuthUsername(), Password: access.Secret}, io.NopCloser(nil), nil
	}
	if !access.Account.IsSsh() {
		return nil, nil, errs.BadRequestError("a token account needs an HTTPS repository URL", "GIT-18400")
	}
	signers, closer, err := g.signers(access)
	if err != nil {
		return nil, nil, err
	}
	user := endpoint.User
	if user == "" {
		user = defaultSshUser
	}
	callback := &gitssh.PublicKeysCallback{User: user, Callback: func() ([]ssh.Signer, error) { return signers, nil }}
	callback.HostKeyCallback = trustedHostKeys(access.TrustedHostKeys)
	return callback, closer, nil
}

func (g *gitWebservice) signers(access entity.GitAccess) ([]ssh.Signer, io.Closer, error) {
	if access.Account.UsesAgent() {
		signers, closer, err := g.sshAgent.Signers()
		if err != nil {
			return nil, nil, errs.BadRequestError("the SSH agent is not available: "+err.Error(), "GIT-23400")
		}
		return signers, closer, nil
	}
	pem, err := os.ReadFile(access.Account.KeyPath)
	if err != nil {
		return nil, nil, errs.BadRequestError("the SSH key file could not be read: "+access.Account.KeyPath, "GIT-24400")
	}
	signer, err := parseSshKey(pem, access.Secret)
	if err != nil {
		return nil, nil, errs.BadRequestError("the SSH key could not be opened: "+err.Error(), "GIT-25400")
	}
	return []ssh.Signer{signer}, io.NopCloser(nil), nil
}

func parseSshKey(pem []byte, passphrase string) (ssh.Signer, error) {
	if passphrase == "" {
		return ssh.ParsePrivateKey(pem)
	}
	return ssh.ParsePrivateKeyWithPassphrase(pem, []byte(passphrase))
}

func trustedHostKeys(keys []entity.HostKey) ssh.HostKeyCallback {
	return func(_ string, _ net.Addr, key ssh.PublicKey) error {
		presented := base64.StdEncoding.EncodeToString(key.Marshal())
		for _, trusted := range keys {
			if trusted.KeyType == key.Type() && trusted.Key == presented {
				return nil
			}
		}
		return errUntrustedHostKey
	}
}

func sshAddress(endpoint *transport.Endpoint) (string, string) {
	port := endpoint.Port
	if port == 0 {
		port = defaultSshPort
	}
	address := net.JoinHostPort(endpoint.Host, strconv.Itoa(port))
	return address, knownhosts.Normalize(address)
}

func newHostKey(host string, key ssh.PublicKey) entity.HostKey {
	return entity.HostKey{
		Host:        host,
		KeyType:     key.Type(),
		Fingerprint: ssh.FingerprintSHA256(key),
		Key:         base64.StdEncoding.EncodeToString(key.Marshal()),
	}
}
