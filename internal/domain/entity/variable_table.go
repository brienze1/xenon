package entity

import (
	"maps"
	"slices"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/secretstate"
	"github.com/brienze1/xenon/internal/domain/enums/vaultstate"
)

const GitHistoryWarning = "git_history"

type VariableTable struct {
	Current  map[string]string
	Secrets  map[string]string
	Scope    ScopeRef
	Vault    enums.VaultState
	Rows     []VariableRow
	Warnings []string
}

func NewVariableTable(scope ScopeRef, rows []VariableRow, local LocalVariables, vault Vault) VariableTable {
	return VariableTable{
		Scope:    scope,
		Rows:     append([]VariableRow{}, rows...),
		Current:  copyValues(local.CurrentValues(scope)),
		Secrets:  vault.SecretsFor(scope),
		Vault:    vault.State,
		Warnings: []string{},
	}
}

func (v VariableTable) WithWarningsOf(other VariableTable) VariableTable {
	v.Warnings = other.Warnings
	return v
}

func (v VariableTable) Entries() []VariableTableEntry {
	entries := make([]VariableTableEntry, 0, len(v.Rows))
	for _, row := range v.Rows {
		entries = append(entries, v.entry(row))
	}
	return entries
}

func (v VariableTable) Entry(key string) (VariableTableEntry, bool) {
	row, found := v.row(key)
	if !found {
		return VariableTableEntry{}, false
	}
	return v.entry(row), true
}

func (v VariableTable) Contains(key string) bool {
	_, found := v.row(key)
	return found
}

func (v VariableTable) HasValidKeys() bool {
	for _, row := range v.Rows {
		if !row.HasValidKey() {
			return false
		}
	}
	return true
}

func (v VariableTable) Secret(key string) (string, bool) {
	row, found := v.row(key)
	if !found || !row.Secret {
		return "", false
	}
	value, stored := v.Secrets[key]
	return value, stored
}

// RevealedBy reports whether the edits turn a secret row into a plain one.
func (v VariableTable) RevealedBy(edits []VariableEdit) bool {
	for _, edit := range edits {
		if previous, found := v.row(edit.SourceKey()); found && previous.Secret && !edit.Row.Secret {
			return true
		}
	}
	return false
}

// Edited rebuilds the table from the edits, carrying current and secret values across renames and type switches.
func (v VariableTable) Edited(edits []VariableEdit) VariableTable {
	edited := v.withRows(make([]VariableRow, 0, len(edits)), map[string]string{}, map[string]string{})
	for _, edit := range edits {
		previous, existed := v.row(edit.SourceKey())
		if edit.Row.Secret {
			edited = edited.withSecretEdit(v, edit, previous, existed)
		} else {
			edited = edited.withPlainEdit(v, edit, previous, existed)
		}
	}
	return edited
}

func (v VariableTable) withSecretEdit(original VariableTable, edit VariableEdit, previous VariableRow, existed bool) VariableTable {
	row := edit.Row.WithoutSecretValue()
	value, stored := original.carriedSecret(previous, existed)
	if existed && (!previous.Secret || previous.HasStraySecretValue()) {
		v.Warnings = withWarning(v.Warnings, GitHistoryWarning)
	}
	if edit.SecretValue != nil {
		value, stored = *edit.SecretValue, *edit.SecretValue != ""
	}
	if stored {
		v.Secrets[row.Key] = value
	}
	v.Rows = append(v.Rows, row)
	return v
}

func (v VariableTable) withPlainEdit(original VariableTable, edit VariableEdit, previous VariableRow, existed bool) VariableTable {
	row := edit.Row
	if existed && previous.Secret {
		if value, stored := original.Secrets[previous.Key]; stored {
			row.InitialValue = value
		}
	}
	if existed && !previous.Secret {
		if value, overridden := original.Current[previous.Key]; overridden {
			v.Current[row.Key] = value
		}
	}
	if edit.SetCurrent {
		v.Current = withOverride(v.Current, row.Key, edit.CurrentValue)
	}
	v.Rows = append(v.Rows, row)
	return v
}

func (v VariableTable) carriedSecret(previous VariableRow, existed bool) (string, bool) {
	switch {
	case !existed:
		return "", false
	case previous.Secret:
		value, stored := v.Secrets[previous.Key]
		return value, stored
	default:
		value := v.currentValue(previous)
		return value, value != ""
	}
}

// WithCurrent stores a plain row's current value or a secret row's value; empty clears it.
func (v VariableTable) WithCurrent(key, value string) VariableTable {
	row, _ := v.row(key)
	if row.Secret {
		v.Secrets = withOverride(v.Secrets, key, value)
	} else {
		v.Current = withOverride(v.Current, key, value)
	}
	return v
}

// WithScriptChanges applies a script's writes: a known row gets the value as its current value (a secret row's value
// goes to the vault), a new name becomes a row with an empty initial value, and unset drops a plain row's current value.
// Secret changes (pm.vault) only touch secret rows.
func (v VariableTable) WithScriptChanges(set []KeyValue, unset []string, secret bool) VariableTable {
	for _, change := range set {
		row, found := v.row(change.Key)
		switch {
		case found && row.Secret != secret && secret:
			continue
		case found:
			v = v.WithCurrent(change.Key, change.Value)
		default:
			added := NewVariableRow(change.Key, "")
			added.Secret = secret
			v = v.Added(added, "").WithCurrent(change.Key, change.Value)
		}
	}
	for _, key := range unset {
		row, found := v.row(key)
		switch {
		case !found:
			continue
		case secret && row.Secret:
			v = v.WithCurrent(key, "")
		case !secret && !row.Secret:
			v = v.Reset([]string{key})
		}
	}
	return v
}

