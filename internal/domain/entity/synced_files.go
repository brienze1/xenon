package entity

import (
	"sort"
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums/changetype"
)

const (
	localOnlyFolder     = ".xenon"
	importStagingPrefix = ".xenon-import-"
)

type SyncedFiles map[string]string

func NewSyncedFiles(files map[string]string) SyncedFiles {
	synced := SyncedFiles{}
	for path, hash := range files {
		if !isLocalOnlyPath(path) {
			synced[path] = hash
		}
	}
	return synced
}

func (s SyncedFiles) ChangesFrom(head map[string]string) []PathChange {
	changes := make([]PathChange, 0)
	for path, hash := range s {
		switch current, tracked := head[path]; {
		case !tracked:
			changes = append(changes, PathChange{Path: path, Change: changetype.Added})
		case current != hash:
			changes = append(changes, PathChange{Path: path, Change: changetype.Modified})
		}
	}
	for path := range head {
		if _, present := s[path]; !present {
			changes = append(changes, PathChange{Path: path, Change: changetype.Deleted})
		}
	}
	sort.Slice(changes, func(i, j int) bool { return changes[i].Path < changes[j].Path })
	return changes
}

func (s SyncedFiles) Differs(head map[string]string, path string) bool {
	return s[path] != head[path]
}

func isLocalOnlyPath(path string) bool {
	segments := strings.Split(path, "/")
	for _, segment := range segments {
		if segment == localOnlyFolder || strings.HasPrefix(segment, importStagingPrefix) {
			return true
		}
	}
	return strings.HasSuffix(path, temporaryFileSuffix)
}
