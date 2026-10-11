package usecase

import "github.com/brienze1/xenon/internal/domain/entity"

func ApplyApiKey(request entity.EffectiveRequest, apiKey entity.ApiKeyAuth) entity.EffectiveRequest {
	switch {
	case apiKey.SendsNothing():
		return request
	case apiKey.IsInQuery():
		return request.WithQueryParam(apiKey.Key, apiKey.Value)
	default:
		return request.WithHeader(apiKey.Key, apiKey.Value)
	}
}