func (v VariableTable) Reset(keys []string) VariableTable {
	current := copyValues(v.Current)
	for _, row := range v.selected(keys) {
		if !row.Secret {
			delete(current, row.Key)
		}
	}
	v.Current = current
	return v
}

func (v VariableTable) PersistsSecret(keys []string) bool {
	for _, key := range keys {
		if row, found := v.row(key); found && row.Secret {
			return true
		}
	}
	return false
}

func (v VariableTable) Persisted(keys []string) VariableTable {
	current := copyValues(v.Current)
	rows := append([]VariableRow{}, v.Rows...)
	for index, row := range rows {
		value, overridden := current[row.Key]
		if !row.Secret && overridden && isSelectedKey(keys, row.Key) {
			rows[index].InitialValue = value
			delete(current, row.Key)
		}
	}
	v.Rows, v.Current = rows, current
	return v
}

func (v VariableTable) Added(row VariableRow, value string) VariableTable {
	if row.Secret {
		v = v.withRows(append(append([]VariableRow{}, v.Rows...), row), v.Current, v.Secrets)
		return v.WithCurrent(row.Key, value)
	}
	row.InitialValue = value
	return v.withRows(append(append([]VariableRow{}, v.Rows...), row), v.Current, v.Secrets)
}

// WithOverridesOfSecretRowsMoved moves plain current values of rows that became secret into the secrets, or drops them.
func (v VariableTable) WithOverridesOfSecretRowsMoved(keepValues bool) VariableTable {
	current := copyValues(v.Current)
	secrets := copyValues(v.Secrets)
	for key, value := range v.Current {
		if row, found := v.row(key); found && row.Secret {
			delete(current, key)
			if _, stored := secrets[key]; keepValues && !stored {
				secrets[key] = value
			}
		}
	}
	v.Current, v.Secrets = current, secrets
	return v
}

func (v VariableTable) StoredCurrent() map[string]string {
	return v.valuesOf(v.Current, false)
}

func (v VariableTable) StoredSecrets() map[string]string {
	return v.valuesOf(v.Secrets, true)
}

func (v VariableTable) CurrentDiffers(other VariableTable) bool {
	return !maps.Equal(v.StoredCurrent(), other.Current)
}

func (v VariableTable) SecretsDiffer(other VariableTable) bool {
	return !maps.Equal(v.StoredSecrets(), other.Secrets)
}

func (v VariableTable) LogValue() any {
	return map[string]any{"scope": v.Scope.String(), "rows": len(v.Rows), "vault": v.Vault, "warnings": v.Warnings}
}

func (v VariableTable) entry(row VariableRow) VariableTableEntry {
	entry := VariableTableEntry{Row: row.WithoutSecretValue(), Duplicate: v.isDuplicate(row.Key)}
	if row.Secret {
		_, entry.HasCurrentOverride = v.Secrets[row.Key]
		entry.SecretState = secretState(v.Vault, entry.HasCurrentOverride)
		return entry
	}
	_, entry.HasCurrentOverride = v.Current[row.Key]
	entry.CurrentValue = v.currentValue(row)
	return entry
}

func (v VariableTable) currentValue(row VariableRow) string {
	if value, overridden := v.Current[row.Key]; overridden {
		return value
	}
	return row.InitialValue
}

func (v VariableTable) row(key string) (VariableRow, bool) {
	for index := len(v.Rows) - 1; index >= 0; index-- {
		if v.Rows[index].Key == key {
			return v.Rows[index], true
		}
	}
	return VariableRow{}, false
}

func isSelectedKey(keys []string, key string) bool {
	return len(keys) == 0 || slices.Contains(keys, key)
}

func (v VariableTable) selected(keys []string) []VariableRow {
	selected := make([]VariableRow, 0, len(v.Rows))
	for _, row := range v.Rows {
		if isSelectedKey(keys, row.Key) {
			selected = append(selected, row)
		}
	}
	return selected
}

func (v VariableTable) isDuplicate(key string) bool {
	count := 0
	for _, row := range v.Rows {
		if row.Key == key {
			count++
		}
	}
	return count > 1
}

func (v VariableTable) valuesOf(values map[string]string, secret bool) map[string]string {
	stored := map[string]string{}
	for _, row := range v.Rows {
		if value, found := values[row.Key]; found && row.Secret == secret {
			stored[row.Key] = value
		}
	}
	return stored
}

func (v VariableTable) withRows(rows []VariableRow, current, secrets map[string]string) VariableTable {
	v.Rows, v.Current, v.Secrets = rows, copyValues(current), copyValues(secrets)
	v.Warnings = append([]string{}, v.Warnings...)
	return v
}

func secretState(vault enums.VaultState, stored bool) enums.SecretState {
	switch {
	case vault == vaultstate.Locked:
		return secretstate.Locked
	case stored:
		return secretstate.Set
	default:
		return secretstate.Missing
	}
}

func withOverride(values map[string]string, key, value string) map[string]string {
	updated := copyValues(values)
	delete(updated, key)
	if value != "" {
		updated[key] = value
	}
	return updated
}

func withWarning(warnings []string, warning string) []string {
	if slices.Contains(warnings, warning) {
		return warnings
	}
	return append(warnings, warning)
}
