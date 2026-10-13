package entity

type ExportEntry struct {
	Path string
	Data []byte
}

func NewExportEntry(path string, data []byte) ExportEntry {
	return ExportEntry{Path: path, Data: data}
}
