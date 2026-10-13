// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package cookiejar is a fork of net/http/cookiejar that lists, edits, exports and imports its entries.
package cookiejar

import (
	"cmp"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"slices"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/idna"
	"golang.org/x/net/publicsuffix"

	"github.com/brienze1/xenon/pkg/timeutils"
)

var (
	ErrIllegalDomain   = errors.New("the cookie's domain does not match the URL's host or is a public suffix")
	ErrMalformedDomain = errors.New("the cookie's domain is malformed")
	ErrExpired         = errors.New("the cookie has already expired")
	ErrInvalidUrl      = errors.New("the URL is not a valid http or https URL")
)

var endOfTime = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)

type Entry struct {
	Expires    time.Time
	Creation   time.Time
	LastAccess time.Time
	Name       string
	Value      string
	Domain     string
	Path       string
	SameSite   string
	seqNum     uint64
	Quoted     bool
	Secure     bool
	HttpOnly   bool
	Persistent bool
	HostOnly   bool
}

func (e *Entry) id() string {
	return fmt.Sprintf("%s;%s;%s", e.Domain, e.Path, e.Name)
}

func (e *Entry) shouldSend(https bool, host, path string) bool {
	return e.domainMatch(host) && e.pathMatch(path) && e.secureMatch(https)
}

func (e *Entry) domainMatch(host string) bool {
	if e.Domain == host {
		return true
	}
	return !e.HostOnly && hasDotSuffix(host, e.Domain)
}

func (e *Entry) pathMatch(requestPath string) bool {
	if requestPath == e.Path {
		return true
	}
	if strings.HasPrefix(requestPath, e.Path) {
		return e.Path[len(e.Path)-1] == '/' || requestPath[len(e.Path)] == '/'
	}
	return false
}

func (e *Entry) secureMatch(https bool) bool {
	if !e.Secure || https || isLocalhost(e.Domain) {
		return true
	}
	ip, err := netip.ParseAddr(e.Domain)
	return err == nil && ip.IsLoopback()
}

func (e *Entry) isExpired(now time.Time) bool {
	return e.Persistent && !e.Expires.After(now)
}

type Jar struct {
	entries    map[string]map[string]Entry
	nextSeqNum uint64
	revision   uint64
	mu         sync.Mutex
}

func New() *Jar {
	return &Jar{entries: make(map[string]map[string]Entry)}
}

func (j *Jar) Revision() uint64 {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.revision
}

func (j *Jar) Cookies(u *url.URL) []*http.Cookie {
	cookies := make([]*http.Cookie, 0)
	for _, e := range j.Matching(u, timeutils.Now()) {
		cookies = append(cookies, &http.Cookie{Name: e.Name, Value: e.Value, Quoted: e.Quoted})
	}
	return cookies
}

// Matching returns the entries sent to u at now, longest path first, then oldest first.
func (j *Jar) Matching(u *url.URL, now time.Time) []Entry {
	selected := make([]Entry, 0)
	host, err := urlHost(u)
	if err != nil {
		return selected
	}
	key := jarKey(host)
	https := u.Scheme == "https"
	path := u.Path
	if path == "" {
		path = "/"
	}

	j.mu.Lock()
	defer j.mu.Unlock()
	submap := j.entries[key]
	for id, e := range submap {
		if e.isExpired(now) {
			j.remove(key, id)
			continue
		}
		if e.shouldSend(https, host, path) {
			e.LastAccess = now
			submap[id] = e
			selected = append(selected, e)
		}
	}
	slices.SortFunc(selected, func(a, b Entry) int {
		if r := cmp.Compare(len(b.Path), len(a.Path)); r != 0 {
			return r
		}
		if r := a.Creation.Compare(b.Creation); r != 0 {
			return r
		}
		return cmp.Compare(a.seqNum, b.seqNum)
	})
	return selected
}

func (j *Jar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	for _, cookie := range cookies {
		_, _ = j.SetCookie(u, cookie, timeutils.Now())
	}
}

// SetCookie stores c as received from u at now; an expired c removes the stored cookie and returns ErrExpired.
func (j *Jar) SetCookie(u *url.URL, c *http.Cookie, now time.Time) (Entry, error) {
	e, err := Parse(u, c, now)
	if errors.Is(err, ErrExpired) {
		j.Delete(e.Domain, e.Path, e.Name)
	}
	if err != nil {
		return Entry{}, err
	}

	e.LastAccess = now
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.put(e), nil
}

// Parse returns the entry c received from u at now would be stored as, without storing it.
func Parse(u *url.URL, c *http.Cookie, now time.Time) (Entry, error) {
	host, err := urlHost(u)
	if err != nil {
		return Entry{}, err
	}
	return newEntry(c, now, defaultPath(u.Path), host)
}

func (j *Jar) put(e Entry) Entry {
	key := jarKey(e.Domain)
	submap := j.entries[key]
	if submap == nil {
		submap = make(map[string]Entry)
		j.entries[key] = submap
	}
	if old, found := submap[e.id()]; found {
		e.Creation = old.Creation
		e.seqNum = old.seqNum
	} else {
		e.seqNum = j.nextSeqNum
		j.nextSeqNum++
	}
	submap[e.id()] = e
	j.revision++
	return e
}

func (j *Jar) Delete(domain, path, name string) bool {
	j.mu.Lock()
	defer j.mu.Unlock()
	target := Entry{Domain: domain, Path: path, Name: name}
	key := jarKey(domain)
	if _, found := j.entries[key][target.id()]; !found {
		return false
	}
	j.remove(key, target.id())
	return true
}

