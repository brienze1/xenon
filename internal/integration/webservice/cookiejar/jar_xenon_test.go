package cookiejar

import (
	"errors"
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/brienze1/xenon/pkg/timeutils"
)

func setCookie(t *testing.T, jar *Jar, rawUrl, line string, now time.Time) (Entry, error) {
	t.Helper()
	cookie, err := http.ParseSetCookie(line)
	if err != nil {
		t.Fatalf("ParseSetCookie(%q): %v", line, err)
	}
	return jar.SetCookie(mustParseURL(rawUrl), cookie, now)
}

func names(entries []Entry) []string {
	result := make([]string, 0, len(entries))
	for _, entry := range entries {
		result = append(result, entry.Name)
	}
	return result
}

func TestSetCookieReturnsTheStoredEntryAndCountsRevisions(t *testing.T) {
	jar := New()
	entry, err := setCookie(t, jar, "https://api.example.test/v1/users", "a=1; Domain=example.test; Max-Age=60; Secure; HttpOnly; SameSite=Strict", tNow)
	if err != nil {
		t.Fatal(err)
	}
	want := Entry{Name: "a", Value: "1", Domain: "example.test", Path: "/v1", SameSite: "Strict", Secure: true, HttpOnly: true, Persistent: true, Expires: tNow.Add(time.Minute), Creation: tNow, LastAccess: tNow}
	if !reflect.DeepEqual(entry, want) {
		t.Errorf("entry = %+v, want %+v", entry, want)
	}
	if jar.Revision() != 1 {
		t.Errorf("revision = %d, want 1", jar.Revision())
	}
	updated, _ := setCookie(t, jar, "https://api.example.test/v1/users", "a=2; Domain=example.test; Max-Age=60", tNow.Add(time.Second))
	if updated.Creation != tNow || updated.seqNum != entry.seqNum || updated.LastAccess != tNow.Add(time.Second) || jar.Revision() != 2 {
		t.Errorf("an update keeps the creation and order: %+v, revision %d", updated, jar.Revision())
	}
}

func TestSameSiteModes(t *testing.T) {
	cases := map[string]string{"SameSite=Lax": "Lax", "SameSite=None": "None", "SameSite=Strict": "Strict", "SameSite": ""}
	for attribute, want := range cases {
		entry, err := setCookie(t, New(), "http://host.test", "a=1; "+attribute, tNow)
		if err != nil || entry.SameSite != want {
			t.Errorf("%s: SameSite = %q (%v), want %q", attribute, entry.SameSite, err, want)
		}
	}
}

func TestSetCookieRefusesInvalidUrlsAndDomainsAndRemovesExpiredCookies(t *testing.T) {
	jar := New()
	for _, rawUrl := range []string{"ftp://host.test/", "mailto:someone"} {
		if _, err := jar.SetCookie(mustParseURLAnyScheme(rawUrl), &http.Cookie{Name: "a", Value: "1"}, tNow); !errors.Is(err, ErrInvalidUrl) {
			t.Errorf("%s: err = %v", rawUrl, err)
		}
	}
	if _, err := Parse(nil, &http.Cookie{Name: "a"}, tNow); !errors.Is(err, ErrInvalidUrl) {
		t.Errorf("nil url: err = %v", err)
	}
	if _, err := setCookie(t, jar, "http://www.host.test", "a=1; Domain=other.test", tNow); !errors.Is(err, ErrIllegalDomain) {
		t.Errorf("foreign domain: err = %v", err)
	}
	if _, err := setCookie(t, jar, "http://www.host.test", "a=1; Domain=..host.test", tNow); !errors.Is(err, ErrMalformedDomain) {
		t.Errorf("malformed domain: err = %v", err)
	}
	if _, err := setCookie(t, jar, "http://www.host.test", "a=1; Max-Age=60", tNow); err != nil {
		t.Fatal(err)
	}
	revision := jar.Revision()
	if _, err := setCookie(t, jar, "http://www.host.test", "a=1; Max-Age=0", tNow); !errors.Is(err, ErrExpired) {
		t.Errorf("expired: err = %v", err)
	}
	if len(jar.All(tNow)) != 0 || jar.Revision() != revision+1 {
		t.Errorf("an expired cookie removes the stored one: %v, revision %d", jar.All(tNow), jar.Revision())
	}
	if _, err := setCookie(t, jar, "http://www.host.test", "b=1; Expires=Thu, 01 Jan 2015 00:00:00 GMT", tNow.Add(time.Hour*24*365*3)); !errors.Is(err, ErrExpired) {
		t.Errorf("expired by date: err = %v", err)
	}
}

