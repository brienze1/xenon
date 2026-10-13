package usecase

import (
	"strconv"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/pmapi"
	"github.com/brienze1/xenon/internal/domain/enums/warningcode"
)

func ListPmUsageWarnings(usages []entity.PmUsage) []entity.ImportWarning {
	warnings := make([]entity.ImportWarning, 0)
	seen := map[string]bool{}
	for _, usage := range usages {
		for _, occurrence := range usage.Occurrences {
			key := usage.API + "\x00" + occurrence.ItemID + "\x00" + occurrence.Phase.String()
			if seen[key] {
				continue
			}
			seen[key] = true
			location := occurrence.Phase.String() + ":" + strconv.Itoa(occurrence.Line)
			warnings = append(warnings, pmUsageWarning(usage.API, occurrence).AtLine(location))
		}
	}
	return warnings
}

func pmUsageWarning(api string, occurrence entity.PmOccurrence) entity.ImportWarning {
	if api == pmapi.Require.String() {
		return entity.NewImportWarning(warningcode.PmRequire, "The script uses pm.require, which only works inside Postman", occurrence.ItemID, occurrence.Path)
	}
	return entity.NewImportWarning(warningcode.PmApiOutsideMvp, "The script uses "+api+", which is outside Xenon's script subset", occurrence.ItemID, occurrence.Path)
}
