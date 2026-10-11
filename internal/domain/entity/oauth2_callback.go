package entity

type OAuth2Callback struct {
	Code             string
	State            string
	Error            string
	ErrorDescription string
	AccessToken      string
	TokenType        string
	ExpiresIn        string
	Scope            string
	IDToken          string
}

func (o OAuth2Callback) IsError() bool {
	return o.Error != ""
}

func (o OAuth2Callback) MatchesState(state string) bool {
	return o.State == state
}

func (o OAuth2Callback) LogValue() any {
	return map[string]any{"hasCode": o.Code != "", "hasAccessToken": o.AccessToken != "", "error": o.Error}
}
