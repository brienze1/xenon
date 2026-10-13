package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/exportformat"
	"github.com/brienze1/xenon/internal/domain/enums/exportitemkind"
	"github.com/brienze1/xenon/internal/domain/enums/exportshape"
)

type Export struct {
	LocalId                string `json:"local_id" validate:"required,not_blank"`
	ItemKind               string `json:"item_kind" validate:"required,oneof=workspace collection folder request environment globals"`
	ItemId                 string `json:"item_id" validate:"required_unless=ItemKind workspace ItemKind globals"`
	Format                 string `json:"format" validate:"required,oneof=xenon_yaml postman"`
	Shape                  string `json:"shape" validate:"required,oneof=zip bundle file"`
	Passphrase             string `json:"passphrase"`
	PassphraseConfirmation string `json:"passphrase_confirmation"`
	TargetPath             string `json:"target_path" validate:"required,absolute_path"`
	IncludeSecrets         bool   `json:"include_secrets"`
	CurrentValuesAsInitial bool   `json:"current_values_as_initial"`
}

func (e *Export) ToEntity() entity.ExportRequest {
	return entity.ExportRequest{
		LocalID:                e.LocalId,
		ItemKind:               exportitemkind.GetExportItemKind(e.ItemKind),
		ItemID:                 e.ItemId,
		Format:                 exportformat.GetExportFormat(e.Format),
		Shape:                  exportshape.GetExportShape(e.Shape),
		Passphrase:             e.Passphrase,
		PassphraseConfirmation: e.PassphraseConfirmation,
		TargetPath:             e.TargetPath,
		IncludeSecrets:         e.IncludeSecrets,
		CurrentValuesAsInitial: e.CurrentValuesAsInitial,
	}
}

func (e Export) LogValue() any {
	return map[string]any{
		"localId":                e.LocalId,
		"itemKind":               e.ItemKind,
		"itemId":                 e.ItemId,
		"format":                 e.Format,
		"shape":                  e.Shape,
		"targetPath":             e.TargetPath,
		"includeSecrets":         e.IncludeSecrets,
		"currentValuesAsInitial": e.CurrentValuesAsInitial,
	}
}
