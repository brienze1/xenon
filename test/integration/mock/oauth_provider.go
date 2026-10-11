package mock

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"time"
)

type OAuthProviderConfig struct {
	TokenType        string `json:"token_type"`
	AuthorizeError   string `json:"authorize_error"`
	ErrorDescription string `json:"error_description"`
	TokenError       string `json:"token_error"`
	ExpiresIn        int    `json:"expires_in"`
	DelayMs          int    `json:"delay_ms"`
	NoRefreshToken   bool   `json:"no_refresh_token"`
	NoExpiry         bool   `json:"no_expiry"`
	WrongState       bool   `json:"wrong_state"`
	IdToken          bool   `json:"id_token"`
}

type OAuthTokenRequest struct {
	Header http.Header
	Query  url.Values
	Form   url.Values
}

type OAuthProvider struct {
	server         *httptest.Server
	codes          map[string]url.Values
	refreshTokens  map[string]bool
	authorizations []url.Values
	tokenRequests  []OAuthTokenRequest
	config         OAuthProviderConfig
	issued         int
	mu             sync.Mutex
}

var oauthProviderOnce sync.Once
var oauthProviderInstance *OAuthProvider

func NewOAuthProvider() *OAuthProvider {
	oauthProviderOnce.Do(func() {
		oauthProviderInstance = &OAuthProvider{}
		mux := http.NewServeMux()
		mux.HandleFunc("/authorize", oauthProviderInstance.authorize)
		mux.HandleFunc("/token", oauthProviderInstance.token)
		oauthProviderInstance.server = httptest.NewServer(mux)
		oauthProviderInstance.Reset()
	})
	return oauthProviderInstance
}

func (o *OAuthProvider) URL() string {
	return o.server.URL
}

func (o *OAuthProvider) Reset() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.config = OAuthProviderConfig{TokenType: "Bearer", ExpiresIn: 3600}
	o.codes = map[string]url.Values{}
	o.refreshTokens = map[string]bool{}
	o.authorizations = nil
	o.tokenRequests = nil
	o.issued = 0
}

func (o *OAuthProvider) Configure(raw string) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	return json.Unmarshal([]byte(raw), &o.config)
}

func (o *OAuthProvider) Authorizations() []url.Values {
	o.mu.Lock()
	defer o.mu.Unlock()
	return append([]url.Values{}, o.authorizations...)
}

func (o *OAuthProvider) TokenRequests(grantType string) []OAuthTokenRequest {
	o.mu.Lock()
	defer o.mu.Unlock()
	requests := make([]OAuthTokenRequest, 0)
	for _, request := range o.tokenRequests {
		if request.Form.Get("grant_type") == grantType || request.Query.Get("grant_type") == grantType {
			requests = append(requests, request)
		}
	}
	return requests
}

func (o *OAuthProvider) authorize(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	o.mu.Lock()
	o.authorizations = append(o.authorizations, query)
	config := o.config
	o.mu.Unlock()
	target, err := url.Parse(query.Get("redirect_uri"))
	if err != nil || target.Host == "" {
		http.Error(w, "invalid redirect_uri", http.StatusBadRequest)
		return
	}
	state := query.Get("state")
	if config.WrongState {
		state = "forged-" + state
	}
	answer := url.Values{}
	switch {
	case config.AuthorizeError != "":
		answer.Set("error", config.AuthorizeError)
		answer.Set("error_description", config.ErrorDescription)
		answer.Set("state", state)
	case query.Get("response_type") == "token":
		for key, values := range o.tokenAnswer(query.Get("scope")) {
			answer.Set(key, fmt.Sprint(values))
		}
		answer.Del("refresh_token")
		answer.Set("state", state)
		target.Fragment = answer.Encode()
		http.Redirect(w, r, target.String(), http.StatusFound)
		return
	default:
		code := o.newCode(query)
		answer.Set("code", code)
		answer.Set("state", state)
	}
	target.RawQuery = answer.Encode()
	http.Redirect(w, r, target.String(), http.StatusFound)
}

func (o *OAuthProvider) newCode(query url.Values) string {
	o.mu.Lock()
	defer o.mu.Unlock()
	code := fmt.Sprintf("code-%d", len(o.codes)+1)
	o.codes[code] = query
	return code
}

func (o *OAuthProvider) token(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	form, _ := url.ParseQuery(string(body))
	o.mu.Lock()
	o.tokenRequests = append(o.tokenRequests, OAuthTokenRequest{Header: r.Header.Clone(), Query: r.URL.Query(), Form: form})
	config := o.config
	o.mu.Unlock()
	time.Sleep(time.Duration(config.DelayMs) * time.Millisecond)
	if config.TokenError != "" {
		o.fail(w, config.TokenError, config.ErrorDescription)
		return
	}
	grant := form.Get("grant_type")
	if grant == "" {
		grant = r.URL.Query().Get("grant_type")
	}
	scope := form.Get("scope")
	switch grant {
	case "authorization_code":
		authorizedScope, problem := o.redeemCode(form)
		if problem != "" {
			o.fail(w, "invalid_grant", problem)
			return
		}
		scope = authorizedScope
	case "refresh_token":
		if !o.redeemRefreshToken(form.Get("refresh_token")) {
			o.fail(w, "invalid_grant", "the refresh token is not valid")
			return
		}
	case "client_credentials", "password":
	default:
		o.fail(w, "unsupported_grant_type", grant)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(o.tokenAnswer(scope))
}

func (o *OAuthProvider) redeemCode(form url.Values) (string, string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	authorization, found := o.codes[form.Get("code")]
	if !found {
		return "", "unknown code"
	}
	delete(o.codes, form.Get("code"))
	if authorization.Get("redirect_uri") != form.Get("redirect_uri") {
		return "", "redirect_uri mismatch"
	}
	return authorization.Get("scope"), pkceProblem(authorization, form.Get("code_verifier"))
}

func pkceProblem(authorization url.Values, verifier string) string {
	challenge := authorization.Get("code_challenge")
	if challenge == "" {
		return ""
	}
	if authorization.Get("code_challenge_method") == "plain" {
		if verifier != challenge {
			return "code_verifier mismatch"
		}
		return ""
	}
	sum := sha256.Sum256([]byte(verifier))
	if base64.RawURLEncoding.EncodeToString(sum[:]) != challenge {
		return "code_verifier mismatch"
	}
	return ""
}

func (o *OAuthProvider) redeemRefreshToken(token string) bool {
	o.mu.Lock()
	defer o.mu.Unlock()
	if !o.refreshTokens[token] {
		return false
	}
	delete(o.refreshTokens, token)
	return true
}

func (o *OAuthProvider) tokenAnswer(scope string) map[string]any {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.issued++
	answer := map[string]any{"access_token": fmt.Sprintf("access-%d", o.issued), "token_type": o.config.TokenType}
	if !o.config.NoExpiry {
		answer["expires_in"] = o.config.ExpiresIn
	}
	if scope != "" {
		answer["scope"] = scope
	}
	if !o.config.NoRefreshToken {
		refresh := fmt.Sprintf("refresh-%d", o.issued)
		o.refreshTokens[refresh] = true
		answer["refresh_token"] = refresh
	}
	if o.config.IdToken || strings.Contains(scope, "openid") {
		answer["id_token"] = fmt.Sprintf("id-%d", o.issued)
	}
	return answer
}

func (o *OAuthProvider) fail(w http.ResponseWriter, code, description string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": code, "error_description": description})
}
