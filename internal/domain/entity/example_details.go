package entity

type ExampleDetails struct {
	RequestID string
	Revision  string
	Example   SavedExample
}

func NewExampleDetails(request Request, example SavedExample) ExampleDetails {
	return ExampleDetails{RequestID: request.ID, Revision: request.Revision, Example: example}
}

func (e ExampleDetails) ServesBodyByRoute() bool {
	return !e.Example.Response.HasInlineBody()
}

func (e ExampleDetails) LogValue() any {
	return map[string]any{"requestId": e.RequestID, "example": e.Example.LogValue()}
}
