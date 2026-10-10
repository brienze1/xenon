package persistence

import (
	"os"
	"path/filepath"

	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/fileutils"
)

type commitJournal struct {
	writes map[string][]byte
	moves  [][2]string
}

func (c *commitJournal) move(source, target string) error {
	if err := os.MkdirAll(filepath.Dir(target), workspaceDirMode); err != nil {
		return errs.FileError(err, "IMP-27500")
	}
	if exists(target) {
		return errs.ConflictError("a file with this name already exists", "IMP-28409")
	}
	if err := os.Rename(source, target); err != nil {
		return errs.FileError(err, "IMP-29500")
	}
	c.moves = append(c.moves, [2]string{source, target})
	return nil
}

func (c *commitJournal) undo() {
	for path, original := range c.writes {
		_ = fileutils.WriteFileAtomic(path, original, workspaceFileMode)
	}
	for position := len(c.moves) - 1; position >= 0; position-- {
		_ = os.Rename(c.moves[position][1], c.moves[position][0])
	}
}
