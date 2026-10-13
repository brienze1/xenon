package model

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/filestate"
	"github.com/brienze1/xenon/pkg/yamlnode"
)

func fileRefState(ref entity.FileRef) string {
	if ref.NeedsRepick() {
		return filestate.Repick.String()
	}
	return filestate.Ok.String()
}

func fileRefStateField(ref entity.FileRef) yamlnode.Field {
	if ref.NeedsRepick() {
		return yamlnode.String("srcState", filestate.Repick.String())
	}
	return yamlnode.Absent("srcState")
}

func fileRefToEntity(src, state string) entity.FileRef {
	if filestate.GetFileState(state) == filestate.Repick {
		return entity.NewRepickFileRef(src)
	}
	return entity.FileRef{Src: src}
}
