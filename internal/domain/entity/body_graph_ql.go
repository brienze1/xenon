package entity

type BodyGraphQl struct {
	Query     string
	Variables string
}

func (b BodyGraphQl) MapText(transform func(string) string) BodyGraphQl {
	b.Query = transform(b.Query)
	b.Variables = transform(b.Variables)
	return b
}

func (b BodyGraphQl) IsEmpty() bool {
	return b == BodyGraphQl{}
}
