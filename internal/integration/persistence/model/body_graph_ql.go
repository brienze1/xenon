package model

type BodyGraphQl struct {
	Query     string `yaml:"query" json:"query"`
	Variables string `yaml:"variables" json:"variables"`
}
