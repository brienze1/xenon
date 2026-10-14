package entity

import "github.com/brienze1/xenon/internal/domain/enums/authtype"

// IsOwn is true when the auth is configured on the request itself rather than none or inherited.
func (a Auth) IsOwn() bool {
	return a.Type != authtype.None && a.Type != authtype.Inherit
}
