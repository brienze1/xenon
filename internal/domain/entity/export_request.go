package entity

import (
	"slices"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/exportformat"
	"github.com/brienze1/xenon/internal/domain/enums/exportitemkind"
)

const minExportPassphraseLength = 8

type ExportRequest struct {
	LocalID                string
	ItemKind               enums.ExportItemKind
	ItemID                 string
	Format                 enums.ExportFormat
	Shape                  enums.ExportShape
	Passphrase             string
	PassphraseConfirmation string
	TargetPath             string
	IncludeSecrets         bool
	CurrentValuesAsInitial bool
}

func (e ExportRequest) HasValidCombination() bool {
	option, found := ExportFormatOptionOf(e.ItemKind, e.Format)
	if !found || !slices.Contains(option.Shapes, e.Shape) {
		return false
	}
	return option.IncludeSecretsAllowed || !e.IncludeSecrets
}

func (e ExportRequest) HasValidPassphrase() bool {
	if !e.IncludeSecrets {
		return true
	}
	return len([]rune(e.Passphrase)) >= minExportPassphraseLength && e.Passphrase == e.PassphraseConfirmation
}

func (e ExportRequest) IsPostman() bool {
	return e.Format == exportformat.Postman
}

func (e ExportRequest) IsWorkspace() bool {
	return e.ItemKind == exportitemkind.Workspace
}

func (e ExportRequest) LogValue() any {
	return map[string]any{
		"localId":                e.LocalID,
		"itemKind":               e.ItemKind,
		"itemId":                 e.ItemID,
		"format":                 e.Format,
		"shape":                  e.Shape,
		"includeSecrets":         e.IncludeSecrets,
		"currentValuesAsInitial": e.CurrentValuesAsInitial,
		"targetPath":             e.TargetPath,
	}
}
