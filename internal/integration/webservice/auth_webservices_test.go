package webservice

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/bodymode"
	"github.com/brienze1/xenon/internal/domain/enums/clientauthentication"
	"github.com/brienze1/xenon/internal/domain/enums/digestalgorithm"
	"github.com/brienze1/xenon/internal/domain/enums/digestqop"
	"github.com/brienze1/xenon/internal/domain/enums/oauth2grant"
	"github.com/brienze1/xenon/internal/domain/enums/rawlanguage"
	"github.com/brienze1/xenon/pkg/errs"
	httpClient "github.com/brienze1/xenon/pkg/http"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/timeutils"
)

func codeOf(t *testing.T, err error) string {
	t.Helper()
	var custom errs.BaseError
	if !errors.As(err, &custom) {
		t.Fatalf("not a coded error: %v", err)
	}
	return custom.Code()
}

type tokenEndpoint struct {
	server *httptest.Server
	forms  []url.Values
	header []http.Header
	query  []url.Values
	status int
	body   string
	mu     sync.Mutex
}

func newTokenEndpoint(t *testing.T, status int, body string) *tokenEndpoint {
	endpoint := &tokenEndpoint{status: status, body: body}
	endpoint.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		form, _ := url.ParseQuery(string(raw))
		endpoint.mu.Lock()
		endpoint.forms = append(endpoint.forms, form)
		endpoint.header = append(endpoint.header, r.Header.Clone())
		endpoint.query = append(endpoint.query, r.URL.Query())
		endpoint.mu.Unlock()
		w.WriteHeader(endpoint.status)
		_, _ = w.Write([]byte(endpoint.body))
	}))
	t.Cleanup(endpoint.server.Close)
	return endpoint
}

func oauthService() *oauth2Webservice {
	return NewOAuth2Webservice(httpClient.HttpClientFactory(logger.LoggerWrapper())).(*oauth2Webservice)
}

func TestOAuth2ExchangeBuildsEachGrant(t *testing.T) {
	endpoint := newTokenEndpoint(t, http.StatusOK, `{"access_token":"a","token_type":"Bearer","expires_in":60,"refresh_token":"r","id_token":"i"}`)
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	timeutils.SetCurrentTimeProvider(func() time.Time { return now })
	defer timeutils.SetCurrentTimeProvider(time.Now)
	config := entity.DefaultOAuth2Config()
	config.AccessTokenURL, config.ClientID, config.ClientSecret, config.Scope = endpoint.server.URL+"/token?v=1", "id", "sec", "api"
	config.Audience, config.Resource = "aud", "res"
	config.TokenRequestParams = []entity.OAuth2Param{{Key: "q", Value: "1", Enabled: true, SendAs: entity.OAuth2SendAsUrl}, {Key: "X-H", Value: "h", Enabled: true, SendAs: entity.OAuth2SendAsHeader}, {Key: "b", Value: "2", Enabled: true, SendAs: entity.OAuth2SendAsBody}, {Key: "off", Value: "x"}}
	config.GrantType = oauth2grant.AuthorizationCodeWithPkce
	token, err := oauthService().Exchange(context.Background(), entity.NewOAuth2GrantRequest(config).WithCode("c", "v"))
	if err != nil || token.AccessToken != "a" || token.RefreshToken != "r" || token.IDToken != "i" || token.Scope != "api" || !token.ExpiresAt.Equal(now.Add(time.Minute)) || !token.ObtainedAt.Equal(now) {
		t.Fatalf("token = %+v, %v", token, err)
	}
	form, header, query := endpoint.forms[0], endpoint.header[0], endpoint.query[0]
	if form.Get("grant_type") != "authorization_code" || form.Get("code") != "c" || form.Get("code_verifier") != "v" || form.Get("redirect_uri") != entity.DefaultOAuth2Callback || form.Has("scope") || form.Get("audience") != "aud" || form.Get("resource") != "res" || form.Get("b") != "2" || form.Has("off") {
		t.Fatalf("form = %v", form)
	}
	if header.Get("Authorization") != "Basic aWQ6c2Vj" || header.Get("X-H") != "h" || header.Get("Content-Type") != "application/x-www-form-urlencoded" || query.Get("q") != "1" || query.Get("v") != "1" {
		t.Fatalf("header = %v query = %v", header, query)
	}
	config.GrantType = oauth2grant.AuthorizationCode
	_, _ = oauthService().Exchange(context.Background(), entity.NewOAuth2GrantRequest(config).WithCode("c", "v"))
	if endpoint.forms[1].Has("code_verifier") {
		t.Fatal("no verifier without PKCE")
	}
	config.GrantType, config.ClientAuthentication, config.Username, config.Password = oauth2grant.PasswordCredentials, clientauthentication.Body, "u", "p"
	config.TokenRequestParams, config.Audience, config.Resource = nil, "", ""
	config.AccessTokenURL = endpoint.server.URL
	_, _ = oauthService().Exchange(context.Background(), entity.NewOAuth2GrantRequest(config))
	password := endpoint.forms[2]
	if password.Get("grant_type") != "password" || password.Get("username") != "u" || password.Get("password") != "p" || password.Get("scope") != "api" || password.Get("client_id") != "id" || password.Get("client_secret") != "sec" || endpoint.header[2].Get("Authorization") != "" || password.Has("audience") {
		t.Fatalf("password = %v", password)
	}
	config.GrantType, config.ClientSecret, config.Scope = oauth2grant.ClientCredentials, "", ""
	_, _ = oauthService().Exchange(context.Background(), entity.NewOAuth2GrantRequest(config))
	credentials := endpoint.forms[3]
	if credentials.Get("grant_type") != "client_credentials" || credentials.Has("scope") || credentials.Has("client_secret") || credentials.Get("client_id") != "id" {
		t.Fatalf("client credentials = %v", credentials)
	}
}

