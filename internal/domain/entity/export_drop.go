package entity

type ExportDrop struct {
	ItemPath string
	Field    string
	Reason   string
}

func NewExportDrop(itemPath, field, reason string) ExportDrop {
	return ExportDrop{ItemPath: itemPath, Field: field, Reason: reason}
}
