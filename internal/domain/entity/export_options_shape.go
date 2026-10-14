package entity

import (
	"slices"
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/exportshape"
)

// ShapeFor picks the format's zip shape when the path ends in .zip and its first other shape otherwise.
func (o ExportOptions) ShapeFor(format enums.ExportFormat, path string) enums.ExportShape {
	index := slices.IndexFunc(o.Formats, func(option ExportFormatOption) bool { return option.Format == format })
	if index < 0 || len(o.Formats[index].Shapes) == 0 {
		return exportshape.File
	}
	shapes := o.Formats[index].Shapes
	wantsZip := strings.HasSuffix(strings.ToLower(path), xenonZipSuffix)
	for _, shape := range shapes {
		if (shape == exportshape.Zip) == wantsZip {
			return shape
		}
	}
	return shapes[0]
}
