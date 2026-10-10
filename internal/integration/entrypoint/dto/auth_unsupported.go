package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type AuthUnsupported struct {
	Source string `json:"source"`
	Type   string `json:"type"`
}

func NewAuthUnsupported(unsupported entity.AuthUnsupported) *AuthUnsupported {
	if unsupported.IsEmpty() {
		return nil
	}
	return &AuthUnsupported{Source: unsupported.Origin.String(), Type: unsupported.Type}
}
