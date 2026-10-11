package webservice

import (
	"context"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/infra/properties"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/logger"
)

const (
	cancelledStatus       = 499
	callbackHeaderTimeout = 10 * time.Second
	callbackDonePage      = `<!doctype html><meta charset="utf-8"><title>Xenon</title><p>%s</p><p>You can close this window and return to Xenon.</p>`
	callbackFragmentPage  = `<!doctype html><meta charset="utf-8"><title>Xenon</title><p>Finishing the sign-in…</p><script>if (location.hash.length > 1) { location.replace(location.pathname + "?" + location.hash.substring(1)); }</script>`
)

type oauthBrowserWebservice struct {
	browser  adapter.Browser
	loopback adapter.Loopback
}

func NewOAuthBrowserWebservice(browser adapter.Browser, loopback adapter.Loopback) adapter.OAuthBrowserWebservice {
	return &oauthBrowserWebservice{browser: browser, loopback: loopback}
}

func (o *oauthBrowserWebservice) Authorize(ctx context.Context, authorizationUrl, redirectUri string) (*entity.OAuth2Callback, error) {
	logger.Debug(ctx, "Started", map[string]any{"redirectUri": redirectUri})

	target, err := url.Parse(redirectUri)
	if err != nil {
		return nil, errs.BadRequestError("the callback URL is not valid", "OAU-13400")
	}
	listener, err := o.loopback.Listen(target.Host)
	if err != nil {
		return nil, errs.ConflictError("the callback port "+target.Port()+" is in use; close the program using it or change the callback URL", "OAU-02409")
	}
	callbacks := make(chan entity.OAuth2Callback, 1)
	server := &http.Server{Handler: callbackHandler(target.Path, callbacks), ReadHeaderTimeout: callbackHeaderTimeout}
	go func() { _ = server.Serve(listener) }()
	defer func() {
		_ = server.Close()
		_ = listener.Close()
	}()
	if err := o.browser.OpenURL(authorizationUrl); err != nil {
		return nil, errs.InternalServerError(err, "OAU-09500", "the system browser could not be opened")
	}
	select {
	case callback := <-callbacks:
		logger.Debug(ctx, "Finished", callback)
		return &callback, nil
	case <-time.After(properties.Properties().OAuth.CallbackTimeout):
		return nil, errs.ClientError("no sign-in came back from the browser in time; try again", "OAU-03408", http.StatusRequestTimeout)
	case <-ctx.Done():
		return nil, errs.ClientError("the sign-in was cancelled", "OAU-08499", cancelledStatus)
	}
}

func callbackHandler(path string, callbacks chan<- entity.OAuth2Callback) http.Handler {
	if path == "" {
		path = "/"
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if r.URL.Path != path {
			http.NotFound(w, r)
			return
		}
		query := r.URL.Query()
		if !query.Has("code") && !query.Has("error") && !query.Has("access_token") {
			_, _ = w.Write([]byte(callbackFragmentPage))
			return
		}
		callback := entity.OAuth2Callback{
			Code:             query.Get("code"),
			State:            query.Get("state"),
			Error:            query.Get("error"),
			ErrorDescription: query.Get("error_description"),
			AccessToken:      query.Get("access_token"),
			TokenType:        query.Get("token_type"),
			ExpiresIn:        query.Get("expires_in"),
			Scope:            query.Get("scope"),
			IDToken:          query.Get("id_token"),
		}
		select {
		case callbacks <- callback:
		default:
		}
		message := "Xenon received the sign-in."
		if callback.IsError() {
			message = "The sign-in failed: " + callback.Error
		}
		_, _ = fmt.Fprintf(w, callbackDonePage, html.EscapeString(message))
	})
}
