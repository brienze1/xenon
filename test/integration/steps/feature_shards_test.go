package steps

import (
	"fmt"
	"testing"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	featuresRoot = "../features"
	shardEnv     = "XENON_BDD_SHARD"
)

type featureFile struct {
	path string
	size int64
}

func featurePaths() ([]string, error) {
	shard, ok := os.LookupEnv(shardEnv)
	if !ok || shard == "" {
		return []string{featuresRoot}, nil
	}
	index, count, err := parseShard(shard)
	if err != nil {
		return nil, err
	}
	files, err := featureFiles()
	if err != nil {
		return nil, err
	}
	return shardOf(files, index, count), nil
}

func parseShard(shard string) (int, int, error) {
	parts := strings.Split(shard, "/")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("%s must look like 1/4, got %q", shardEnv, shard)
	}
	index, indexErr := strconv.Atoi(parts[0])
	count, countErr := strconv.Atoi(parts[1])
	if indexErr != nil || countErr != nil || count < 1 || index < 1 || index > count {
		return 0, 0, fmt.Errorf("%s must look like 1/4, got %q", shardEnv, shard)
	}
	return index, count, nil
}

func featureFiles() ([]featureFile, error) {
	files := make([]featureFile, 0)
	err := filepath.WalkDir(featuresRoot, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || filepath.Ext(path) != ".feature" {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		files = append(files, featureFile{path: path, size: info.Size()})
		return nil
	})
	return files, err
}

func shardOf(files []featureFile, index, count int) []string {
	sort.Slice(files, func(i, j int) bool {
		if files[i].size != files[j].size {
			return files[i].size > files[j].size
		}
		return files[i].path < files[j].path
	})
	loads := make([]int64, count)
	paths := make([]string, 0)
	for _, file := range files {
		lightest := 0
		for shard := range loads {
			if loads[shard] < loads[lightest] {
				lightest = shard
			}
		}
		loads[lightest] += file.size
		if lightest == index-1 {
			paths = append(paths, file.path)
		}
	}
	return paths
}

func TestFeatureShardsCoverEveryFileOnce(t *testing.T) {
	files, err := featureFiles()
	if err != nil {
		t.Fatal(err)
	}
	for _, count := range []int{1, 2, 3, 4} {
		seen := map[string]int{}
		for index := 1; index <= count; index++ {
			for _, path := range shardOf(append([]featureFile{}, files...), index, count) {
				seen[path]++
			}
		}
		if len(seen) != len(files) {
			t.Fatalf("%d shards cover %d of %d feature files", count, len(seen), len(files))
		}
		for path, times := range seen {
			if times != 1 {
				t.Fatalf("%d shards run %s %d times", count, path, times)
			}
		}
	}
}
