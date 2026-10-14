package entity

import (
	"testing"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/authtype"
)

func TestAuthIsOwn(t *testing.T) {
	cases := map[enums.AuthType]bool{
		authtype.None:        false,
		authtype.Inherit:     false,
		authtype.Bearer:      true,
		authtype.Basic:       true,
		authtype.Unsupported: true,
		authtype.Unknown:     true,
	}
	for authType, own := range cases {
		if got := (Auth{Type: authType}).IsOwn(); got != own {
			t.Errorf("%s: IsOwn %v, want %v", authType, got, own)
		}
	}
}
