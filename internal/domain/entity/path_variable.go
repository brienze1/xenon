package entity

type PathVariable struct {
	Key         string
	Value       string
	Description string
}

func (p PathVariable) HasKey(name string) bool {
	return p.Key == name
}
