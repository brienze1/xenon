package entity

type PathVariable struct {
	Key         string
	Value       string
	Description string
}

func (p PathVariable) MapText(transform func(string) string) PathVariable {
	p.Value = transform(p.Value)
	return p
}

func (p PathVariable) HasKey(name string) bool {
	return p.Key == name
}
