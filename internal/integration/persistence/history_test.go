package persistence

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
)

func TestAppendHistoryLineKeepsOneEntryPerLine(t *testing.T) {
	cases := map[string]string{
		"":                  "{\"n\":2}\n",
		"{\"n\":1}":         "{\"n\":1}\n{\"n\":2}\n",
		"{\"n\":1}\n":       "{\"n\":1}\n{\"n\":2}\n",
		"{\"n\":1}\n\n  \n": "{\"n\":1}\n\n  \n{\"n\":2}\n",
	}
	for existing, want := range cases {
		path := filepath.Join(t.TempDir(), historyFile)
		if existing != "" {
			if err := os.WriteFile(path, []byte(existing), appDataFileMode); err != nil {
				t.Fatal(err)
			}
		}
		if err := appendHistoryLine(path, []byte(`{"n":2}`)); err != nil {
			t.Fatalf("appendHistoryLine() error = %v", err)
		}
		got, _ := os.ReadFile(path)
		if string(got) != want {
			t.Errorf("existing %q: got %q, want %q", existing, got, want)
		}
	}
}

func TestAppendHistoryLineCompactsToTheNewestEntries(t *testing.T) {
	path := filepath.Join(t.TempDir(), historyFile)
	lines := make([]string, 0, entity.HistoryCompactionThreshold)
	for index := 1; index <= entity.HistoryCompactionThreshold; index++ {
		lines = append(lines, strconv.Itoa(index))
	}
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), appDataFileMode); err != nil {
		t.Fatal(err)
	}
	if err := appendHistoryLine(path, []byte("1201")); err != nil {
		t.Fatalf("appendHistoryLine() error = %v", err)
	}
	got, _ := os.ReadFile(path)
	kept := strings.Split(strings.TrimSuffix(string(got), "\n"), "\n")
	if len(kept) != entity.HistoryCap || kept[0] != "202" || kept[len(kept)-1] != "1201" {
		t.Fatalf("kept %d lines from %q to %q", len(kept), kept[0], kept[len(kept)-1])
	}
}

func TestHistoryFileErrorsAreReported(t *testing.T) {
	folder := filepath.Join(t.TempDir(), historyFile)
	if err := os.Mkdir(folder, appDataDirMode); err != nil {
		t.Fatal(err)
	}
	if err := appendHistoryLine(folder, []byte("{}")); err == nil {
		t.Error("appending to a folder must fail")
	}
	if _, err := readHistoryFile(folder); err == nil {
		t.Error("reading a folder must fail")
	}
	if err := compactHistory(filepath.Join(folder, "missing", historyFile), [][]byte{[]byte("{}")}); err == nil {
		t.Error("compacting into a missing folder must fail")
	}
}
