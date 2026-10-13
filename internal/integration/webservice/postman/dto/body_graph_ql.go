package dto

type BodyGraphQl struct {
	Variables any    `json:"variables"`
	Query     string `json:"query"`
}
