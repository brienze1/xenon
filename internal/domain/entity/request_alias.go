package entity

import (
	"slices"
	"strings"
)

const (
	reservedAliasName   = "default"
	aliasOwnerSeparator = ":"
)

// RequestAlias is a named, independent version of a request's sendable content.
type RequestAlias struct {
	ID             string
	Name           string
	EnvironmentIDs []string
	Version        Request
}

func NewRequestAlias(id, name string, of Request, environmentIds []string) RequestAlias {
	return RequestAlias{ID: id, Name: strings.TrimSpace(name), EnvironmentIDs: NormalizedEnvironmentIDs(environmentIds), Version: Request{}.SendableOf(of)}
}

// AliasAuthOwner is the owner id of an alias's own auth, distinct from the Default's.
func AliasAuthOwner(requestId, aliasId string) string {
	return requestId + aliasOwnerSeparator + aliasId
}

// SplitAuthOwner reads an owner id back into its request and alias; the alias is empty for any other owner.
func SplitAuthOwner(owner string) (string, string) {
	requestId, aliasId, found := strings.Cut(owner, aliasOwnerSeparator)
	if !found {
		return owner, ""
	}
	return requestId, aliasId
}

func (a RequestAlias) HasValidName() bool {
	return isValidItemName(a.Name)
}

func (a RequestAlias) HasReservedName() bool {
	return strings.EqualFold(strings.TrimSpace(a.Name), reservedAliasName)
}

func (a RequestAlias) HasName(name string) bool {
	return strings.EqualFold(strings.TrimSpace(a.Name), strings.TrimSpace(name))
}

func (a RequestAlias) Renamed(name string) RequestAlias {
	a.Name = strings.TrimSpace(name)
	return a
}

// Edited takes the edit's sendable content; scripts not edited in the app stay the stored ones.
func (a RequestAlias) Edited(edit Request) RequestAlias {
	a.Version = Request{}.SendableOf(edit.WithScriptsOf(a.Version))
	return a
}

// IsLinked is an alias limited to some environments rather than available in all of them.
func (a RequestAlias) IsLinked() bool {
	return len(a.EnvironmentIDs) > 0
}

// IsAvailableIn is an alias for all environments, or one whose environments include the active one.
func (a RequestAlias) IsAvailableIn(active *Environment) bool {
	if !a.IsLinked() {
		return true
	}
	return active != nil && slices.Contains(a.EnvironmentIDs, active.ID)
}

func (a RequestAlias) WithEnvironments(ids []string) RequestAlias {
	a.EnvironmentIDs = NormalizedEnvironmentIDs(ids)
	return a
}
