package entity

import "testing"

func TestMergeCookiesReplacesTheSameCookieAndAddsOthers(t *testing.T) {
	workspace := []Cookie{
		{Name: "session", Value: "old", Domain: "api.test", Path: "/"},
		{Name: "session", Value: "other-path", Domain: "api.test", Path: "/admin"},
		{Name: "theme", Value: "dark", Domain: "api.test", Path: "/"},
	}
	run := []Cookie{
		{Name: "session", Value: "new", Domain: "API.test", Path: "/"},
		{Name: "cart", Value: "1", Domain: "api.test", Path: "/"},
	}
	merged := MergeCookies(workspace, run)
	if len(merged) != 4 {
		t.Fatalf("merged = %+v", merged)
	}
	if merged[0].Value != "new" || merged[1].Value != "other-path" || merged[2].Value != "dark" || merged[3].Name != "cart" {
		t.Errorf("merged = %+v", merged)
	}
	if workspace[0].Value != "old" {
		t.Error("the workspace cookies are not changed in place")
	}
	if len(MergeCookies(nil, nil)) != 0 {
		t.Error("nothing merged is empty")
	}
}
