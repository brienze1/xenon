package dto

type MCPDocs struct {
	Yaml string `json:"yaml"`
}

func NewMCPDocs(documentation Documentation) MCPDocs {
	return MCPDocs{Yaml: documentation.Yaml}
}
