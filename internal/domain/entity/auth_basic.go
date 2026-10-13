package entity

import (
	"encoding/base64"
	"strings"
)

const basicScheme = "Basic "

type AuthBasic struct {
	Username string
	Password string
}

func (a AuthBasic) Header() string {
	return basicScheme + a.Credentials()
}

func (a AuthBasic) MapText(transform func(string) string) AuthBasic {
	a.Username = transform(a.Username)
	a.Password = transform(a.Password)
	return a
}

func (a AuthBasic) IsEmpty() bool {
	return a == AuthBasic{}
}

func (a AuthBasic) LogValue() any {
	return map[string]any{"username": a.Username, "hasPassword": a.Password != ""}
}

func (a AuthBasic) Credentials() string {
	return base64.StdEncoding.EncodeToString([]byte(a.Username + ":" + a.Password))
}

func (a AuthBasic) Uses(secrets []string) bool {
	for _, secret := range secrets {
		if secret != "" && (strings.Contains(a.Username, secret) || strings.Contains(a.Password, secret)) {
			return true
		}
	}
	return false
}
