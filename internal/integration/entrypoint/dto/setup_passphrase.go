package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type SetupPassphrase struct {
	Passphrase   string `json:"passphrase" validate:"required"`
	Confirmation string `json:"confirmation"`
}

func (s *SetupPassphrase) ToEntity() entity.PassphraseSetup {
	return entity.PassphraseSetup{Passphrase: s.Passphrase, Confirmation: s.Confirmation}
}

func (s SetupPassphrase) LogValue() any {
	return map[string]any{"passphrase": entity.MaskedValue}
}
