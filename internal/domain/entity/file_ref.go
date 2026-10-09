package entity

import (
	"path/filepath"
	"strings"
)

type FileRef struct {
	Src string
}

func NewFileRef(workspaceRoot, absolutePath string) FileRef {
	relative, err := filepath.Rel(workspaceRoot, absolutePath)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return FileRef{Src: absolutePath}
	}
	return FileRef{Src: filepath.ToSlash(relative)}
}

func (f FileRef) IsEmpty() bool {
	return strings.TrimSpace(f.Src) == ""
}

func (f FileRef) Resolve(workspaceRoot string) string {
	if f.IsEmpty() || filepath.IsAbs(f.Src) {
		return f.Src
	}
	return filepath.Join(workspaceRoot, filepath.FromSlash(f.Src))
}

func (f FileRef) RelativeTo(workspaceRoot string) FileRef {
	if !filepath.IsAbs(f.Src) {
		return f
	}
	return NewFileRef(workspaceRoot, f.Src)
}
