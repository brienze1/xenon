package importstatus

import "github.com/brienze1/xenon/internal/domain/enums"

const (
	Imported enums.ImportStatus = "imported"
	Replaced enums.ImportStatus = "replaced"
	Skipped  enums.ImportStatus = "skipped"
	Failed   enums.ImportStatus = "failed"
)
