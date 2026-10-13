package mock

import (
	"errors"
	"io"
	"sync"

	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

var errSshAgentNotRunning = errors.New("no SSH agent is running")

type SshAgent struct {
	keyring agent.Agent
	running bool
	mu      sync.Mutex
}

func NewSshAgent() *SshAgent {
	return &SshAgent{keyring: agent.NewKeyring()}
}

func (s *SshAgent) Signers() ([]gossh.Signer, io.Closer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.running {
		return nil, nil, errSshAgentNotRunning
	}
	signers, err := s.keyring.Signers()
	return signers, io.NopCloser(nil), err
}

func (s *SshAgent) Hold(key any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.running = true
	return s.keyring.Add(agent.AddedKey{PrivateKey: key})
}

func (s *SshAgent) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.running = true
}

func (s *SshAgent) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.keyring = agent.NewKeyring()
	s.running = false
}
