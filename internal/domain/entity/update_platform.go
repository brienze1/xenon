package entity

import "strings"

const (
	zipArchive   = "zip"
	tarGzArchive = "tar.gz"
)

type UpdatePlatform string

func (u UpdatePlatform) ArchiveExtension() string {
	if strings.HasPrefix(string(u), "windows-") {
		return zipArchive
	}
	return tarGzArchive
}

func (u UpdatePlatform) IsBundle() bool {
	return strings.HasPrefix(string(u), "darwin-")
}
