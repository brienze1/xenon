package usecase

import "github.com/brienze1/xenon/internal/domain/entity"

func BuildUrlFromParams(request entity.Request) entity.Request {
	request.URL = entity.ParseRequestUrl(request.URL).WithPathVariableNames(request.PathVariableKeys()).WithQuery(request.Params).String()
	return request
}
