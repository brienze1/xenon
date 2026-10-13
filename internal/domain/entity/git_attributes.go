package entity

import (
	"path"
	"strings"
)

const (
	gitAttributesFileName = ".gitattributes"
	lfsFilter             = "filter=lfs"
)

type GitAttributes map[string]string

func IsGitAttributesFor(filePath, subfolder string) bool {
	if path.Base(filePath) != gitAttributesFileName {
		return false
	}
	folder := strings.Trim(path.Dir(filePath), ".")
	return folder == "" || subfolder == "" || isSameOrInsideSlashed(subfolder, folder) || isSameOrInsideSlashed(folder, subfolder)
}

func (g GitAttributes) UsesLfs() bool {
	for _, content := range g {
		for _, line := range strings.Split(content, "\n") {
			trimmed := strings.TrimSpace(line)
			if !strings.HasPrefix(trimmed, "#") && strings.Contains(trimmed, lfsFilter) {
				return true
			}
		}
	}
	return false
}

func isSameOrInsideSlashed(child, parent string) bool {
	return child == parent || strings.HasPrefix(child, parent+"/")
}
