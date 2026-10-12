package entity

import (
	"path"
	"strings"
)

const parentSegment = ".."

type SyncSettings struct {
	Repository string
	Branch     string
	Subfolder  string
}

func NewSyncSettings(repository, branch, subfolder string) SyncSettings {
	return SyncSettings{Repository: strings.TrimSpace(repository), Branch: strings.TrimSpace(branch), Subfolder: cleanSubfolder(subfolder)}
}

func (s SyncSettings) IsEmpty() bool {
	return s.Repository == "" && s.Branch == "" && s.Subfolder == ""
}

func (s SyncSettings) Differs(other SyncSettings) bool {
	return s != other
}

func HasSafeSubfolder(subfolder string) bool {
	slashed := strings.ReplaceAll(strings.TrimSpace(subfolder), "\\", "/")
	if strings.HasPrefix(slashed, "/") || strings.Contains(slashed, ":") {
		return false
	}
	for _, segment := range strings.Split(slashed, "/") {
		if segment == parentSegment {
			return false
		}
	}
	return true
}

func cleanSubfolder(subfolder string) string {
	slashed := strings.ReplaceAll(strings.TrimSpace(subfolder), "\\", "/")
	return strings.Trim(path.Clean("/"+slashed), "/")
}
