package dto

type MCPTreeExample struct {
	Id       string                 `json:"id"`
	Name     string                 `json:"name"`
	Request  ExampleRequest         `json:"request"`
	Response MCPTreeExampleResponse `json:"response"`
}

func NewMCPTreeExample(example Example) MCPTreeExample {
	return MCPTreeExample{Id: example.Id, Name: example.Name, Request: example.Request, Response: NewMCPTreeExampleResponse(example.Response)}
}
