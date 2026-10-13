package entity

const (
	PkceS256  = "S256"
	PkcePlain = "plain"
)

type Pkce struct {
	Verifier  string
	Challenge string
	Method    string
}
