package entity

const bearerScheme = "Bearer "

type AuthBearer struct {
	Token string
}

func (a AuthBearer) Header() string {
	return bearerScheme + a.Token
}

func (a AuthBearer) IsEmpty() bool {
	return a == AuthBearer{}
}

func (a AuthBearer) LogValue() any {
	return map[string]any{"hasToken": a.Token != ""}
}
