package dto

func (e ExampleResponse) InlineText() (string, bool) {
	if e.Body == nil {
		return "", false
	}
	return *e.Body, true
}

func (e ExampleResponse) StreamsText() bool {
	return e.Body == nil && !e.IsBinary
}
