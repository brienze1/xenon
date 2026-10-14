package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/warningcode"
	"github.com/brienze1/xenon/internal/domain/enums/warningseverity"
)

var infoWarningCodes = []enums.WarningCode{warningcode.MergedIntoFolder, warningcode.WrappedInFolder, warningcode.VariableKept}

type ImportWarning struct {
	Code     enums.WarningCode
	Severity enums.WarningSeverity
	Message  string
	ItemKey  string
	ItemID   string
	LocalID  string
	Path     string
	Location string
}

func NewImportWarning(code enums.WarningCode, message, itemId, path string) ImportWarning {
	warning := ImportWarning{Code: code, Severity: warningseverity.Warning, Message: message, ItemID: itemId, Path: path}
	for _, info := range infoWarningCodes {
		if info == code {
			warning.Severity = warningseverity.Info
		}
	}
	return warning
}

func (i ImportWarning) Relocated(move ImportRelocation) ImportWarning {
	i.ItemID, i.Path = move.Apply(i.ItemID, i.Path)
	return i
}

func (i ImportWarning) AtLine(location string) ImportWarning {
	i.Location = location
	return i
}

func (i ImportWarning) OnItem(id string) ImportWarning {
	i.ItemID = id
	return i
}

func (i ImportWarning) ForItem(key string) ImportWarning {
	i.ItemKey = key
	return i
}

func (i ImportWarning) InWorkspace(localId string) ImportWarning {
	i.LocalID = localId
	return i
}
