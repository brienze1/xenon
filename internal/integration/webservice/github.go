package webservice

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/signinstatus"
	"github.com/brienze1/xenon/internal/infra/properties"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/webservice/dto"
	"github.com/brienze1/xenon/pkg/errs"
	httpClient "github.com/brienze1/xenon/pkg/http"
	"github.com/brienze1/xenon/pkg/idutils"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/timeutils"
)

const (
	gitHubTimeout = 30 * time.Second
	gitHubJson    = "application/json"
	gitHubForm    = "application/x-www-form-urlencoded"
	gitHubAccept  = "application/vnd.github+json"
)

type gitHubWebservice struct {
	newClient httpClient.ClientFactory
	browser   adapter.Browser
	sessions  map[string]entity.GitHubDeviceSession
	mu        sync.Mutex
}

func NewGitHubWebservice(newClient httpClient.ClientFactory, browser adapter.Browser) adapter.GitHubWebservice {
	return &gitHubWebservice{newClient: newClient, browser: browser, sessions: map[string]entity.GitHubDeviceSession{}}
}

func (g *gitHubWebservice) Start(ctx context.Context) (*entity.GitHubDeviceCode, error) {
	logger.Debug(ctx, "Started")

	clientId := properties.Properties().Git.GitHubClientID
	if clientId == "" {
		return nil, remoteFailure("GitHub sign-in is not configured in this build of Xenon", "ACC-01503", http.StatusServiceUnavailable,
			"Add the account with a personal access token instead")
	}
	status, body, err := g.send(ctx, http.MethodPost, properties.Properties().Git.GitHubLoginURL+"/login/device/code", gitHubForm, dto.NewGitHubDeviceCodeRequest(clientId).Form(), "")
	if err != nil {
		return nil, errs.ServiceUnavailableError(err, "ACC-44503", "GitHub could not be reached")
	}
	var answer dto.GitHubDeviceCodeResponse
	if status != http.StatusOK || json.Unmarshal(body, &answer) != nil || answer.DeviceCode == "" {
		return nil, errs.ServerError("GitHub refused to start the sign-in", "ACC-45502", http.StatusBadGateway)
	}
	sessionId := idutils.NewUlid()
	g.mu.Lock()
	g.sessions[sessionId] = answer.Session(sessionId, timeutils.Now())
	g.mu.Unlock()
	if err := g.browser.OpenURL(answer.VerificationURI); err != nil {
		logger.Error(ctx, err, "The verification page could not be opened", map[string]any{"uri": answer.VerificationURI})
	}
	code := answer.ToEntity(sessionId)

	logger.Debug(ctx, "Finished", code)
	return &code, nil
}

func (g *gitHubWebservice) Poll(ctx context.Context, sessionId string) (*entity.GitHubSignIn, error) {
	logger.Debug(ctx, "Started", map[string]any{"sessionId": sessionId})

	g.mu.Lock()
	session, found := g.sessions[sessionId]
	g.mu.Unlock()
	if !found {
		return nil, errs.NotFoundError("no GitHub sign-in is waiting with this session; start again", "ACC-14404")
	}
	now := timeutils.Now()
	if session.IsExpired(now) {
		g.forget(sessionId)
		return &entity.GitHubSignIn{Status: signinstatus.Expired, Interval: session.IntervalSeconds()}, nil
	}
	if session.IsTooEarly(now) {
		logger.Debug(ctx, "Finished", map[string]any{"sessionId": sessionId, "tooEarly": true})
		return &entity.GitHubSignIn{Status: signinstatus.Pending, Interval: session.IntervalSeconds()}, nil
	}
	form := dto.NewGitHubTokenPollRequest(properties.Properties().Git.GitHubClientID, session.DeviceCode).Form()
	status, body, err := g.send(ctx, http.MethodPost, properties.Properties().Git.GitHubLoginURL+"/login/oauth/access_token", gitHubForm, form, "")
	if err != nil {
		return nil, errs.ServiceUnavailableError(err, "ACC-46503", "GitHub could not be reached")
	}
	var answer dto.GitHubTokenPollResponse
	if status != http.StatusOK || json.Unmarshal(body, &answer) != nil || answer.Status() == signinstatus.Unknown {
		return nil, errs.ServerError("GitHub answered the sign-in poll unexpectedly", "ACC-47502", http.StatusBadGateway)
	}
	session = session.Polled(now)
	if answer.IsSlowDown() {
		session = session.SlowedDown(now)
	}
	signIn := answer.ToEntity(session)
	g.mu.Lock()
	g.sessions[sessionId] = session
	g.mu.Unlock()
	if signIn.Status != signinstatus.Pending {
		g.forget(sessionId)
	}

	logger.Debug(ctx, "Finished", signIn)
	return &signIn, nil
}

