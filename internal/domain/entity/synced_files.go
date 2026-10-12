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

func (s SyncedFiles) Holds(path, hash string) bool {
	return s[path] == hash
}

func (s SyncedFiles) Differs(head map[string]string, path string) bool {
	return s[path] != head[path]
}

func (s SyncedFiles) LocateChanges(changes []PathChange, head map[string]string) ([]string, map[string]string) {
	present, deleted := make([]string, 0, len(changes)), map[string]string{}
	for _, change := range changes {
		if change.IsDeletion() {
			deleted[change.Path] = head[change.Path]
			continue
		}
		present = append(present, change.Path)
	}
	return present, deleted
}

func (s SyncedFiles) LocatePlans(plans []PathPlan) ([]string, map[string]string) {
	present, blobs := make([]string, 0, len(plans)), map[string]string{}
	for _, plan := range plans {
		if _, found := s[plan.Path]; found {
			present = append(present, plan.Path)
			continue
		}
		blobs[plan.Path] = plan.AnyVersion()
	}
	return present, blobs
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
