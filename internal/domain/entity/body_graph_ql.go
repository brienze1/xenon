package entity

type BodyGraphQl struct {
	Query     string
	Variables string
}

func (b BodyGraphQl) IsEmpty() bool {
	return b == BodyGraphQl{}
}
