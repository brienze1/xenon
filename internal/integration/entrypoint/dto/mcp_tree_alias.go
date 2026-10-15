package dto

type MCPTreeAlias struct {
	Scripts *Scripts   `json:"scripts,omitempty"`
	Auth    Auth       `json:"auth"`
	Body    Body       `json:"body"`
	Id      string     `json:"id"`
	Name    string     `json:"name"`
	Method  string     `json:"method"`
	Url     string     `json:"url"`
	Params  []KeyValue `json:"params"`
	Headers []KeyValue `json:"headers"`
}

func NewMCPTreeAlias(alias RequestAlias) MCPTreeAlias {
	version := alias.Request
	return MCPTreeAlias{
		Id:      alias.Id,
		Name:    alias.Name,
		Method:  version.Method,
		Url:     version.Url,
		Params:  version.Params,
		Headers: version.Headers,
		Body:    version.Body,
		Auth:    version.Auth,
		Scripts: version.Scripts,
	}
}

func NewMCPTreeAliases(aliases []RequestAlias) []MCPTreeAlias {
	result := make([]MCPTreeAlias, 0, len(aliases))
	for _, alias := range aliases {
		result = append(result, NewMCPTreeAlias(alias))
	}
	return result
}
