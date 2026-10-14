package entity

import (
	"testing"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/exportformat"
	"github.com/brienze1/xenon/internal/domain/enums/exportitemkind"
	"github.com/brienze1/xenon/internal/domain/enums/exportshape"
)

func TestExportOptionsShapeForFollowsThePathExtension(t *testing.T) {
	collection := ExportOptions{Formats: ExportFormatOptions(exportitemkind.Collection, "api")}
	workspace := ExportOptions{Formats: ExportFormatOptions(exportitemkind.Workspace, "ws")}
	cases := []struct {
		name    string
		options ExportOptions
		format  enums.ExportFormat
		path    string
		want    enums.ExportShape
	}{
		{"zip path picks zip", collection, exportformat.XenonYaml, "/tmp/api.ZIP", exportshape.Zip},
		{"yaml path picks bundle", collection, exportformat.XenonYaml, "/tmp/api.yaml", exportshape.Bundle},
		{"single shape is kept", workspace, exportformat.Postman, "/tmp/ws.json", exportshape.Zip},
		{"unknown format falls back to file", collection, exportformat.Unknown, "/tmp/api.zip", exportshape.File},
		{"format without shapes falls back to file", ExportOptions{Formats: []ExportFormatOption{{Format: exportformat.Postman}}}, exportformat.Postman, "/tmp/a", exportshape.File},
	}
	for _, c := range cases {
		if got := c.options.ShapeFor(c.format, c.path); got != c.want {
			t.Errorf("%s: ShapeFor() = %q, want %q", c.name, got, c.want)
		}
	}
}
