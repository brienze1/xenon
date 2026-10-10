package dto

const ImportRequestedEvent = "import:requested"

type ImportRequested struct {
	Source string `json:"source"`
}

func NewImportRequested(source string) ImportRequested {
	return ImportRequested{Source: source}
}
