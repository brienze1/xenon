//go:build !windows

package adapter

import (
	"crypto/ed25519"
	"crypto/rand"
	"net"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/crypto/ssh/agent"
)

func TestSshAgentWithoutSocketFails(t *testing.T) {
	t.Setenv("SSH_AUTH_SOCK", "")
	signers, closer, err := NewSshAgent().Signers()
	if err == nil || signers != nil || closer != nil {
		t.Fatalf("Signers() = %v, %v, %v; want an error", signers, closer, err)
	}
}

func TestSshAgentWithUnreachableSocketFails(t *testing.T) {
	t.Setenv("SSH_AUTH_SOCK", filepath.Join(t.TempDir(), "missing.sock"))
	if _, _, err := NewSshAgent().Signers(); err == nil {
		t.Fatal("Signers() on a missing socket must fail")
	}
}

func TestSshAgentListsTheKeysOfTheAgentOnTheSocket(t *testing.T) {
	directory, err := os.MkdirTemp("", "agt")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(directory) }()
	socket := filepath.Join(directory, "a.sock")
	listener, err := net.Listen("unix", socket)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = listener.Close() }()
	keyring := agent.NewKeyring()
	_, private, _ := ed25519.GenerateKey(rand.Reader)
	if err := keyring.Add(agent.AddedKey{PrivateKey: private}); err != nil {
		t.Fatal(err)
	}
	go func() {
		connection, err := listener.Accept()
		if err == nil {
			_ = agent.ServeAgent(keyring, connection)
		}
	}()
	t.Setenv("SSH_AUTH_SOCK", socket)
	signers, closer, err := NewSshAgent().Signers()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = closer.Close() }()
	if len(signers) != 1 {
		t.Fatalf("Signers() returned %d keys, want 1", len(signers))
	}
}