func TestOAuth2RefreshUsesTheRefreshUrlAndParams(t *testing.T) {
	endpoint := newTokenEndpoint(t, http.StatusOK, "access_token=b&token_type=Bearer&expires_in=x&scope=s")
	config := entity.DefaultOAuth2Config()
	config.AccessTokenURL, config.RefreshTokenURL = "http://127.0.0.1:1/unused", endpoint.server.URL+"/refresh?via=r"
	config.RefreshRequestParams = []entity.OAuth2Param{{Key: "tenant", Value: "acme", Enabled: true, SendAs: entity.OAuth2SendAsBody}}
	token, err := oauthService().Refresh(context.Background(), config, entity.OAuth2Token{RefreshToken: "r1"})
	if err != nil || token.AccessToken != "b" || token.ExpiresAt != nil || token.Scope != "s" {
		t.Fatalf("token = %+v, %v", token, err)
	}
	if form := endpoint.forms[0]; form.Get("grant_type") != "refresh_token" || form.Get("refresh_token") != "r1" || form.Get("tenant") != "acme" || endpoint.query[0].Get("via") != "r" {
		t.Fatalf("form = %v", form)
	}
}

func TestOAuth2ErrorsCarryTheProviderError(t *testing.T) {
	endpoint := newTokenEndpoint(t, http.StatusBadRequest, `{"error":"invalid_client","error_description":"who"}`)
	config := entity.DefaultOAuth2Config()
	config.GrantType, config.AccessTokenURL = oauth2grant.ClientCredentials, endpoint.server.URL
	if _, err := oauthService().Exchange(context.Background(), entity.NewOAuth2GrantRequest(config)); codeOf(t, err) != "OAU-04401" {
		t.Fatalf("err = %v", err)
	}
	endpoint.status, endpoint.body = http.StatusInternalServerError, "<html>down</html>"
	_, err := oauthService().Exchange(context.Background(), entity.NewOAuth2GrantRequest(config))
	var custom errs.BaseError
	if !errors.As(err, &custom) || custom.Code() != "OAU-04401" || !strings.Contains(custom.Description(), "http_500") {
		t.Fatalf("err = %v", err)
	}
	config.AccessTokenURL = "http://127.0.0.1:1/token"
	if _, err := oauthService().Exchange(context.Background(), entity.NewOAuth2GrantRequest(config)); codeOf(t, err) != "OAU-11503" {
		t.Fatalf("err = %v", err)
	}
	config.AccessTokenURL = "://bad"
	if _, err := oauthService().Exchange(context.Background(), entity.NewOAuth2GrantRequest(config)); codeOf(t, err) != "OAU-01400" {
		t.Fatalf("err = %v", err)
	}
}

type recordingBrowser struct {
	opened []string
	follow func(string)
	err    error
}

func (r *recordingBrowser) OpenURL(target string) error {
	r.opened = append(r.opened, target)
	if r.follow != nil {
		go r.follow(target)
	}
	return r.err
}

type tcpLoopback struct{}

func (tcpLoopback) Listen(address string) (net.Listener, error) {
	return net.Listen("tcp", address)
}

func freeCallback(t *testing.T) string {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	_ = listener.Close()
	return "http://" + address + "/callback"
}

