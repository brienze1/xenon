package persistence

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/historyfilter"
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
		if err := appendHistoryLine(path, []byte(`{"n":2}`), entity.DefaultHistoryCaps); err != nil {
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
	threshold := entity.DefaultHistoryCap + entity.DefaultHistoryCap/5
	lines := make([]string, 0, threshold)
	for index := 1; index <= threshold; index++ {
		lines = append(lines, strconv.Itoa(index))
	}
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), appDataFileMode); err != nil {
		t.Fatal(err)
	}
	if err := appendHistoryLine(path, []byte("1201"), entity.DefaultHistoryCaps); err != nil {
		t.Fatalf("appendHistoryLine() error = %v", err)
	}
	got, _ := os.ReadFile(path)
	kept := strings.Split(strings.TrimSuffix(string(got), "\n"), "\n")
	if len(kept) != entity.DefaultHistoryCap || kept[0] != "202" || kept[len(kept)-1] != "1201" {
		t.Fatalf("kept %d lines from %q to %q", len(kept), kept[0], kept[len(kept)-1])
	}
}

func TestAppendHistoryLineCapsEachSourceOnItsOwn(t *testing.T) {
	path := filepath.Join(t.TempDir(), historyFile)
	caps := entity.HistoryCaps{User: 1000, AI: 100}
	if err := os.WriteFile(path, []byte(`{"id":"user","source":"user"}`+"\n"), appDataFileMode); err != nil {
		t.Fatal(err)
	}
	for index := 0; index < 121; index++ {
		if err := appendHistoryLine(path, []byte(`{"id":"`+strconv.Itoa(index)+`","source":"ai"}`), caps); err != nil {
			t.Fatal(err)
		}
	}
	got, _ := os.ReadFile(path)
	kept := strings.Split(strings.TrimSuffix(string(got), "\n"), "\n")
	if len(kept) != 101 || kept[0] != `{"id":"user","source":"user"}` || kept[100] != `{"id":"120","source":"ai"}` {
		t.Fatalf("kept %d lines: %q … %q", len(kept), kept[0], kept[len(kept)-1])
	}
}

func TestClearHistoryBySourceKeepsTheOtherSource(t *testing.T) {
	path := filepath.Join(t.TempDir(), historyFile)
	stored := `{"id":"1"}` + "\n" + `{"id":"2","source":"ai"}` + "\n" + `{"id":"3","source":"user"}` + "\n"
	if err := os.WriteFile(path, []byte(stored), appDataFileMode); err != nil {
		t.Fatal(err)
	}
	cases := map[enums.HistoryFilter]int{historyfilter.AI: 2, historyfilter.User: 1, historyfilter.All: 0}
	for source, want := range cases {
		kept, err := historyLinesKeptOnClear(path, source)
		if err != nil || len(kept) != want {
			t.Errorf("clear %s kept %d lines (%v), want %d", source, len(kept), err, want)
		}
	}
	if joined := joinHistoryLines([][]byte{}); len(joined) != 0 {
		t.Errorf("joining no lines = %q", joined)
	}
}

func TestHistoryFileErrorsAreReported(t *testing.T) {
	folder := filepath.Join(t.TempDir(), historyFile)
	if err := os.Mkdir(folder, appDataDirMode); err != nil {
		t.Fatal(err)
	}
	if err := appendHistoryLine(folder, []byte("{}"), entity.DefaultHistoryCaps); err == nil {
		t.Error("appending to a folder must fail")
	}
	if _, err := readHistoryFile(folder); err == nil {
		t.Error("reading a folder must fail")
	}
	if _, err := compactedHistoryLines(folder, entity.DefaultHistoryCaps); err == nil {
		t.Error("compacting a folder must fail")
	}
	if _, err := historyLinesKeptOnClear(folder, historyfilter.AI); err == nil {
		t.Error("reading a folder must fail")
	}
	if err := compactHistory(filepath.Join(folder, "missing", historyFile), [][]byte{[]byte("{}")}); err == nil {
		t.Error("compacting into a missing folder must fail")
	}
}
