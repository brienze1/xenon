package entity

import "testing"

func TestCookiesMatchingFiltersDomainsIgnoringCase(t *testing.T) {
	domains := []CookieDomain{
		{Domain: "api.example.com", Cookies: []Cookie{{Name: "a"}, {Name: "b"}}},
		{Domain: "other.dev", Cookies: []Cookie{{Name: "c"}}},
	}
	if all := CookiesMatching(domains, ""); len(all) != 3 {
		t.Errorf("no filter = %d cookies, want 3", len(all))
	}
	if matched := CookiesMatching(domains, "API.Example.com"); len(matched) != 2 || matched[0].Name != "a" {
		t.Errorf("filtered = %+v", matched)
	}
	if none := CookiesMatching(domains, "missing.dev"); none == nil || len(none) != 0 {
		t.Errorf("unknown domain = %+v, want an empty list", none)
	}
}

func TestCookieAndConsoleActivityTargets(t *testing.T) {
	cookie := Cookie{Name: "session", Domain: "example.com", Path: "/api"}
	if target := cookie.ActivityTarget(); target != (AIActivityTarget{Name: "session", Path: "example.com/api"}) {
		t.Errorf("cookie target = %+v", target)
	}
	if target := (ConsoleEntry{ID: "c1"}).ActivityTarget(); target != (AIActivityTarget{ID: "c1"}) {
		t.Errorf("console target = %+v", target)
	}
}
