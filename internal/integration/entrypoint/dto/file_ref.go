package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/filestate"
)

func newFileState(ref entity.FileRef) string {
	if ref.NeedsRepick() {
		return filestate.Repick.String()
	}
	return filestate.Ok.String()
}

func fileRefToEntity(src, state string) entity.FileRef {
	if filestate.GetFileState(state) == filestate.Repick {
		return entity.NewRepickFileRef(src)
	}
	return entity.FileRef{Src: src}
}