func (j *Jar) remove(key, id string) {
	delete(j.entries[key], id)
	if len(j.entries[key]) == 0 {
		delete(j.entries, key)
	}
	j.revision++
}

// All returns every entry not expired at now, dropping the expired ones.
func (j *Jar) All(now time.Time) []Entry {
	j.mu.Lock()
	defer j.mu.Unlock()
	all := make([]Entry, 0)
	for key, submap := range j.entries {
		for id, e := range submap {
			if e.isExpired(now) {
				j.remove(key, id)
				continue
			}
			all = append(all, e)
		}
	}
	slices.SortFunc(all, func(a, b Entry) int { return cmp.Compare(a.seqNum, b.seqNum) })
	return all
}

// Import adds the entries the jar does not hold yet without counting as a change.
func (j *Jar) Import(entries []Entry, now time.Time) {
	j.mu.Lock()
	defer j.mu.Unlock()
	revision := j.revision
	for _, e := range entries {
		if _, found := j.entries[jarKey(e.Domain)][e.id()]; !found && !e.isExpired(now) {
			j.put(e)
		}
	}
	j.revision = revision
}

func urlHost(u *url.URL) (string, error) {
	if u == nil || (u.Scheme != "http" && u.Scheme != "https") {
		return "", ErrInvalidUrl
	}
	return canonicalHost(u.Host)
}

func isLocalhost(host string) bool {
	host = strings.TrimSuffix(host, ".")
	if idx := strings.LastIndex(host, "."); idx >= 0 {
		host = host[idx+1:]
	}
	return strings.EqualFold(host, "localhost")
}

func hasDotSuffix(s, suffix string) bool {
	return len(s) > len(suffix) && s[len(s)-len(suffix)-1] == '.' && s[len(s)-len(suffix):] == suffix
}

func canonicalHost(host string) (string, error) {
	var err error
	if hasPort(host) {
		host, _, err = net.SplitHostPort(host)
		if err != nil {
			return "", err
		}
	}
	host = strings.TrimSuffix(host, ".")
	encoded, err := idna.Punycode.ToASCII(host)
	if err != nil {
		return "", err
	}
	return strings.ToLower(encoded), nil
}

func hasPort(host string) bool {
	colons := strings.Count(host, ":")
	if colons == 0 {
		return false
	}
	if colons == 1 {
		return true
	}
	return host[0] == '[' && strings.Contains(host, "]:")
}

func jarKey(host string) string {
	if isIP(host) {
		return host
	}
	suffix := publicsuffix.List.PublicSuffix(host)
	i := len(host) - len(suffix)
	if suffix == host || host[i-1] != '.' {
		return host
	}
	prevDot := strings.LastIndex(host[:i-1], ".")
	return host[prevDot+1:]
}

func isIP(host string) bool {
	if strings.ContainsAny(host, ":%") {
		return true
	}
	return net.ParseIP(host) != nil
}

func defaultPath(path string) string {
	if len(path) == 0 || path[0] != '/' {
		return "/"
	}
	i := strings.LastIndex(path, "/")
	if i == 0 {
		return "/"
	}
	return path[:i]
}

func newEntry(c *http.Cookie, now time.Time, defPath, host string) (Entry, error) {
	e := Entry{Name: c.Name, Path: c.Path}
	if c.Path == "" || c.Path[0] != '/' {
		e.Path = defPath
	}
	var err error
	e.Domain, e.HostOnly, err = domainAndType(host, c.Domain)
	if err != nil {
		return e, err
	}
	switch {
	case c.MaxAge < 0:
		return e, ErrExpired
	case c.MaxAge > 0:
		e.Expires, e.Persistent = now.Add(time.Duration(c.MaxAge)*time.Second), true
	case c.Expires.IsZero():
		e.Expires = endOfTime
	case !c.Expires.After(now):
		return e, ErrExpired
	default:
		e.Expires, e.Persistent = c.Expires, true
	}
	e.Value, e.Quoted, e.Secure, e.HttpOnly = c.Value, c.Quoted, c.Secure, c.HttpOnly
	e.Creation, e.SameSite = now, sameSite(c.SameSite)
	return e, nil
}

func sameSite(mode http.SameSite) string {
	switch mode {
	case http.SameSiteStrictMode:
		return "Strict"
	case http.SameSiteLaxMode:
		return "Lax"
	case http.SameSiteNoneMode:
		return "None"
	default:
		return ""
	}
}

func domainAndType(host, domain string) (string, bool, error) {
	if domain == "" {
		return host, true, nil
	}
	if isIP(host) {
		if host != domain {
			return "", false, ErrIllegalDomain
		}
		return host, true, nil
	}
	domain = strings.TrimPrefix(domain, ".")
	if domain == "" || domain[0] == '.' || domain[len(domain)-1] == '.' || !isASCII(domain) {
		return "", false, ErrMalformedDomain
	}
	domain = strings.ToLower(domain)
	if ps := publicsuffix.List.PublicSuffix(domain); ps != "" && !hasDotSuffix(domain, ps) {
		if host == domain {
			return host, true, nil
		}
		return "", false, ErrIllegalDomain
	}
	if host != domain && !hasDotSuffix(host, domain) {
		return "", false, ErrIllegalDomain
	}
	return domain, false, nil
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= 0x80 {
			return false
		}
	}
	return true
}
