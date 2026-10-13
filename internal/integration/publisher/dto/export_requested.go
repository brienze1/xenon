package dto

const ExportRequestedEvent = "export:requested"

type ExportRequested struct {
	Source string `json:"source"`
}

func NewExportRequested(source string) ExportRequested {
	return ExportRequested{Source: source}
}
