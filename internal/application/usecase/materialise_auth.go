package usecase

import "github.com/brienze1/xenon/internal/domain/entity"

func MaterialiseAuth(request entity.EffectiveRequest) entity.EffectiveRequest {
	header, found := request.Auth.Header()
	if !found || request.HasHeader(entity.HeaderAuthorization) || request.IsAutoHeaderDisabled(entity.HeaderAuthorization) {
		return request
	}
	return request.WithHeader(entity.HeaderAuthorization, header)
}
