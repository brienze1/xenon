package usecase

import (
	"strings"
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/scriptownerkind"
	"github.com/brienze1/xenon/internal/domain/enums/scriptphase"
)

func owner(kind enums.ScriptOwnerKind, name string, inherit bool, disabled ...enums.ScriptPhase) entity.ScriptOwner {
	return entity.NewScriptOwner(kind, "id-"+name, name, entity.Scripts{PreRequest: name + "-pre", PostResponse: name + "-post", Disabled: disabled}, inherit)
}

func chainNames(chain []entity.ScriptRef) string {
	names := make([]string, 0, len(chain))
	for _, ref := range chain {
		names = append(names, ref.Text)
	}
	return strings.Join(names, ",")
}

func TestCollectScriptChain(t *testing.T) {
	collection := owner(scriptownerkind.Collection, "C", false)
	folderA := owner(scriptownerkind.Folder, "A", true)
	folderB := owner(scriptownerkind.Folder, "B", true)
	request := owner(scriptownerkind.Request, "R", true)
	empty := entity.NewScriptOwner(scriptownerkind.Folder, "id-E", "E", entity.Scripts{}, true)
	cases := []struct {
		name   string
		phase  enums.ScriptPhase
		want   string
		owners []entity.ScriptOwner
	}{
		{"pre order", scriptphase.PreRequest, "C-pre,A-pre,B-pre,R-pre", []entity.ScriptOwner{collection, folderA, folderB, request}},
		{"post order", scriptphase.PostResponse, "C-post,A-post,B-post,R-post", []entity.ScriptOwner{collection, folderA, folderB, request}},
		{"inherit off on B", scriptphase.PreRequest, "B-pre,R-pre", []entity.ScriptOwner{collection, folderA, owner(scriptownerkind.Folder, "B", false), request}},
		{"inherit off on B post", scriptphase.PostResponse, "B-post,R-post", []entity.ScriptOwner{collection, folderA, owner(scriptownerkind.Folder, "B", false), request}},
		{"inherit off on request", scriptphase.PreRequest, "R-pre", []entity.ScriptOwner{collection, owner(scriptownerkind.Folder, "A", false), owner(scriptownerkind.Request, "R", false)}},
		{"inherit off on first folder", scriptphase.PreRequest, "A-pre,R-pre", []entity.ScriptOwner{collection, owner(scriptownerkind.Folder, "A", false), request}},
		{"disabled skipped", scriptphase.PreRequest, "C-pre,R-pre", []entity.ScriptOwner{collection, owner(scriptownerkind.Folder, "A", true, scriptphase.PreRequest), request}},
		{"disabled other phase kept", scriptphase.PostResponse, "C-post,A-post,R-post", []entity.ScriptOwner{collection, owner(scriptownerkind.Folder, "A", true, scriptphase.PreRequest), request}},
		{"empty scripts skipped", scriptphase.PreRequest, "C-pre,R-pre", []entity.ScriptOwner{collection, empty, request}},
		{"no owners", scriptphase.PreRequest, "", nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := chainNames(CollectScriptChain(tc.owners, tc.phase)); got != tc.want {
				t.Errorf("chain = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestCollectScriptChainRefs(t *testing.T) {
	chain := CollectScriptChain([]entity.ScriptOwner{owner(scriptownerkind.Collection, "C", true), owner(scriptownerkind.Request, "R", true)}, scriptphase.PostResponse)
	want := entity.ScriptRef{Phase: scriptphase.PostResponse, OwnerKind: scriptownerkind.Request, ItemID: "id-R", OwnerName: "R", Text: "R-post"}
	if len(chain) != 2 || chain[1] != want || chain[0].OwnerKind != scriptownerkind.Collection {
		t.Errorf("chain = %+v", chain)
	}
}

func TestMaskSecrets(t *testing.T) {
	cases := []struct {
		name    string
		text    string
		want    string
		secrets []string
		masked  bool
	}{
		{"no secrets", "token abc", "token abc", nil, false},
		{"empty secret ignored", "token abc", "token abc", []string{""}, false},
		{"every occurrence", "abc and abc", "••••• and •••••", []string{"abc"}, true},
		{"longest first", "secret-long", "•••••", []string{"secret", "secret-long"}, true},
		{"absent secret", "nothing", "nothing", []string{"zzz"}, false},
		{"several secrets", "u=al p=pw", "u=••••• p=•••••", []string{"al", "pw"}, true},
		{"in url header body", "https://x/?k=s3cr3t\nAuthorization: Bearer s3cr3t\n{\"t\":\"s3cr3t\"}", "https://x/?k=•••••\nAuthorization: Bearer •••••\n{\"t\":\"•••••\"}", []string{"s3cr3t"}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, masked := MaskSecrets(tc.text, tc.secrets)
			if got != tc.want || masked != tc.masked {
				t.Errorf("MaskSecrets = %q, %v; want %q, %v", got, masked, tc.want, tc.masked)
			}
		})
	}
}

func TestMaskSecretsKeepsCallerOrder(t *testing.T) {
	secrets := []string{"a", "abc"}
	MaskSecrets("abc", secrets)
	if secrets[0] != "a" || secrets[1] != "abc" {
		t.Errorf("caller slice reordered: %v", secrets)
	}
}
