package model

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/persistence/yamlnode"
)

type Settings struct {
	TimeoutMs       *int  `yaml:"timeoutMs" json:"timeout_ms,omitempty"`
	FollowRedirects *bool `yaml:"followRedirects" json:"follow_redirects,omitempty"`
	VerifySsl       *bool `yaml:"verifySsl" json:"verify_ssl,omitempty"`
	EncodeUrl       *bool `yaml:"encodeUrl" json:"encode_url,omitempty"`
}

func NewSettings(settings entity.RequestSettings) Settings {
	return Settings{
		TimeoutMs:       &settings.TimeoutMs,
		FollowRedirects: &settings.FollowRedirects,
		VerifySsl:       &settings.VerifySsl,
		EncodeUrl:       &settings.EncodeUrl,
	}
}

func (s *Settings) ToEntity() entity.RequestSettings {
	settings := entity.DefaultRequestSettings()
	if s.TimeoutMs != nil {
		settings.TimeoutMs = *s.TimeoutMs
	}
	settings.FollowRedirects = isEnabled(s.FollowRedirects)
	settings.VerifySsl = isEnabled(s.VerifySsl)
	settings.EncodeUrl = isEnabled(s.EncodeUrl)
	return settings
}

func (s *Settings) Field() yamlnode.Field {
	settings := s.ToEntity()
	return yamlnode.Object("settings",
		yamlnode.Int("timeoutMs", settings.TimeoutMs),
		yamlnode.Bool("followRedirects", settings.FollowRedirects),
		yamlnode.Bool("verifySsl", settings.VerifySsl),
		yamlnode.Bool("encodeUrl", settings.EncodeUrl),
	)
}
