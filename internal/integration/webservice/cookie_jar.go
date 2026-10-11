package webservice

import (
	"context"
	"net/http"
	"net/url"
	"path/filepath"
	"sync"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/infra/properties"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/webservice/cookiejar"
	"github.com/brienze1/xenon/internal/integration/webservice/dto"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/timeutils"
)

type workspaceJar struct {
	jar           *cookiejar.Jar
	savedRevision uint64
	loaded        bool
}

type cookieJarWebservice struct {
	jars map[string]*workspaceJar
	mu   sync.Mutex
}

type storingJar struct {
	jar *cookiejar.Jar
}

func NewCookieJarWebservice() adapter.CookieJarWebservice {
	return &cookieJarWebservice{jars: map[string]*workspaceJar{}}
}

func (c *cookieJarWebservice) JarState(ctx context.Context, localId string) entity.CookieJarState {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId})

	workspace := c.workspace(localId)
	c.mu.Lock()
	loaded, saved := workspace.loaded, workspace.savedRevision
	c.mu.Unlock()
	state := entity.CookieJarState{
		Cookies:       dto.NewJarCookies(workspace.jar.All(timeutils.Now())),
		Revision:      workspace.jar.Revision(),
		SavedRevision: saved,
		Loaded:        loaded,
	}

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "cookies": len(state.Cookies), "revision": state.Revision})
	return state
}

func (c *cookieJarWebservice) LoadCookies(ctx context.Context, localId string, cookies []entity.Cookie) {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId, "cookies": len(cookies)})

	workspace := c.workspace(localId)
	entries := make([]cookiejar.Entry, 0, len(cookies))
	for _, cookie := range cookies {
		entries = append(entries, dto.NewJarCookie(cookie).Entry())
	}
	workspace.jar.Import(entries, timeutils.Now())
	c.mu.Lock()
	workspace.loaded = true
	c.mu.Unlock()

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId})
}

func (c *cookieJarWebservice) MarkCookiesSaved(ctx context.Context, localId string, revision uint64) {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId, "revision": revision})

	workspace := c.workspace(localId)
	c.mu.Lock()
	workspace.savedRevision = revision
	c.mu.Unlock()

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "revision": revision})
}

func (c *cookieJarWebservice) CookiesFor(ctx context.Context, localId, rawUrl string) ([]entity.Cookie, error) {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId, "url": rawUrl})

	target, err := url.Parse(rawUrl)
	if err != nil {
		return nil, errs.BadRequestError("the URL is not valid: "+err.Error(), "CKE-10400")
	}
	cookies := dto.NewJarCookies(c.workspace(localId).jar.Matching(target, timeutils.Now()))

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "url": rawUrl, "cookies": len(cookies)})
	return cookies, nil
}

func (c *cookieJarWebservice) SetCookie(ctx context.Context, localId, rawUrl, setCookie string) (*entity.Cookie, error) {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId, "url": rawUrl})

	entry, err := c.store(localId, rawUrl, setCookie)
	if err != nil {
		return nil, errs.BadRequestError("the cookie is not valid: "+err.Error(), "CKE-00400")
	}
	cookie := dto.JarCookie(entry).ToEntity()

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "url": rawUrl, "result": cookie})
	return &cookie, nil
}

func (c *cookieJarWebservice) store(localId, rawUrl, setCookie string) (cookiejar.Entry, error) {
	target, err := url.Parse(rawUrl)
	if err != nil {
		return cookiejar.Entry{}, err
	}
	parsed, err := http.ParseSetCookie(setCookie)
	if err != nil {
		return cookiejar.Entry{}, err
	}
	return c.workspace(localId).jar.SetCookie(target, parsed, timeutils.Now())
}

func (c *cookieJarWebservice) RemoveCookies(ctx context.Context, localId string, cookies []entity.Cookie) {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId, "cookies": len(cookies)})

	jar := c.workspace(localId).jar
	for _, cookie := range cookies {
		jar.Delete(cookie.Domain, cookie.Path, cookie.Name)
	}

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "cookies": len(cookies)})
}

func (c *cookieJarWebservice) HttpJar(localId string, sendCookies bool) http.CookieJar {
	jar := c.workspace(localId).jar
	if sendCookies {
		return jar
	}
	return storingJar{jar: jar}
}

func (c *cookieJarWebservice) workspace(localId string) *workspaceJar {
	key := filepath.Join(properties.Properties().Application.DataDir, localId)
	c.mu.Lock()
	defer c.mu.Unlock()
	workspace, found := c.jars[key]
	if !found {
		workspace = &workspaceJar{jar: cookiejar.New()}
		c.jars[key] = workspace
	}
	return workspace
}

func (s storingJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	s.jar.SetCookies(u, cookies)
}

func (s storingJar) Cookies(*url.URL) []*http.Cookie {
	return nil
}