func mustParseURLAnyScheme(raw string) *url.URL {
	u, _ := url.Parse(raw)
	return u
}

func TestDeleteRemovesOneCookieAndEmptyDomains(t *testing.T) {
	jar := New()
	_, _ = setCookie(t, jar, "http://host.test/a/b", "a=1", tNow)
	_, _ = setCookie(t, jar, "http://host.test/", "b=2", tNow)
	revision := jar.Revision()
	if jar.Delete("host.test", "/", "missing") || jar.Delete("other.test", "/", "b") || jar.Revision() != revision {
		t.Error("deleting what is not there changes nothing")
	}
	if !jar.Delete("host.test", "/a", "a") || jar.Revision() != revision+1 {
		t.Error("an existing cookie is deleted")
	}
	if !reflect.DeepEqual(names(jar.All(tNow)), []string{"b"}) {
		t.Errorf("All = %v", names(jar.All(tNow)))
	}
	jar.Delete("host.test", "/", "b")
	if len(jar.entries) != 0 {
		t.Errorf("an emptied domain is dropped: %v", jar.entries)
	}
}

func TestAllListsInCreationOrderAndDropsExpiredCookies(t *testing.T) {
	jar := New()
	_, _ = setCookie(t, jar, "http://b.test", "late=1; Max-Age=10", tNow)
	_, _ = setCookie(t, jar, "http://a.test", "first=1", tNow)
	_, _ = setCookie(t, jar, "http://b.test", "second=1; Max-Age=100", tNow)
	if got := names(jar.All(tNow)); !reflect.DeepEqual(got, []string{"late", "first", "second"}) {
		t.Errorf("All = %v", got)
	}
	revision := jar.Revision()
	if got := names(jar.All(tNow.Add(10 * time.Second))); !reflect.DeepEqual(got, []string{"first", "second"}) {
		t.Errorf("All after expiry = %v", got)
	}
	if jar.Revision() != revision+1 {
		t.Errorf("dropping an expired cookie is a change: %d", jar.Revision())
	}
}

func TestMatchingDropsExpiredCookiesAndKeepsTheRevisionForAccessOnly(t *testing.T) {
	jar := New()
	_, _ = setCookie(t, jar, "http://host.test", "a=1; Max-Age=10", tNow)
	_, _ = setCookie(t, jar, "http://host.test", "b=1", tNow)
	revision := jar.Revision()
	if got := names(jar.Matching(mustParseURL("http://host.test"), tNow.Add(time.Second))); !reflect.DeepEqual(got, []string{"a", "b"}) || jar.Revision() != revision {
		t.Errorf("Matching = %v, revision %d", got, jar.Revision())
	}
	if got := names(jar.Matching(mustParseURL("http://host.test"), tNow.Add(10*time.Second))); !reflect.DeepEqual(got, []string{"b"}) || jar.Revision() != revision+1 {
		t.Errorf("Matching after expiry = %v, revision %d", got, jar.Revision())
	}
	if got := jar.Matching(mustParseURLAnyScheme("ftp://host.test"), tNow); len(got) != 0 {
		t.Errorf("non-http URLs match nothing: %v", got)
	}
	if got := jar.All(tNow); got[0].LastAccess != tNow.Add(10*time.Second) {
		t.Errorf("a match records the access: %v", got[0].LastAccess)
	}
}

