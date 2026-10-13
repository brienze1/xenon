package entity

type OAuth2GrantRequest struct {
	Code        string
	Verifier    string
	RedirectURI string
	Config      OAuth2Config
}

func NewOAuth2GrantRequest(config OAuth2Config) OAuth2GrantRequest {
	return OAuth2GrantRequest{Config: config, RedirectURI: config.RedirectURI}
}

func (o OAuth2GrantRequest) WithCode(code, verifier string) OAuth2GrantRequest {
	o.Code = code
	o.Verifier = verifier
	return o
}

func (o OAuth2GrantRequest) LogValue() any {
	return map[string]any{"config": o.Config.LogValue(), "hasCode": o.Code != "", "hasVerifier": o.Verifier != ""}
}
