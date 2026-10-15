package entity

import (
	"slices"
	"strconv"
	"strings"

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

func NewUnsupportedRequestWarning(name, kind, parentPath string) ImportWarning {
	return NewImportWarning(warningcode.UnsupportedRequestType, name+" is a "+kind+" request, which Xenon does not support; it was skipped", "", parentPath)
}

func NewMultiAuthWarning(owner, path string, authTypes []string) ImportWarning {
	message := "Postman multi-auth on " + owner + " lists " + strconv.Itoa(len(authTypes)) + " auths; Xenon imported the first (" + authTypes[0] + ") and kept " + strings.Join(authTypes[1:], ", ") + " for export only"
	return NewImportWarning(warningcode.UnsupportedAuth, message, "", path)
}

func NewProtocolScriptWarning(scriptType, owner, protocol, path string) ImportWarning {
	return NewImportWarning(warningcode.ProtocolScript, "The "+scriptType+" script on "+owner+" runs only for "+protocol+" requests in Postman; it is kept for export and Xenon does not run it", "", path)
}

func NewLinkedSecretWarning(key, path string) ImportWarning {
	return NewImportWarning(warningcode.LinkedSecret, key+" is linked to a Postman vault secret; enter its value in Xenon", "", path)
}

func NewSkippedEntriesWarning(folder string, entries []ImportSkippedEntry) ImportWarning {
	parts := make([]string, 0, len(entries))
	for _, entry := range slices.SortedFunc(slices.Values(entries), compareSkippedEntries) {
		parts = append(parts, entry.Name+" ("+entry.ShortReason()+")")
	}
	return NewImportWarning(warningcode.SkippedPart, "Skipped in "+folder+": "+strings.Join(parts, ", "), "", "")
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
