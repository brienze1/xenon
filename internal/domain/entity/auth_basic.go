package entity

import "encoding/base64"

const basicScheme = "Basic "

type AuthBasic struct {
	Username string
	Password string
}

func (a AuthBasic) Header() string {
	return basicScheme + base64.StdEncoding.EncodeToString([]byte(a.Username+":"+a.Password))
}

func (a AuthBasic) IsEmpty() bool {
	return a == AuthBasic{}
}

func (a AuthBasic) LogValue() any {
	return map[string]any{"username": a.Username, "hasPassword": a.Password != ""}
}