func TestImportAddsMissingCookiesWithoutCountingAChange(t *testing.T) {
	jar := New()
	_, _ = setCookie(t, jar, "http://host.test", "a=memory", tNow)
	revision := jar.Revision()
	stored := []Entry{
		{Name: "a", Value: "disk", Domain: "host.test", Path: "/", HostOnly: true, Expires: endOfTime},
		{Name: "b", Value: "disk", Domain: "host.test", Path: "/", HostOnly: true, Persistent: true, Expires: tNow.Add(time.Hour), Creation: tNow.Add(-time.Hour), LastAccess: tNow.Add(-time.Minute)},
		{Name: "c", Value: "old", Domain: "host.test", Path: "/", HostOnly: true, Persistent: true, Expires: tNow},
	}
	jar.Import(stored, tNow)
	all := jar.All(tNow)
	if jar.Revision() != revision || !reflect.DeepEqual(names(all), []string{"a", "b"}) || all[0].Value != "memory" {
		t.Errorf("All = %+v, revision %d", all, jar.Revision())
	}
	if all[1].Creation != tNow.Add(-time.Hour) || all[1].LastAccess != tNow.Add(-time.Minute) {
		t.Errorf("imported times are kept: %+v", all[1])
	}
}

func TestCookiesAndSetCookiesUseTheCurrentTime(t *testing.T) {
	timeutils.SetCurrentTimeProvider(func() time.Time { return tNow })
	t.Cleanup(func() { timeutils.SetCurrentTimeProvider(time.Now) })
	jar := New()
	target := mustParseURL("http://host.test/x")
	jar.SetCookies(target, []*http.Cookie{{Name: "a", Value: "1", MaxAge: 5}, {Name: "b", Value: "2", Quoted: true}})
	got := jar.Cookies(target)
	if len(got) != 2 || got[0].String() != "a=1" || got[1].String() != `b="2"` {
		t.Errorf("Cookies = %v", got)
	}
	if entries := jar.All(tNow); entries[0].Expires != tNow.Add(5*time.Second) {
		t.Errorf("Max-Age counts from the current time: %v", entries[0].Expires)
	}
}

func TestSecureCookiesGoToHttpsAndLoopbackOnly(t *testing.T) {
	cases := map[string]bool{
		"https://host.test":   true,
		"http://host.test":    false,
		"http://localhost":    true,
		"http://a.localhost.": true,
		"http://127.0.0.1":    true,
		"http://10.0.0.1":     false,
	}
	for rawUrl, sent := range cases {
		jar := New()
		if _, err := setCookie(t, jar, rawUrl, "a=1; Secure", tNow); err != nil {
			t.Fatalf("%s: %v", rawUrl, err)
		}
		if got := len(jar.Matching(mustParseURL(rawUrl), tNow)) == 1; got != sent {
			t.Errorf("%s: sent = %v, want %v", rawUrl, got, sent)
		}
	}
}

func TestCookiesWithAnUppercaseOrNonAsciiDomain(t *testing.T) {
	entry, err := setCookie(t, New(), "http://www.Example.TEST", "a=1; Domain=EXAMPLE.test", tNow)
	if err != nil || entry.Domain != "example.test" || entry.HostOnly {
		t.Errorf("entry = %+v, %v", entry, err)
	}
	if _, _, err := domainAndType("www.xn--bcher-kva.de", "bücher.de"); !errors.Is(err, ErrMalformedDomain) {
		t.Errorf("non-ASCII domain attribute: %v", err)
	}
	if _, err := setCookie(t, New(), "http://www.example.test", "a=1; Domain=example.test.", tNow); !errors.Is(err, ErrMalformedDomain) {
		t.Errorf("trailing dot: %v", err)
	}
}
