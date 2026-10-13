package exportsecrets

import "github.com/brienze1/xenon/internal/domain/enums"

const (
	Included  enums.ExportSecrets = "included"
	Omitted   enums.ExportSecrets = "omitted"
	NamesOnly enums.ExportSecrets = "names_only"
)
