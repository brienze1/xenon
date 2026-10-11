package entity

import (
	"slices"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/exportformat"
	"github.com/brienze1/xenon/internal/domain/enums/exportitemkind"
	"github.com/brienze1/xenon/internal/domain/enums/exportshape"
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

func (e ExportRequest) IsArchive() bool {
	return e.Shape == exportshape.Zip
}

func (e ExportRequest) ExportedScopes(scopes []ExportScope) []ExportScope {
	if e.IsPostman() || e.ItemKind == exportitemkind.Workspace || e.ItemKind == exportitemkind.Collection {
		return scopes
	}
	exported := make([]ExportScope, 0, len(scopes))
	for _, scope := range scopes {
		if !scope.Ref.IsCollection() {
			exported = append(exported, scope)
		}
	}
	return exported
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
