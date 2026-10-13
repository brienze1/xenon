package entity

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums/bodymode"
)

const fileMarker = "@"

// PreviewText is the body as one text for display: raw text, form rows, a GraphQL query and variables, or the file sent.
func (b Body) PreviewText() string {
	switch b.Mode {
	case bodymode.Raw:
		return b.Raw.Content
	case bodymode.UrlEncoded:
		pairs := make([]string, 0, len(b.UrlEncoded))
		for _, row := range b.UrlEncoded {
			if row.IsActive() {
				pairs = append(pairs, row.Key+"="+row.Value)
			}
		}
		return strings.Join(pairs, "&")
	case bodymode.FormData:
		lines := make([]string, 0, len(b.FormData))
		for _, row := range b.FormData {
			if row.Enabled && row.IsFile() {
				lines = append(lines, row.Key+": "+fileMarker+row.Src.Src)
			} else if row.Enabled {
				lines = append(lines, row.Key+": "+row.Value)
			}
		}
		return strings.Join(lines, "\n")
	case bodymode.GraphQl:
		return strings.TrimSpace(b.GraphQl.Query + "\n\n" + b.GraphQl.Variables)
	case bodymode.Binary:
		return fileMarker + b.Binary.Src
	default:
		return ""
	}
}
