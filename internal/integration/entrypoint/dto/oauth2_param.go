package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type OAuth2Param struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	SendAs  string `json:"send_as" validate:"omitempty,oneof=request_url request_body request_header"`
	Enabled bool   `json:"enabled"`
}

func NewOAuth2Params(params []entity.OAuth2Param) []OAuth2Param {
	result := make([]OAuth2Param, 0, len(params))
	for _, param := range params {
		result = append(result, OAuth2Param{Key: param.Key, Value: param.Value, SendAs: param.SendAs, Enabled: param.Enabled})
	}
	return result
}

func oauth2ParamsToEntity(params []OAuth2Param, defaultSendAs string) []entity.OAuth2Param {
	result := make([]entity.OAuth2Param, 0, len(params))
	for _, param := range params {
		sendAs := param.SendAs
		if sendAs == "" {
			sendAs = defaultSendAs
		}
		result = append(result, entity.OAuth2Param{Key: param.Key, Value: param.Value, SendAs: sendAs, Enabled: param.Enabled})
	}
	return result
}