func (g *gitHubWebservice) FindUser(ctx context.Context, token string) (*entity.GitHubUser, error) {
	logger.Debug(ctx, "Started")

	api := properties.Properties().Git.GitHubApiURL
	status, body, err := g.send(ctx, http.MethodGet, api+"/user", "", "", token)
	if err != nil {
		return nil, errs.ServiceUnavailableError(err, "ACC-48503", "GitHub could not be reached")
	}
	var answer dto.GitHubUserResponse
	if status == http.StatusUnauthorized || status == http.StatusForbidden {
		return nil, remoteFailure("GitHub refused the token", "ACC-17401", http.StatusUnauthorized, "Create a new token or sign in again")
	}
	if status != http.StatusOK || json.Unmarshal(body, &answer) != nil || answer.Login == "" {
		return nil, errs.ServerError("GitHub answered the user request unexpectedly", "ACC-49502", http.StatusBadGateway)
	}
	emails := make([]dto.GitHubEmail, 0)
	if status, body, err := g.send(ctx, http.MethodGet, api+"/user/emails", "", "", token); err == nil && status == http.StatusOK {
		_ = json.Unmarshal(body, &emails)
	}
	user := answer.ToEntity(emails)

	logger.Debug(ctx, "Finished", user)
	return &user, nil
}

func (g *gitHubWebservice) UploadKey(ctx context.Context, token, title, publicKey string) error {
	logger.Debug(ctx, "Started", map[string]any{"title": title})

	payload, _ := json.Marshal(dto.NewGitHubKeyRequest(title, publicKey))
	status, body, err := g.send(ctx, http.MethodPost, properties.Properties().Git.GitHubApiURL+"/user/keys", gitHubJson, string(payload), token)
	switch {
	case err != nil:
		return errs.ServiceUnavailableError(err, "ACC-50503", "GitHub could not be reached")
	case status == http.StatusUnprocessableEntity:
		return remoteFailure("GitHub refused the key: "+strings.TrimSpace(string(body)), "ACC-13422", http.StatusUnprocessableEntity,
			"The key may already be on a GitHub account; remove it there or generate another key")
	case status == http.StatusUnauthorized || status == http.StatusForbidden:
		return remoteFailure("GitHub refused the token for adding keys", "ACC-51401", http.StatusUnauthorized,
			"Sign in again so Xenon may add SSH keys (write:public_key)")
	case status != http.StatusCreated && status != http.StatusOK:
		return errs.ServerError("GitHub answered the key upload unexpectedly", "ACC-52502", http.StatusBadGateway)
	}

	logger.Debug(ctx, "Finished", map[string]any{"title": title})
	return nil
}

func (g *gitHubWebservice) forget(sessionId string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.sessions, sessionId)
}

func (g *gitHubWebservice) send(ctx context.Context, method, target, contentType, body, token string) (int, []byte, error) {
	request, err := http.NewRequestWithContext(ctx, method, target, bytes.NewReader([]byte(body)))
	if err != nil {
		return 0, nil, err
	}
	request.Header.Set("Accept", gitHubJson)
	if token != "" {
		request.Header.Set("Accept", gitHubAccept)
		request.Header.Set("Authorization", "Bearer "+token)
	}
	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}
	client := g.newClient.Build(httpClient.Options{Timeout: gitHubTimeout, FollowRedirects: true, VerifySsl: true})
	defer client.CloseIdleConnections()
	response, err := client.Do(ctx, request)
	if err != nil {
		return 0, nil, err
	}
	defer func() { _ = response.Body.Close() }()
	content, err := io.ReadAll(response.Body)
	return response.StatusCode, content, err
}
