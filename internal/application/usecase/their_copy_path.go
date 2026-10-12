package usecase

import (
	"path"

	"github.com/brienze1/xenon/internal/domain/entity"
)

const repositoryRootDir = "."

// TheirCopyPath names the "<name> (theirs)" file next to the conflict, unique among the given file sets.
func TheirCopyPath(conflict entity.ConflictItem, files ...map[string]string) string {
	dir := path.Dir(conflict.Path)
	siblings := make([]string, 0)
	for _, set := range files {
		for filePath := range set {
			if path.Dir(filePath) == dir {
				siblings = append(siblings, path.Base(filePath))
			}
		}
	}
	name := DeriveFileName(conflict.CopyName(), conflict.ItemKind, siblings)
	if dir == repositoryRootDir {
		return name
	}
	return dir + "/" + name
}

// OrderParents lists the collection and folder files that may list the file at filePath in their order.
func OrderParents(filePath string) []string {
	dir := path.Dir(filePath)
	if dir == repositoryRootDir {
		return []string{}
	}
	return []string{dir + "/collection.yaml", dir + "/folder.yaml"}
}