func TestAuthorizeInBrowserReceivesTheRedirect(t *testing.T) {
	callback := freeCallback(t)
	browser := &recordingBrowser{follow: func(string) {
		for _, path := range []string{"/other", "/callback", "/callback?code=c&state=s&scope=x"} {
			response, err := http.Get(strings.TrimSuffix(callback, "/callback") + path)
			if err == nil {
				_ = response.Body.Close()
			}
		}
	}}
	result, err := NewOAuthBrowserWebservice(browser, tcpLoopback{}).Authorize(context.Background(), "https://sso/authorize", callback)
	if err != nil || result.Code != "c" || result.State != "s" || result.Scope != "x" || len(browser.opened) != 1 {
		t.Fatalf("result = %+v, %v", result, err)
	}
	if listener, err := net.Listen("tcp", strings.TrimSuffix(strings.TrimPrefix(callback, "http://"), "/callback")); err != nil {
		t.Fatal("the listener must be closed")
	} else {
		_ = listener.Close()
	}
}

func TestCallbackHandlerPages(t *testing.T) {
	callbacks := make(chan entity.OAuth2Callback, 1)
	handler := callbackHandler("", callbacks)
	fragment := httptest.NewRecorder()
	handler.ServeHTTP(fragment, httptest.NewRequest(http.MethodGet, "/", nil))
	if !strings.Contains(fragment.Body.String(), "location.hash") || fragment.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("fragment page = %s", fragment.Body.String())
	}
	failed := httptest.NewRecorder()
	handler.ServeHTTP(failed, httptest.NewRequest(http.MethodGet, "/?error=access_denied&error_description=%3Cb%3E", nil))
	if !strings.Contains(failed.Body.String(), "The sign-in failed: access_denied") {
		t.Fatalf("error page = %s", failed.Body.String())
	}
	received := <-callbacks
	if received.Error != "access_denied" || received.ErrorDescription != "<b>" {
		t.Fatalf("callback = %+v", received)
	}
	implicit := httptest.NewRecorder()
	handler.ServeHTTP(implicit, httptest.NewRequest(http.MethodGet, "/?access_token=a&token_type=Bearer&expires_in=60&id_token=i", nil))
	second := httptest.NewRecorder()
	handler.ServeHTTP(second, httptest.NewRequest(http.MethodGet, "/?code=late", nil))
	if got := <-callbacks; got.AccessToken != "a" || got.TokenType != "Bearer" || got.ExpiresIn != "60" || got.IDToken != "i" || !strings.Contains(implicit.Body.String(), "Xenon received the sign-in.") {
		t.Fatalf("implicit = %+v", got)
	}
	missing := httptest.NewRecorder()
	handler.ServeHTTP(missing, httptest.NewRequest(http.MethodGet, "/x?code=c", nil))
	if missing.Code != http.StatusNotFound {
		t.Fatalf("another path = %d", missing.Code)
	}
}

