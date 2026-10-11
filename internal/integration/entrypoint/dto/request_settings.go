package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type RequestSettings struct {
	FollowRedirects  *bool `json:"follow_redirects,omitempty"`
	VerifySsl        *bool `json:"verify_ssl,omitempty"`
	EncodeUrl        *bool `json:"encode_url,omitempty"`
	DisableCookieJar bool  `json:"disable_cookie_jar"`
	TimeoutMs        int   `json:"timeout_ms" validate:"gte=0"`
}

func NewRequestSettings(settings entity.RequestSettings) RequestSettings {
	return RequestSettings{
		TimeoutMs:        settings.TimeoutMs,
		FollowRedirects:  &settings.FollowRedirects,
		VerifySsl:        &settings.VerifySsl,
		EncodeUrl:        &settings.EncodeUrl,
		DisableCookieJar: settings.DisableCookieJar,
	}
}

func (r *RequestSettings) ToEntity() entity.RequestSettings {
	settings := entity.DefaultRequestSettings()
	settings.TimeoutMs = r.TimeoutMs
	if r.FollowRedirects != nil {
		settings.FollowRedirects = *r.FollowRedirects
	}
	if r.VerifySsl != nil {
		settings.VerifySsl = *r.VerifySsl
	}
	if r.EncodeUrl != nil {
		settings.EncodeUrl = *r.EncodeUrl
	}
	settings.DisableCookieJar = r.DisableCookieJar
	return settings
}
