package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type RequestSettings struct {
	FollowRedirects *bool `json:"follow_redirects,omitempty"`
	VerifySsl       *bool `json:"verify_ssl,omitempty"`
	EncodeUrl       *bool `json:"encode_url,omitempty"`
	TimeoutMs       int   `json:"timeout_ms" validate:"gte=0"`
}

func NewRequestSettings(settings entity.RequestSettings) RequestSettings {
	return RequestSettings{
		TimeoutMs:       settings.TimeoutMs,
		FollowRedirects: &settings.FollowRedirects,
		VerifySsl:       &settings.VerifySsl,
		EncodeUrl:       &settings.EncodeUrl,
	}
}

func (r *RequestSettings) ToEntity() entity.RequestSettings {
	return entity.RequestSettings{
		TimeoutMs:       r.TimeoutMs,
		FollowRedirects: isEnabled(r.FollowRedirects),
		VerifySsl:       isEnabled(r.VerifySsl),
		EncodeUrl:       isEnabled(r.EncodeUrl),
	}
}
