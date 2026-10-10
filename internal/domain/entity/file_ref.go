package entity

import (
	"path/filepath"
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/filestate"
)

type FileRef struct {
	Src   string
	State enums.FileState
}

func NewRepickFileRef(src string) FileRef {
	return FileRef{Src: src, State: filestate.Repick}
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

func (f FileRef) NeedsRepick() bool {
	return f.State == filestate.Repick
}

func (f FileRef) WithSrc(src string) FileRef {
	f.Src = src
	return f
}

func (f FileRef) Resolve(workspaceRoot string) string {
	if f.IsEmpty() || isAbsolutePath(f.Src) {
		return f.Src
	}
	return filepath.Join(workspaceRoot, filepath.FromSlash(f.Src))
}

func (f FileRef) RelativeTo(workspaceRoot string) FileRef {
	if !isAbsolutePath(f.Src) {
		return f
	}
	return NewFileRef(workspaceRoot, f.Src).withState(f.State)
}

func (f FileRef) withState(state enums.FileState) FileRef {
	f.State = state
	return f
}

// isAbsolutePath also treats a rooted path without a drive letter ("/x" or "\x" on Windows)
// as absolute: it never names a file inside the workspace.
func isAbsolutePath(path string) bool {
	return filepath.IsAbs(path) || strings.HasPrefix(filepath.ToSlash(path), "/")
}
