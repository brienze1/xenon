package usecase

import "github.com/brienze1/xenon/internal/domain/entity"

func BuildUrlFromParams(request entity.Request) entity.Request {
	names := make([]string, 0, len(request.PathVariables))
	for _, variable := range request.PathVariables {
		names = append(names, variable.Key)
	}
	request.URL = entity.ParseRequestUrl(request.URL).WithPathVariableNames(names).WithQuery(request.Params).String()
	return request
}
