package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type UnlockVault struct {
	Passphrase string `json:"passphrase" validate:"required"`
}

func (u UnlockVault) LogValue() any {
	return map[string]any{"passphrase": entity.MaskedValue}
}
