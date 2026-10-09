package usecase

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/authtype"
)

func ResolveEffectiveRequest(request entity.Request, ancestors []entity.Collection, workspaceRoot string) entity.EffectiveRequest {
	return entity.EffectiveRequest{
		Method:              request.NormalizedMethod(),
		URL:                 sendUrl(request),
		Headers:             effectiveHeaders(request, ancestors),
		Auth:                effectiveAuth(request, ancestors),
		Body:                request.Body.ResolveFiles(workspaceRoot),
		Settings:            request.Settings,
		DisabledAutoHeaders: append([]string{}, request.DisabledAutoHeaders...),
	}
}

func sendUrl(request entity.Request) string {
	url := entity.ParseRequestUrl(request.URL).WithPathValues(request.PathVariables).WithDefaultScheme().WithoutFragment()
	if request.ShouldEncodeUrl() {
		url = url.Encoded()
	}
	return url.String()
}

func effectiveAuth(request entity.Request, ancestors []entity.Collection) entity.Auth {
	if !request.Auth.IsInherited() {
		return request.Auth
	}
	for index := len(ancestors) - 1; index >= 0; index-- {
		if !ancestors[index].Auth.IsInherited() {
			return ancestors[index].Auth
		}
	}
	return entity.Auth{Type: authtype.None}
}

func effectiveHeaders(request entity.Request, ancestors []entity.Collection) []entity.KeyValue {
	headers := make([]entity.KeyValue, 0)
	for _, ancestor := range ancestors {
		headers = overrideHeaders(headers, ancestor.Headers)
	}
	return overrideHeaders(headers, request.Headers)
}

func overrideHeaders(headers, overrides []entity.KeyValue) []entity.KeyValue {
	for _, override := range overrides {
		if !override.IsActive() {
			continue
		}
		headers = replaceOrAppend(headers, override)
	}
	return headers
}

func replaceOrAppend(headers []entity.KeyValue, header entity.KeyValue) []entity.KeyValue {
	for index := range headers {
		if headers[index].HasKey(header.Key) {
			headers[index] = header
			return headers
		}
	}
	return append(headers, header)
}
