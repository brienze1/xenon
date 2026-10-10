package entity

import (
	"slices"
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/importitemkind"
	"github.com/brienze1/xenon/internal/domain/enums/importstatus"
	"github.com/brienze1/xenon/internal/domain/enums/warningcode"
)

type ImportItem struct {
	Variables *ImportVariables
	Root      ImportNode
	Key       string
	Kind      enums.ImportItemKind
	Name      string
	Warnings  []ImportWarning
	PmUsages  []PmUsage
}

func NewImportItem(key string, kind enums.ImportItemKind, root ImportNode, warnings []ImportWarning) ImportItem {
	item := ImportItem{Key: key, Kind: kind, Name: root.Name(), Root: root, Warnings: []ImportWarning{}, PmUsages: []PmUsage{}}
	return item.WithWarnings(warnings)
}

func NewVariablesImportItem(key string, kind enums.ImportItemKind, variables ImportVariables) ImportItem {
	return ImportItem{Key: key, Kind: kind, Name: variables.Environment.Name, Variables: &variables, Warnings: []ImportWarning{}, PmUsages: []PmUsage{}}
}

func (i ImportItem) Named(name string) ImportItem {
	if strings.TrimSpace(name) == "" {
		return i
	}
	if i.IsVariables() {
		variables := *i.Variables
		variables.Environment.Name = strings.TrimSpace(name)
		i.Variables = &variables
	} else {
		i.Root = i.Root.Renamed(strings.TrimSpace(name))
	}
	i.Name = strings.TrimSpace(name)
	return i
}

func (i ImportItem) IsRequest() bool {
	return i.Kind == importitemkind.Request
}

func (i ImportItem) IsVariables() bool {
	return i.Variables != nil
}

func (i ImportItem) WithWarnings(warnings []ImportWarning) ImportItem {
	all := append([]ImportWarning{}, i.Warnings...)
	for _, warning := range warnings {
		all = append(all, warning.ForItem(i.Key))
	}
	i.Warnings = all
	return i
}

func (i ImportItem) WithPmUsages(usages []PmUsage) ImportItem {
	tagged := make([]PmUsage, 0, len(usages))
	for _, usage := range usages {
		tagged = append(tagged, usage.ForItem(i.Key))
	}
	i.PmUsages = tagged
	return i
}

func (i ImportItem) Counts() ImportCounts {
	if !i.IsVariables() {
		return i.Root.Counts()
	}
	counts := ImportCounts{Variables: len(i.Variables.Rows())}
	if i.Kind == importitemkind.Environment {
		counts.Environments = 1
	}
	return counts
}

func (i ImportItem) Scripts() []ScriptRef {
	if i.IsVariables() {
		return []ScriptRef{}
	}
	return i.Root.Scripts("")
}

func (i ImportItem) Skipped(localId string) ImportOutcome {
	return ImportOutcome{Item: ImportItemResult{Status: importstatus.Skipped, Key: i.Key, LocalID: localId, Name: i.Name}, Skipped: true}
}

func (i ImportItem) Failed(localId string, failure ImportFailure) ImportOutcome {
	return ImportOutcome{Item: ImportItemResult{Status: importstatus.Failed, Key: i.Key, LocalID: localId, Name: i.Name}, Failure: &failure}
}

func (i ImportItem) Completed(write ImportWrite, location ImportLocation) ImportOutcome {
	move := i.relocation(write, location)
	ids := append(write.IDs(), i.Root.ID())
	outcome := ImportOutcome{
		Item:     ImportItemResult{Status: importstatus.Imported, Key: i.Key, LocalID: location.LocalID, ID: move.ToID, Name: lastPathSegment(move.ToPath)},
		Warnings: []ImportWarning{},
		PmUsages: []PmUsage{},
		Created:  write.Counts(),
		Replaced: len(write.Replaced) > 0,
	}
	if outcome.Replaced {
		outcome.Item.Status = importstatus.Replaced
	}
	for _, warning := range i.Warnings {
		if slices.Contains(ids, warning.ItemID) {
			outcome.Warnings = append(outcome.Warnings, warning.Relocated(move))
		}
	}
	outcome.Warnings = append(outcome.Warnings, i.placementWarnings(write, location, move)...)
	for _, usage := range i.PmUsages {
		if kept := usage.Within(ids); len(kept.Occurrences) > 0 {
			outcome.PmUsages = append(outcome.PmUsages, kept.Relocated(move))
		}
	}
	return outcome
}

func (i ImportItem) relocation(write ImportWrite, location ImportLocation) ImportRelocation {
	move := ImportRelocation{FromID: i.Root.ID(), ToID: i.Root.ID(), FromPath: i.Root.Name()}
	if write.Merge {
		move.ToID = write.Parent.ID
		move.ToPath = location.ParentPath
		return move
	}
	move.ToPath = JoinImportPath(location.ParentPath, write.Nodes[0].Name())
	return move
}

func (i ImportItem) placementWarnings(write ImportWrite, location ImportLocation, move ImportRelocation) []ImportWarning {
	warnings := make([]ImportWarning, 0, len(write.KeptVariables)+1)
	if write.Merge {
		message := "The collection's items were added to " + location.ParentPath + ", which now uses its auth, headers and scripts"
		warnings = append(warnings, NewImportWarning(warningcode.MergedIntoFolder, message, move.ToID, move.ToPath))
	} else if write.Wrapped {
		message := location.ParentPath + " has auth, headers or scripts of its own, so the items were placed in the new folder " + lastPathSegment(move.ToPath)
		warnings = append(warnings, NewImportWarning(warningcode.WrappedInFolder, message, move.ToID, move.ToPath))
	}
	for _, row := range write.KeptVariables {
		warnings = append(warnings, NewImportWarning(warningcode.VariableKept, keptVariableMessage(row), write.collectionID(), location.CollectionPath))
	}
	for index := range warnings {
		warnings[index] = warnings[index].ForItem(i.Key)
	}
	return warnings
}

func (i ImportItem) CompletedVariables(write ImportVariablesWrite, localId string) ImportOutcome {
	id, name, path := write.target(i.Name)
	outcome := ImportOutcome{
		Item:     ImportItemResult{Status: importstatus.Imported, Key: i.Key, LocalID: localId, ID: id, Name: name},
		Warnings: []ImportWarning{},
		PmUsages: []PmUsage{},
		Created:  write.Counts(),
		Replaced: write.Replaced != nil,
	}
	if outcome.Replaced {
		outcome.Item.Status = importstatus.Replaced
	}
	for _, row := range write.KeptVariables {
		outcome.Warnings = append(outcome.Warnings, NewImportWarning(warningcode.VariableKept, keptVariableMessage(row), id, path).ForItem(i.Key))
	}
	return outcome
}

func keptVariableMessage(row VariableRow) string {
	imported := row.InitialValue
	if row.Secret {
		imported = MaskedValue
	}
	return row.Key + " keeps its value; the imported value was " + imported
}

func (i ImportItem) LogValue() any {
	return map[string]any{"key": i.Key, "kind": i.Kind, "name": i.Name, "warnings": len(i.Warnings)}
}
