package model

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/yamlnode"
)

type Settings struct {
	TimeoutMs        *int  `yaml:"timeoutMs" json:"timeout_ms,omitempty"`
	FollowRedirects  *bool `yaml:"followRedirects" json:"follow_redirects,omitempty"`
	VerifySsl        *bool `yaml:"verifySsl" json:"verify_ssl,omitempty"`
	EncodeUrl        *bool `yaml:"encodeUrl" json:"encode_url,omitempty"`
	DisableCookieJar *bool `yaml:"disableCookieJar" json:"disable_cookie_jar,omitempty"`
}

func NewSettings(settings entity.RequestSettings) Settings {
	return Settings{
		TimeoutMs:        &settings.TimeoutMs,
		FollowRedirects:  &settings.FollowRedirects,
		VerifySsl:        &settings.VerifySsl,
		EncodeUrl:        &settings.EncodeUrl,
		DisableCookieJar: &settings.DisableCookieJar,
	}
}

func (s *Settings) ToEntity() entity.RequestSettings {
	settings := entity.DefaultRequestSettings()
	if s.TimeoutMs != nil {
		settings.TimeoutMs = *s.TimeoutMs
	}
	if s.FollowRedirects != nil {
		settings.FollowRedirects = *s.FollowRedirects
	}
	if s.VerifySsl != nil {
		settings.VerifySsl = *s.VerifySsl
	}
	if s.EncodeUrl != nil {
		settings.EncodeUrl = *s.EncodeUrl
	}
	if s.DisableCookieJar != nil {
		settings.DisableCookieJar = *s.DisableCookieJar
	}
	return settings
}

func (s *Settings) Field() yamlnode.Field {
	settings := s.ToEntity()
	return yamlnode.Object("settings",
		yamlnode.Int("timeoutMs", settings.TimeoutMs),
		yamlnode.Bool("followRedirects", settings.FollowRedirects),
		yamlnode.Bool("verifySsl", settings.VerifySsl),
		yamlnode.Bool("encodeUrl", settings.EncodeUrl),
		yamlnode.Bool("disableCookieJar", settings.DisableCookieJar),
	)
}
