package entity

type AIActivityTarget struct {
	ID   string
	Name string
	Path string
}

func NewAIItemTarget(id string) AIActivityTarget {
	return AIActivityTarget{ID: id}
}

// NewAIRequestTarget names an ad-hoc request by its method and URL.
func NewAIRequestTarget(method, rawURL string) AIActivityTarget {
	return AIActivityTarget{Name: method + " " + rawURL}
}

func (a AIActivityTarget) IsEmpty() bool {
	return a == AIActivityTarget{}
}
