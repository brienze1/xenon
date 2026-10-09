package model

type AuthBasic struct {
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
}
