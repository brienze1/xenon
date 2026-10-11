package service

import (
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
)

type exportDocuments struct {
	entries []entity.ExportEntry
	names   []string
	summary entity.ExportSummary
}

func (d *exportDocuments) add(folder, name, suffix string, data []byte) {
	base := usecase.DeriveFileName(name, itemkind.Folder, d.names)
	d.names = append(d.names, base)
	d.entries = append(d.entries, entity.NewExportEntry(folder+base+suffix, data))
}
