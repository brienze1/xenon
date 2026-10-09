package usecase

import "github.com/brienze1/xenon/internal/domain/entity"

func DeriveParamsFromUrl(request entity.Request) entity.Request {
	url := entity.ParseRequestUrl(request.URL)
	request.Params = mergeQueryParams(request.Params, url.QueryParams())
	request.PathVariables = mergePathVariables(request.PathVariables, url.PathVariableNames())
	return request
}

func mergeQueryParams(existing, parsed []entity.KeyValue) []entity.KeyValue {
	merged := make([]entity.KeyValue, 0, len(existing)+len(parsed))
	next := 0
	for _, row := range existing {
		if !row.IsEnabled() {
			merged = append(merged, row)
			continue
		}
		if next < len(parsed) {
			param := parsed[next]
			param.Description = row.Description
			merged = append(merged, param)
			next++
		}
	}
	return append(merged, parsed[next:]...)
}

func mergePathVariables(existing []entity.PathVariable, names []string) []entity.PathVariable {
	merged := make([]entity.PathVariable, 0, len(names))
	for _, name := range names {
		merged = append(merged, findPathVariable(existing, name))
	}
	return merged
}

func findPathVariable(variables []entity.PathVariable, name string) entity.PathVariable {
	for _, variable := range variables {
		if variable.HasKey(name) {
			return variable
		}
	}
	return entity.PathVariable{Key: name}
}
