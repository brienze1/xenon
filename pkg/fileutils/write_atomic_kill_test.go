package fileutils

import (
	"bufio"
	"bytes"
	"math/rand/v2"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

const (
	killTargetEnv = "XENON_ATOMIC_KILL_TARGET"
	killRounds    = 100
	killBodySize  = 256 << 10
)

func killContents() ([]byte, []byte) {
	return bytes.Repeat([]byte("a"), killBodySize), bytes.Repeat([]byte("b"), killBodySize+7)
}

func TestAtomicWriteKillHelper(t *testing.T) {
	target := os.Getenv(killTargetEnv)
	if target == "" {
		t.Skip("runs only as the subprocess of TestWriteFileAtomicSurvivesBeingKilled")
	}
	first, second := killContents()
	for round := 0; ; round++ {
		content := first
		if round%2 == 1 {
			content = second
		}
		if err := WriteFileAtomic(target, content, 0o600); err != nil {
			os.Exit(2)
		}
		if round == 0 {
			_, _ = os.Stdout.WriteString("ready\n")
		}
	}
}

func TestWriteFileAtomicSurvivesBeingKilled(t *testing.T) {
	if os.Getenv(killTargetEnv) != "" {
		t.Skip("already the subprocess")
	}
	first, second := killContents()
	target := filepath.Join(t.TempDir(), "request.yaml")
	if err := os.WriteFile(target, first, 0o600); err != nil {
		t.Fatal(err)
	}
	for round := 1; round <= killRounds; round++ {
		killWriterMidWrite(t, target)
		content, err := os.ReadFile(target)
		if err != nil {
			t.Fatalf("round %d: target unreadable after kill: %v", round, err)
		}
		if !bytes.Equal(content, first) && !bytes.Equal(content, second) {
			t.Fatalf("round %d: target has %d bytes that are neither the old nor the new content", round, len(content))
		}
	}
}

func killWriterMidWrite(t *testing.T, target string) {
	t.Helper()
	command := exec.Command(os.Args[0], "-test.run=^TestAtomicWriteKillHelper$")
	command.Env = append(os.Environ(), killTargetEnv+"="+target)
	output, err := command.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	if _, err := bufio.NewReader(output).ReadString('\n'); err != nil {
		_ = command.Process.Kill()
		_ = command.Wait()
		t.Fatalf("writer never became ready: %v", err)
	}
	time.Sleep(time.Duration(rand.IntN(5000)) * time.Microsecond)
	if err := command.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	_ = command.Wait()
}
