package persistence

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/hostkeystatus"
)

func TestPreTrustedHostKeysMatchThePublishedFingerprints(t *testing.T) {
	t.Setenv("XENON_DATA_DIR", t.TempDir())
	known, err := NewKnownHostsPersistence().FindKnownHosts(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	got := make([]string, 0)
	for _, key := range known.Keys {
		got = append(got, key.Host+" "+key.Fingerprint)
	}
	sort.Strings(got)
	want := []string{
		"bitbucket.org SHA256:46OSHA1Rmj8E8ERTC6xkNcmGOw9oFxYr0WF6zWW8l1E",
		"bitbucket.org SHA256:FC73VB6C4OQLSCrjEayhMp9UMxS97caD/Yyi2bhW/J0",
		"bitbucket.org SHA256:ybgmFkzwOSotHTHLJgHO0QN8L0xErw6vd0VhFA9m3SM",
		"github.com SHA256:+DiY3wvvV6TuJJhbpZisF/zLDA0zPMSvHdkr4UvCOqU",
		"github.com SHA256:p2QAMXNIC1TJYWeIOttrVc98/R1BUFWu3/LiyKgUfQM",
		"github.com SHA256:uNiVztksCsDhcc0u9e8BujQXVUpKZIDTMczCvj3tD2s",
		"gitlab.com SHA256:HbW3g8zUjNSksFbqTiUWPWg2Bq1x8xdGUrliXFzSnUw",
		"gitlab.com SHA256:ROQFvPThGrW4RuWLoL9tq9I9zJ42fK4XywyRtbOz/EQ",
		"gitlab.com SHA256:eUXGGm1YGsMAS7vkcx6JOJdOGHPem5gQp4taiCfCLB8",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("pre-trusted keys:\n%s", strings.Join(got, "\n"))
	}
	github := known.KeysFor("github.com")[0]
	if known.StatusOf(github) != hostkeystatus.Trusted {
		t.Fatal("github.com must be trusted without a prompt")
	}
}

func TestTrustedHostKeysAreAppendedAndReadBack(t *testing.T) {
	directory := t.TempDir()
	t.Setenv("XENON_DATA_DIR", directory)
	hosts := NewKnownHostsPersistence()
	key := entity.HostKey{Host: "[127.0.0.1]:2222", KeyType: "ssh-ed25519", Key: "AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl"}
	for range 2 {
		if err := hosts.Trust(context.Background(), key); err != nil {
			t.Fatal(err)
		}
	}
	content, _ := os.ReadFile(filepath.Join(directory, "ssh", "known_hosts"))
	if strings.Count(string(content), "[127.0.0.1]:2222 ssh-ed25519 ") != 2 {
		t.Fatalf("known_hosts = %q", content)
	}
	known, err := hosts.FindKnownHosts(context.Background())
	if err != nil || len(known.KeysFor("[127.0.0.1]:2222")) != 2 || known.StatusOf(key) != hostkeystatus.Trusted {
		t.Fatalf("read back %v, %v", known, err)
	}
	if err := os.WriteFile(filepath.Join(directory, "ssh", "known_hosts"), []byte("not a key line\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if known, err := hosts.FindKnownHosts(context.Background()); err != nil || len(known.Keys) != 9 {
		t.Fatalf("garbage lines must be ignored: %v %v", known, err)
	}
}