func TestAuthorizeInBrowserFailures(t *testing.T) {
	t.Setenv("XENON_OAUTH_CALLBACK_TIMEOUT_MS", "50")
	callback := freeCallback(t)
	address := strings.TrimSuffix(strings.TrimPrefix(callback, "http://"), "/callback")
	if _, err := NewOAuthBrowserWebservice(&recordingBrowser{}, tcpLoopback{}).Authorize(context.Background(), "https://sso", callback); codeOf(t, err) != "OAU-03408" {
		t.Fatalf("timeout = %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	t.Setenv("XENON_OAUTH_CALLBACK_TIMEOUT_MS", "5000")
	if _, err := NewOAuthBrowserWebservice(&recordingBrowser{}, tcpLoopback{}).Authorize(ctx, "https://sso", callback); codeOf(t, err) != "OAU-08499" {
		t.Fatalf("cancel = %v", err)
	}
	if _, err := NewOAuthBrowserWebservice(&recordingBrowser{err: errors.New("no browser")}, tcpLoopback{}).Authorize(context.Background(), "https://sso", callback); codeOf(t, err) != "OAU-09500" {
		t.Fatalf("browser = %v", err)
	}
	busy, err := net.Listen("tcp", address)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = busy.Close() }()
	browser := &recordingBrowser{}
	if _, err := NewOAuthBrowserWebservice(browser, tcpLoopback{}).Authorize(context.Background(), "https://sso", callback); codeOf(t, err) != "OAU-02409" || len(browser.opened) != 0 {
		t.Fatalf("busy = %v", err)
	}
	if _, err := NewOAuthBrowserWebservice(browser, tcpLoopback{}).Authorize(context.Background(), "https://sso", "http://%zz"); codeOf(t, err) != "OAU-13400" {
		t.Fatalf("bad callback = %v", err)
	}
}

func md5Hex(text string) string {
	sum := md5.Sum([]byte(text))
	return hex.EncodeToString(sum[:])
}

func digestParam(header, key string) string {
	for _, part := range strings.Split(strings.TrimPrefix(header, "Digest "), ", ") {
		name, value, _ := strings.Cut(part, "=")
		if name == key {
			return strings.Trim(value, `"`)
		}
	}
	return ""
}

func TestDigestReadChallengeAndAnswer(t *testing.T) {
	service := NewDigestWebservice()
	state, err := service.ReadChallenge(context.Background(), `Digest realm="r", nonce="n", opaque="o", algorithm=MD5-sess, qop="auth-int, auth"`)
	if err != nil || *state != (entity.DigestState{Realm: "r", Nonce: "n", Opaque: "o", Algorithm: digestalgorithm.MD5Sess, Qop: digestqop.Auth}) {
		t.Fatalf("state = %+v, %v", state, err)
	}
	if state, _ := service.ReadChallenge(context.Background(), `Digest realm="r", nonce="n", algorithm=SHA-1, qop="auth-int"`); state.Algorithm != digestalgorithm.MD5 || state.Qop != digestqop.AuthInt {
		t.Fatalf("fallbacks = %+v", state)
	}
	if state, _ := service.ReadChallenge(context.Background(), `Digest realm="r", nonce="n"`); state.Qop != digestqop.None {
		t.Fatalf("no qop = %+v", state)
	}
	if _, err := service.ReadChallenge(context.Background(), `Basic realm="r"`); codeOf(t, err) != "DIG-01400" {
		t.Fatalf("basic = %v", err)
	}
	request := entity.EffectiveRequest{Method: "GET", URL: "http://h/dir/index.html?x=1"}
	auth := entity.DigestAuth{Username: "Mufasa", Password: "Circle of Life", Algorithm: digestalgorithm.MD5, ClientNonce: "0a4f113b"}
	header, err := service.Answer(context.Background(), request, auth, entity.DigestState{Realm: "testrealm@host.com", Nonce: "dcd98b", Opaque: "5ccc", Qop: digestqop.Auth, NonceCount: 1})
	if err != nil {
		t.Fatal(err)
	}
	ha1, ha2 := md5Hex("Mufasa:testrealm@host.com:Circle of Life"), md5Hex("GET:/dir/index.html?x=1")
	if want := md5Hex(ha1 + ":dcd98b:00000001:0a4f113b:auth:" + ha2); digestParam(header, "response") != want || digestParam(header, "uri") != "/dir/index.html?x=1" || digestParam(header, "algorithm") != "MD5" || digestParam(header, "opaque") != "5ccc" || !strings.HasPrefix(header, "Digest ") {
		t.Fatalf("header = %s", header)
	}
	sess, _ := service.Answer(context.Background(), request, auth, entity.DigestState{Realm: "r", Nonce: "n", Algorithm: digestalgorithm.MD5Sess, Qop: digestqop.Auth, NonceCount: 2})
	sessionKey := md5Hex(md5Hex("Mufasa:r:Circle of Life") + ":n:0a4f113b")
	if want := md5Hex(sessionKey + ":n:00000002:0a4f113b:auth:" + ha2); digestParam(sess, "response") != want || digestParam(sess, "algorithm") != "MD5-sess" {
		t.Fatalf("sess = %s", sess)
	}
	body := entity.EffectiveRequest{Method: "POST", URL: "http://h/p", Body: entity.Body{Mode: bodymode.Raw, Raw: entity.BodyRaw{Language: rawlanguage.Json, Content: "{}"}}}
	integrity, _ := service.Answer(context.Background(), body, auth, entity.DigestState{Realm: "r", Nonce: "n", Qop: digestqop.AuthInt, NonceCount: 1})
	if want := md5Hex(md5Hex("Mufasa:r:Circle of Life") + ":n:00000001:0a4f113b:auth-int:" + md5Hex("POST:/p:"+md5Hex("{}"))); digestParam(integrity, "response") != want {
		t.Fatalf("auth-int = %s", integrity)
	}
	plain, _ := service.Answer(context.Background(), request, entity.DigestAuth{Username: "u", Password: "p", Algorithm: digestalgorithm.SHA256}, entity.DigestState{Realm: "r", Nonce: "n"})
	if digestParam(plain, "algorithm") != "SHA-256" || digestParam(plain, "qop") != "" || len(digestParam(plain, "cnonce")) != 0 || len(digestParam(plain, "response")) != 64 {
		t.Fatalf("no qop = %s", plain)
	}
	generated, _ := service.Answer(context.Background(), request, entity.DigestAuth{Username: "u", Password: "p"}, entity.DigestState{Realm: "r", Nonce: "n", Algorithm: digestalgorithm.SHA512256Sess, Qop: digestqop.Auth})
	if len(digestParam(generated, "cnonce")) != 16 || len(digestParam(generated, "response")) != 64 {
		t.Fatalf("generated = %s", generated)
	}
	if _, err := service.Answer(context.Background(), request, auth, entity.DigestState{Nonce: "n", Algorithm: "SHA-1"}); codeOf(t, err) != "DIG-02400" {
		t.Fatalf("unsupported = %v", err)
	}
	if requestUri("%zz") != "%zz" {
		t.Fatal("an unparsable URL is its own URI")
	}
}

func TestAwsV4SignsHeadersAndQuery(t *testing.T) {
	now := time.Date(2015, 8, 30, 12, 36, 0, 0, time.UTC)
	timeutils.SetCurrentTimeProvider(func() time.Time { return now })
	defer timeutils.SetCurrentTimeProvider(time.Now)
	service := NewAwsV4Webservice()
	auth := entity.AwsV4Auth{AccessKey: "AKIDEXAMPLE", SecretKey: "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY", Region: "us-east-1", Service: "service"}
	request := entity.EffectiveRequest{Method: "GET", URL: "http://127.0.0.1:9/?X-Amz-Signature=stale", Headers: []entity.KeyValue{entity.NewKeyValue("Host", "example.amazonaws.com"), entity.NewKeyValue("X-Amz-Date", "old"), entity.NewKeyValue("Authorization", "old")}}
	signed, err := service.Sign(context.Background(), request, auth)
	if err != nil {
		t.Fatal(err)
	}
	want := "AWS4-HMAC-SHA256 Credential=AKIDEXAMPLE/20150830/us-east-1/service/aws4_request, SignedHeaders=host;x-amz-date, Signature=5fa00fa31553b73ebf1942676e86291e8372ff2a2260956d9b8aae1d763fbf31"
	headers := map[string]string{}
	for _, header := range signed.Headers {
		headers[header.Key] = header.Value
	}
	if headers["Authorization"] != want || headers["X-Amz-Date"] != "20150830T123600Z" || signed.URL != "http://127.0.0.1:9/" || len(signed.Headers) != 3 {
		t.Fatalf("signed = %+v", signed)
	}
	auth.AddAuthDataToQuery, auth.SessionToken = true, "tok"
	presigned, err := service.Sign(context.Background(), entity.EffectiveRequest{Method: "GET", URL: "https://h.test/p?a=b c"}, auth)
	if err != nil {
		t.Fatal(err)
	}
	query, _ := url.Parse(presigned.URL)
	if query.Query().Get("X-Amz-Expires") != "86400" || query.Query().Get("X-Amz-Security-Token") != "tok" || query.Query().Get("a") != "b c" || len(query.Query().Get("X-Amz-Signature")) != 64 || len(presigned.Headers) != 0 {
		t.Fatalf("presigned = %s", presigned.URL)
	}
	file := filepath.Join(t.TempDir(), "payload.txt")
	_ = os.WriteFile(file, []byte("hello"), 0o600)
	s3 := entity.AwsV4Auth{AccessKey: "a", SecretKey: "s", Service: "s3"}
	binary, _ := service.Sign(context.Background(), entity.EffectiveRequest{Method: "PUT", URL: "https://b.test/k", Body: entity.Body{Mode: bodymode.Binary, Binary: entity.FileRef{Src: file}}}, s3)
	missing, _ := service.Sign(context.Background(), entity.EffectiveRequest{Method: "PUT", URL: "https://b.test/k", Body: entity.Body{Mode: bodymode.Binary, Binary: entity.FileRef{Src: file + ".gone"}}}, s3)
	empty, _ := service.Sign(context.Background(), entity.EffectiveRequest{Method: "GET", URL: "https://b.test/k"}, s3)
	hashes := []string{}
	for _, result := range []*entity.EffectiveRequest{binary, missing, empty} {
		for _, header := range result.Headers {
			if header.Key == "X-Amz-Content-Sha256" {
				hashes = append(hashes, header.Value)
			}
		}
	}
	emptyHash := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	if len(hashes) != 3 || hashes[0] != "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824" || hashes[1] != emptyHash || hashes[2] != emptyHash {
		t.Fatalf("hashes = %v", hashes)
	}
	if _, err := service.Sign(context.Background(), entity.EffectiveRequest{Method: "bad method", URL: "https://h"}, s3); codeOf(t, err) != "AWS-01400" {
		t.Fatalf("bad request = %v", err)
	}
}
