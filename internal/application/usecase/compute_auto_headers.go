package usecase

import "github.com/brienze1/xenon/internal/domain/entity"

const (
	acceptAnything   = "*/*"
	acceptedEncoding = "gzip, deflate, br"
	keepAlive        = "keep-alive"
	userAgentPrefix  = "Xenon/"
)

func ComputeAutoHeaders(request entity.EffectiveRequest, appVersion string) []entity.AutoHeader {
	candidates := make([]entity.AutoHeader, 0)
	if contentType := request.Body.ContentType(); contentType != "" {
		candidates = append(candidates,
			contentTypeHeader(request.Body, contentType),
			entity.NewComputedAutoHeader(entity.HeaderContentLength, entity.ComputedAtSend))
	}
	candidates = append(candidates,
		entity.NewComputedAutoHeader(entity.HeaderHost, hostValue(request.URL)),
		entity.NewAutoHeader(entity.HeaderUserAgent, userAgentPrefix+appVersion),
		entity.NewAutoHeader(entity.HeaderAccept, acceptAnything),
		entity.NewAutoHeader(entity.HeaderAcceptEncoding, acceptedEncoding),
		entity.NewAutoHeader(entity.HeaderConnection, keepAlive))
	if authorization, found := request.Auth.Header(); found {
		candidates = append(candidates, entity.NewAutoHeader(entity.HeaderAuthorization, authorization))
	}
	headers := make([]entity.AutoHeader, 0, len(candidates))
	for _, candidate := range candidates {
		if request.HasHeader(candidate.Key) {
			continue
		}
		candidate.Disabled = request.IsAutoHeaderDisabled(candidate.Key)
		headers = append(headers, candidate)
	}
	return headers
}

func contentTypeHeader(body entity.Body, contentType string) entity.AutoHeader {
	if body.IsMultipart() {
		return entity.NewComputedAutoHeader(entity.HeaderContentType, contentType)
	}
	return entity.NewAutoHeader(entity.HeaderContentType, contentType)
}

func hostValue(url string) string {
	if host := entity.ParseRequestUrl(url).Host(); host != "" {
		return host
	}
	return entity.ComputedAtSend
}
