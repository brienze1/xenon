package entity

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/brienze1/xenon/internal/domain/enums/authtype"
	"github.com/brienze1/xenon/internal/domain/enums/clientauthentication"
	"github.com/brienze1/xenon/internal/domain/enums/digestalgorithm"
	"github.com/brienze1/xenon/internal/domain/enums/digestqop"
	"github.com/brienze1/xenon/internal/domain/enums/oauth2grant"
	"github.com/brienze1/xenon/internal/domain/enums/tokenplacement"
)

func upper(text string) string { return strings.ToUpper(text) }

func TestApiKeyAuth(t *testing.T) {
	if DefaultApiKeyAuth().In != ApiKeyInHeader || !DefaultApiKeyAuth().IsBlank() || !DefaultApiKeyAuth().SendsNothing() {
		t.Fatal("the default API key is a blank header key")
	}
	query := ApiKeyAuth{In: ApiKeyInQuery}
	if !query.IsInQuery() || query.IsBlank() || !query.SendsNothing() {
		t.Fatal("an empty query key sends nothing but is not blank")
	}
	for _, apiKey := range []ApiKeyAuth{{Key: "k"}, {Value: "v"}} {
		if apiKey.SendsNothing() || apiKey.IsBlank() {
			t.Errorf("%+v sends something", apiKey)
		}
	}
	mapped := ApiKeyAuth{Key: "k", Value: "v", In: ApiKeyInHeader}.MapText(upper)
	if mapped != (ApiKeyAuth{Key: "K", Value: "V", In: ApiKeyInHeader}) {
		t.Fatalf("MapText = %+v", mapped)
	}
	fields := mapped.TextFields("auth")
	if len(fields) != 2 || fields[0].Name != "auth.apikey.key" || fields[1].Name != "auth.apikey.value" || fields[1].Text != "V" {
		t.Fatalf("TextFields = %+v", fields)
	}
	if log := mapped.LogValue().(map[string]any); log["hasValue"] != true || log["key"] != "K" || log["value"] != nil {
		t.Fatalf("LogValue = %v", log)
	}
}

func TestOAuth2ParamPlacement(t *testing.T) {
	cases := map[string][3]bool{OAuth2SendAsUrl: {true, false, false}, OAuth2SendAsHeader: {false, true, false}, OAuth2SendAsBody: {false, false, true}, "": {false, false, true}}
	for sendAs, want := range cases {
		param := OAuth2Param{SendAs: sendAs}
		if got := [3]bool{param.SendsInUrl(), param.SendsInHeader(), param.SendsInBody()}; got != want {
			t.Errorf("%q = %v, want %v", sendAs, got, want)
		}
	}
	if !(OAuth2Param{Key: "a", Enabled: true}).IsActive() || (OAuth2Param{Key: " ", Enabled: true}).IsActive() || (OAuth2Param{Key: "a"}).IsActive() {
		t.Fatal("an active param is enabled with a key")
	}
}

func validCodeConfig() OAuth2Config {
	config := DefaultOAuth2Config()
	config.AuthURL = "https://sso.test/authorize"
	config.AccessTokenURL = "https://sso.test/token"
	return config
}

func TestOAuth2ConfigDefaultsAndGrants(t *testing.T) {
	config := DefaultOAuth2Config()
	if config.GrantType != oauth2grant.AuthorizationCode || config.RedirectURI != DefaultOAuth2Callback || config.HeaderPrefix != "Bearer" || !config.AutoRefresh || config.UseToken != UseAccessToken || config.ChallengeAlgorithm != PkceS256 || config.ClientAuthentication != clientauthentication.Header || config.AddTokenTo != tokenplacement.Header {
		t.Fatalf("defaults = %+v", config)
	}
	if !config.IsBlank() {
		t.Fatal("the default configuration is blank")
	}
	grants := map[string][5]bool{
		"authorization_code":           {true, true, false, false, false},
		"authorization_code_with_pkce": {true, true, true, false, false},
		"implicit":                     {true, false, false, true, true},
		"password_credentials":         {false, false, false, false, true},
		"client_credentials":           {false, false, false, false, false},
	}
	for grant, want := range grants {
		config.GrantType = oauth2grant.GetOAuth2Grant(grant)
		if got := [5]bool{config.UsesBrowser(), config.UsesAuthorizationCode(), config.UsesPkce(), config.IsImplicit(), config.IsLegacyGrant()}; got != want {
			t.Errorf("%s = %v, want %v", grant, got, want)
		}
	}
	config.GrantType = oauth2grant.PasswordCredentials
	if !config.IsPasswordGrant() {
		t.Fatal("password grant")
	}
}

func TestOAuth2ConfigPlacementAndHeaders(t *testing.T) {
	config := DefaultOAuth2Config()
	if config.UsesPlainChallenge() || config.SendsClientInBody() || config.AddsTokenToQuery() {
		t.Fatal("defaults are S256, header client auth and header token")
	}
	config.ChallengeAlgorithm, config.ClientAuthentication, config.AddTokenTo = PkcePlain, clientauthentication.Body, tokenplacement.QueryParams
	if !config.UsesPlainChallenge() || !config.SendsClientInBody() || !config.AddsTokenToQuery() {
		t.Fatal("plain, body and query")
	}
	if got := config.AuthorizationHeader("t"); got != "Bearer t" {
		t.Fatalf("AuthorizationHeader = %q", got)
	}
	config.HeaderPrefix = ""
	if got := config.AuthorizationHeader("t"); got != "t" {
		t.Fatalf("no prefix = %q", got)
	}
	config.AccessTokenURL = "https://a/token"
	if config.RefreshURL() != "https://a/token" {
		t.Fatal("the refresh URL falls back to the access token URL")
	}
	config.RefreshTokenURL = "https://a/refresh"
	if config.RefreshURL() != "https://a/refresh" {
		t.Fatal("the refresh URL wins")
	}
	for raw, want := range map[string]bool{"https://oauth.pstmn.io/v1/callback": true, "https://OAUTH.PSTMN.IO/v1/browser-callback": true, DefaultOAuth2Callback: false, "::bad": false} {
		config.RedirectURI = raw
		if config.UsesPostmanCallback() != want {
			t.Errorf("UsesPostmanCallback(%q) != %v", raw, want)
		}
	}
}

func TestOAuth2ConfigInvalidReason(t *testing.T) {
	if reason := validCodeConfig().InvalidReason(); reason != "" {
		t.Fatalf("valid config: %q", reason)
	}
	cases := map[string]func(*OAuth2Config){
		"auth URL is required":         func(c *OAuth2Config) { c.AuthURL = " " },
		"access token URL is required": func(c *OAuth2Config) { c.AccessTokenURL = "" },
		"callback URL must be":         func(c *OAuth2Config) { c.RedirectURI = "https://127.0.0.1:1/cb" },
		"auth URL must use https":      func(c *OAuth2Config) { c.AuthURL = "http://sso.test/authorize" },
	}
	for want, change := range cases {
		config := validCodeConfig()
		change(&config)
		if reason := config.InvalidReason(); !strings.Contains(reason, want) {
			t.Errorf("%s: got %q", want, reason)
		}
	}
	implicit := validCodeConfig()
	implicit.GrantType, implicit.AccessTokenURL = oauth2grant.Implicit, ""
	if reason := implicit.InvalidReason(); reason != "" {
		t.Fatalf("implicit needs no access token URL: %q", reason)
	}
	credentials := DefaultOAuth2Config()
	credentials.GrantType, credentials.AccessTokenURL, credentials.RedirectURI = oauth2grant.ClientCredentials, "https://t", "nope"
	if reason := credentials.InvalidReason(); reason != "" {
		t.Fatalf("client credentials need no browser: %q", reason)
	}
}

func TestLoopbackCallbacksAndSafeBrowserUrls(t *testing.T) {
	callbacks := map[string]bool{
		"http://127.0.0.1:53682/callback": true, "HTTP://localhost:1/cb": true, "http://[::1]:8080/": true,
		"http://127.0.0.1/cb": false, "https://127.0.0.1:1/cb": false, "http://127.0.0.2:1/cb": false, "http://evil.test:1/cb": false, "%zz": false,
	}
	for raw, want := range callbacks {
		if isLoopbackCallback(raw) != want {
			t.Errorf("isLoopbackCallback(%q) != %v", raw, want)
		}
	}
	urls := map[string]bool{
		"https://sso.test/a": true, "HTTPS://sso.test": true, "http://localhost:9/a": true, "http://127.0.0.1/a": true,
		"http://sso.test/a": false, "javascript:alert(1)": false, "file:///etc/passwd": false, "https:///nohost": false, "%zz": false,
	}
	for raw, want := range urls {
		if isSafeBrowserUrl(raw) != want {
			t.Errorf("isSafeBrowserUrl(%q) != %v", raw, want)
		}
	}
}

func TestOAuth2ConfigMapTextAndFields(t *testing.T) {
	config := DefaultOAuth2Config()
	config.ClientSecret, config.Scope = "{{s}}", "a"
	config.TokenRequestParams = []OAuth2Param{{Key: "k", Value: "v", Enabled: true, SendAs: OAuth2SendAsUrl}}
	mapped := config.MapText(upper)
	if mapped.ClientSecret != "{{S}}" || mapped.Scope != "A" || mapped.RedirectURI != upper(DefaultOAuth2Callback) || mapped.TokenRequestParams[0] != (OAuth2Param{Key: "K", Value: "V", Enabled: true, SendAs: OAuth2SendAsUrl}) || len(mapped.AuthRequestParams) != 0 {
		t.Fatalf("MapText = %+v", mapped)
	}
	if mapped.GrantType != config.GrantType || mapParams(nil, upper) != nil {
		t.Fatal("enums are not text and nil params stay nil")
	}
	fields := config.TextFields("auth")
	if len(fields) != 8 || fields[4].Name != "auth.oauth2.clientSecret" || fields[4].Text != "{{s}}" {
		t.Fatalf("TextFields = %+v", fields)
	}
	if log := config.LogValue().(map[string]any); log["hasClientSecret"] != true || log["clientSecret"] != nil {
		t.Fatalf("LogValue = %v", log)
	}
}

func TestOAuth2TokenExpiryAndParts(t *testing.T) {
	now := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	expiresAt := now.Add(31 * time.Second)
	token := OAuth2Token{AccessToken: "a", RefreshToken: "r", IDToken: "i", ExpiresAt: &expiresAt}
	if token.IsExpired(now) || !token.IsExpired(now.Add(time.Second)) || !token.IsExpired(now.Add(2*time.Second)) {
		t.Fatal("a token expires 30 seconds early")
	}
	if (OAuth2Token{}).IsExpired(now) {
		t.Fatal("a token without expiry never expires")
	}
	if token.Value() != "a" || token.Using(UseIDToken).Value() != "i" {
		t.Fatal("Value follows UseToken")
	}
	if token.Part("access") != "a" || token.Part("refresh") != "r" || token.Part("id") != "i" {
		t.Fatal("Part")
	}
	if !reflect.DeepEqual(token.Secrets(), []string{"a", "r", "i"}) || !reflect.DeepEqual((OAuth2Token{AccessToken: "a"}).Secrets(), []string{"a"}) {
		t.Fatal("Secrets lists the values present")
	}
	var missing *OAuth2Token
	if len(missing.SecretsOrNone()) != 0 || len(token.SecretsOrNone()) != 3 {
		t.Fatal("SecretsOrNone")
	}
	if !token.CanRefresh() || !token.HasIDToken() || (OAuth2Token{}).CanRefresh() || (OAuth2Token{}).HasIDToken() {
		t.Fatal("CanRefresh / HasIDToken")
	}
	if !(OAuth2Token{TokenType: "MAC"}).IsMac() || (OAuth2Token{TokenType: "Bearer"}).IsMac() {
		t.Fatal("IsMac")
	}
	owned := token.OwnedBy("o", UseIDToken)
	if owned.Owner != "o" || owned.UseToken != UseIDToken {
		t.Fatal("OwnedBy")
	}
	refreshed := owned.Refreshed(OAuth2Token{AccessToken: "b"})
	if refreshed.AccessToken != "b" || refreshed.RefreshToken != "r" || refreshed.IDToken != "i" || refreshed.Owner != "o" || refreshed.UseToken != UseIDToken {
		t.Fatalf("Refreshed keeps what was not rotated: %+v", refreshed)
	}
	owned.Scope = "api"
	rotated := owned.Refreshed(OAuth2Token{AccessToken: "b", RefreshToken: "r2", IDToken: "i2", Scope: "new"})
	if rotated.RefreshToken != "r2" || rotated.IDToken != "i2" || rotated.Scope != "new" || owned.Refreshed(OAuth2Token{}).Scope != "api" {
		t.Fatalf("Refreshed takes rotated values: %+v", rotated)
	}
	if log := token.LogValue().(map[string]any); log["hasRefreshToken"] != true || log["accessToken"] != nil {
		t.Fatalf("LogValue = %v", log)
	}
}

func TestOAuth2CallbackToken(t *testing.T) {
	now := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	callback := OAuth2Callback{AccessToken: "a", TokenType: "Bearer", ExpiresIn: "60", State: "s"}
	token := callback.Token(now, "api")
	if token.AccessToken != "a" || token.Scope != "api" || !token.ExpiresAt.Equal(now.Add(time.Minute)) || !token.ObtainedAt.Equal(now) {
		t.Fatalf("Token = %+v", token)
	}
	callback.ExpiresIn, callback.Scope = "soon", "granted"
	if token := callback.Token(now, "api"); token.ExpiresAt != nil || token.Scope != "granted" {
		t.Fatalf("Token = %+v", token)
	}
	if !callback.MatchesState("s") || callback.MatchesState("x") || callback.IsError() || !(OAuth2Callback{Error: "e"}).IsError() {
		t.Fatal("state and error predicates")
	}
	if log := callback.LogValue().(map[string]any); log["hasAccessToken"] != true || log["accessToken"] != nil {
		t.Fatalf("LogValue = %v", log)
	}
}

func TestOAuth2GrantRequest(t *testing.T) {
	config := DefaultOAuth2Config()
	grant := NewOAuth2GrantRequest(config).WithCode("c", "v")
	if grant.RedirectURI != DefaultOAuth2Callback || grant.Code != "c" || grant.Verifier != "v" {
		t.Fatalf("grant = %+v", grant)
	}
	if log := grant.LogValue().(map[string]any); log["hasCode"] != true || log["hasVerifier"] != true || log["code"] != nil {
		t.Fatalf("LogValue = %v", log)
	}
}

func TestAwsV4Auth(t *testing.T) {
	auth := AwsV4Auth{}
	if auth.RegionOrDefault() != "us-east-1" || auth.ServiceOrDefault() != "execute-api" || auth.IsS3() || !auth.IsBlank() {
		t.Fatal("defaults")
	}
	auth = AwsV4Auth{Region: " eu-west-1 ", Service: " S3 ", SecretKey: "x"}
	if auth.RegionOrDefault() != "eu-west-1" || auth.ServiceOrDefault() != "S3" || !auth.IsS3() || auth.IsBlank() {
		t.Fatal("values")
	}
	mapped := AwsV4Auth{AccessKey: "a", SecretKey: "s", SessionToken: "t", Region: "r", Service: "v"}.MapText(upper)
	if mapped != (AwsV4Auth{AccessKey: "A", SecretKey: "S", SessionToken: "T", Region: "R", Service: "V"}) {
		t.Fatalf("MapText = %+v", mapped)
	}
	if fields := mapped.TextFields("auth"); len(fields) != 3 || fields[1].Name != "auth.awsv4.secretKey" {
		t.Fatalf("TextFields = %+v", fields)
	}
	if log := mapped.LogValue().(map[string]any); log["hasSecretKey"] != true || log["secretKey"] != nil || log["hasSessionToken"] != true {
		t.Fatalf("LogValue = %v", log)
	}
}

func TestDigestAuthAndState(t *testing.T) {
	digest := DefaultDigestAuth()
	if digest.Algorithm != digestalgorithm.MD5 || digest.Qop != digestqop.None || !digest.IsBlank() || !digest.RetriesChallenge() {
		t.Fatalf("defaults = %+v", digest)
	}
	if _, found := digest.TypedState(); found {
		t.Fatal("no nonce, no typed state")
	}
	digest.Nonce, digest.Realm, digest.Opaque, digest.NonceCount, digest.Qop = "n", "r", "o", "0000000a", digestqop.Auth
	state, found := digest.TypedState()
	if !found || state != (DigestState{Realm: "r", Nonce: "n", Opaque: "o", Algorithm: digestalgorithm.MD5, Qop: digestqop.Auth, NonceCount: 10}) {
		t.Fatalf("TypedState = %+v", state)
	}
	digest.NonceCount = "zz"
	if state, _ := digest.TypedState(); state.NonceCount != 0 {
		t.Fatal("an unreadable nonce count starts at zero")
	}
	if digest.IsBlank() || (DigestAuth{DisableRetryRequest: true}).RetriesChallenge() {
		t.Fatal("IsBlank / RetriesChallenge")
	}
	for _, blank := range []DigestAuth{{Username: "u"}, {Password: "p"}, {Realm: "r"}, {Opaque: "o"}} {
		if blank.IsBlank() {
			t.Errorf("%+v is not blank", blank)
		}
	}
	mapped := DigestAuth{Username: "u", Password: "p", Realm: "r", Nonce: "n", Opaque: "o", NonceCount: "c", ClientNonce: "k"}.MapText(upper)
	if mapped != (DigestAuth{Username: "U", Password: "P", Realm: "R", Nonce: "N", Opaque: "O", NonceCount: "C", ClientNonce: "K"}) {
		t.Fatalf("MapText = %+v", mapped)
	}
	if fields := mapped.TextFields("auth"); len(fields) != 2 || fields[1].Name != "auth.digest.password" {
		t.Fatalf("TextFields = %+v", fields)
	}
	if log := mapped.LogValue().(map[string]any); log["hasPassword"] != true || log["password"] != nil {
		t.Fatalf("LogValue = %v", log)
	}
	next := DigestState{Nonce: "n", NonceCount: 1}.Next()
	if next.NonceCount != 2 || next.IsEmpty() || !(DigestState{}).IsEmpty() {
		t.Fatal("Next / IsEmpty")
	}
	if log := next.LogValue().(map[string]any); log["nonceCount"] != 2 || log["nonce"] != nil {
		t.Fatalf("LogValue = %v", log)
	}
}

func TestCookieClosedTabsAndAcknowledgement(t *testing.T) {
	expires := time.Now()
	if !(Cookie{}).IsSession() || (Cookie{Expires: &expires}).IsSession() {
		t.Fatal("a cookie without expiry is a session cookie")
	}
	if log := (Cookie{Name: "n", Value: "secret"}).LogValue().(map[string]any); log["name"] != "n" || log["value"] != nil {
		t.Fatalf("LogValue = %v", log)
	}
	tabs := make([]ClosedTab, 21)
	for index := range tabs {
		tabs[index].Tab.TabID = string(rune('a' + index))
	}
	if latest := LatestClosedTabs(tabs); len(latest) != 20 || latest[0].Tab.TabID != "b" {
		t.Fatalf("LatestClosedTabs keeps the newest 20, got %d starting at %s", len(latest), latest[0].Tab.TabID)
	}
	if latest := LatestClosedTabs(tabs[:20]); len(latest) != 20 {
		t.Fatal("20 tabs stay")
	}
	if !(SendAcknowledgement{ExpiredToken: true}).AcceptsExpiredToken() || (SendAcknowledgement{}).AcceptsExpiredToken() {
		t.Fatal("AcceptsExpiredToken")
	}
}

func TestAuthTypePredicatesAndSubObjects(t *testing.T) {
	predicates := map[string][4]bool{"apikey": {true, false, false, false}, "oauth2": {false, true, false, false}, "awsv4": {false, false, true, false}, "digest": {false, false, false, true}, "basic": {}}
	for raw, want := range predicates {
		auth := Auth{Type: authtype.GetAuthType(raw)}
		if got := [4]bool{auth.IsApiKey(), auth.IsOAuth2(), auth.IsAwsV4(), auth.IsDigest()}; got != want {
			t.Errorf("%s = %v, want %v", raw, got, want)
		}
	}
	if (Auth{Type: authtype.ApiKey}).HasApiKey() || !(Auth{Type: authtype.ApiKey, ApiKey: DefaultApiKeyAuth()}).HasApiKey() || !(Auth{ApiKey: ApiKeyAuth{Key: "k"}}).HasApiKey() {
		t.Fatal("HasApiKey: given sub-object for the type, or any content")
	}
	if (Auth{Type: authtype.OAuth2}).HasOAuth2() || !(Auth{Type: authtype.OAuth2, OAuth2: DefaultOAuth2Config()}).HasOAuth2() || !(Auth{OAuth2: OAuth2Config{ClientID: "c"}}).HasOAuth2() || (Auth{OAuth2: DefaultOAuth2Config()}).HasOAuth2() {
		t.Fatal("HasOAuth2")
	}
	if (Auth{Type: authtype.AwsV4}).HasAwsV4() || !(Auth{AwsV4: AwsV4Auth{Region: "r"}}).HasAwsV4() {
		t.Fatal("HasAwsV4")
	}
	if (Auth{Type: authtype.Digest}).HasDigest() || !(Auth{Type: authtype.Digest, Digest: DefaultDigestAuth()}).HasDigest() || !(Auth{Digest: DigestAuth{Username: "u"}}).HasDigest() || (Auth{Digest: DefaultDigestAuth()}).HasDigest() {
		t.Fatal("HasDigest")
	}
	auth := Auth{Type: authtype.ApiKey, ApiKey: ApiKeyAuth{Key: "k"}, OAuth2: OAuth2Config{Scope: "s"}, AwsV4: AwsV4Auth{Region: "r"}, Digest: DigestAuth{Username: "u"}}.MapText(upper)
	if auth.ApiKey.Key != "K" || auth.OAuth2.Scope != "S" || auth.AwsV4.Region != "R" || auth.Digest.Username != "U" {
		t.Fatalf("MapText walks every sub-object: %+v", auth)
	}
	for raw, field := range map[string]string{"apikey": "auth.apikey.key", "oauth2": "auth.oauth2.authUrl", "awsv4": "auth.awsv4.accessKey", "digest": "auth.digest.username"} {
		if fields := (Auth{Type: authtype.GetAuthType(raw)}).TextFields("auth"); len(fields) == 0 || fields[0].Name != field {
			t.Errorf("%s TextFields = %+v", raw, fields)
		}
	}
	if log := auth.LogValue().(map[string]any); log["apikey"] == nil || log["oauth2"] == nil || log["awsv4"] == nil || log["digest"] == nil {
		t.Fatalf("LogValue = %v", log)
	}
}

func TestAttributedAuthMapsPostmanKeys(t *testing.T) {
	apiKey := NewAttributedAuth(authtype.ApiKey, []AuthAttribute{{Key: "key", Value: "X"}, {Key: "value", Value: "v"}, {Key: "in", Value: "QUERY"}, {Key: "extra", Value: 1}})
	if apiKey.ApiKey != (ApiKeyAuth{Key: "X", Value: "v", In: ApiKeyInQuery}) || !reflect.DeepEqual(apiKey.Attributes, []AuthAttribute{{Key: "extra", Value: 1}}) {
		t.Fatalf("apikey = %+v", apiKey)
	}
	if NewAttributedAuth(authtype.ApiKey, []AuthAttribute{{Key: "in", Value: "cookie"}}).ApiKey.In != ApiKeyInHeader {
		t.Fatal("unknown locations fall back to header")
	}
	params := []any{map[string]any{"key": "a", "value": "1", "send_as": "request_header", "enabled": false}, map[string]any{"key": "b", "value": 2}, "junk"}
	oauth := NewAttributedAuth(authtype.OAuth2, []AuthAttribute{
		{Key: "grant_type", Value: "client_credentials"}, {Key: "tokenName", Value: "t"}, {Key: "redirect_uri", Value: "r"}, {Key: "authUrl", Value: "a"},
		{Key: "accessTokenUrl", Value: "at"}, {Key: "refreshTokenUrl", Value: "rt"}, {Key: "clientId", Value: "ci"}, {Key: "clientSecret", Value: "cs"},
		{Key: "username", Value: "u"}, {Key: "password", Value: "p"}, {Key: "scope", Value: "s"}, {Key: "state", Value: "st"},
		{Key: "challengeAlgorithm", Value: "plain"}, {Key: "code_verifier", Value: "cv"}, {Key: "client_authentication", Value: "BODY"},
		{Key: "addTokenTo", Value: "queryParams"}, {Key: "headerPrefix", Value: "Token"}, {Key: "useToken", Value: "id_token"},
		{Key: "autoRefresh", Value: "false"}, {Key: "authRequestParams", Value: params}, {Key: "tokenRequestParams", Value: params},
		{Key: "refreshRequestParams", Value: "not a list"}, {Key: "audience", Value: 7}, {Key: "resource", Value: nil}, {Key: "useBrowser", Value: false},
	})
	config := oauth.OAuth2
	if config.GrantType != oauth2grant.ClientCredentials || config.TokenName != "t" || config.RedirectURI != "r" || config.AuthURL != "a" || config.AccessTokenURL != "at" || config.RefreshTokenURL != "rt" || config.ClientID != "ci" || config.ClientSecret != "cs" || config.Username != "u" || config.Password != "p" || config.Scope != "s" || config.State != "st" || config.ChallengeAlgorithm != "plain" || config.CodeVerifier != "cv" || config.ClientAuthentication != clientauthentication.Body || config.AddTokenTo != tokenplacement.QueryParams || config.HeaderPrefix != "Token" || config.UseToken != UseIDToken || config.AutoRefresh || config.Audience != "7" || config.Resource != "" {
		t.Fatalf("oauth2 = %+v", config)
	}
	wantAuthParams := []OAuth2Param{{Key: "a", Value: "1", SendAs: OAuth2SendAsHeader}, {Key: "b", Value: "2", SendAs: OAuth2SendAsUrl, Enabled: true}}
	if !reflect.DeepEqual(config.AuthRequestParams, wantAuthParams) || config.TokenRequestParams[1].SendAs != OAuth2SendAsBody || len(config.RefreshRequestParams) != 0 {
		t.Fatalf("params = %+v / %+v / %+v", config.AuthRequestParams, config.TokenRequestParams, config.RefreshRequestParams)
	}
	if !reflect.DeepEqual(oauth.Attributes, []AuthAttribute{{Key: "useBrowser", Value: false}}) {
		t.Fatalf("unknown oauth2 keys stay: %+v", oauth.Attributes)
	}
	defaults := NewAttributedAuth(authtype.OAuth2, []AuthAttribute{{Key: "client_authentication", Value: "x"}, {Key: "addTokenTo", Value: "x"}, {Key: "autoRefresh", Value: true}})
	if defaults.OAuth2.ClientAuthentication != clientauthentication.Header || defaults.OAuth2.AddTokenTo != tokenplacement.Header || !defaults.OAuth2.AutoRefresh || defaults.OAuth2.GrantType != oauth2grant.AuthorizationCode {
		t.Fatalf("defaults = %+v", defaults.OAuth2)
	}
	aws := NewAttributedAuth(authtype.AwsV4, []AuthAttribute{{Key: "accessKey", Value: "a"}, {Key: "secretKey", Value: "s"}, {Key: "sessionToken", Value: "t"}, {Key: "region", Value: "r"}, {Key: "service", Value: "v"}, {Key: "addAuthDataToQuery", Value: true}})
	if aws.AwsV4 != (AwsV4Auth{AccessKey: "a", SecretKey: "s", SessionToken: "t", Region: "r", Service: "v", AddAuthDataToQuery: true}) {
		t.Fatalf("awsv4 = %+v", aws.AwsV4)
	}
	digest := NewAttributedAuth(authtype.Digest, []AuthAttribute{{Key: "username", Value: "u"}, {Key: "password", Value: "p"}, {Key: "algorithm", Value: "SHA-256-sess"}, {Key: "qop", Value: "auth-int"}, {Key: "realm", Value: "r"}, {Key: "nonce", Value: "n"}, {Key: "opaque", Value: "o"}, {Key: "nonceCount", Value: "1"}, {Key: "clientNonce", Value: "c"}, {Key: "disableRetryRequest", Value: "TRUE"}})
	if digest.Digest != (DigestAuth{Username: "u", Password: "p", Algorithm: digestalgorithm.SHA256Sess, Qop: digestqop.AuthInt, Realm: "r", Nonce: "n", Opaque: "o", NonceCount: "1", ClientNonce: "c", DisableRetryRequest: true}) {
		t.Fatalf("digest = %+v", digest.Digest)
	}
	fallback := NewAttributedAuth(authtype.Digest, []AuthAttribute{{Key: "algorithm", Value: "SHA-1"}, {Key: "qop", Value: "weird"}})
	if fallback.Digest.Algorithm != digestalgorithm.MD5 || fallback.Digest.Qop != digestqop.None {
		t.Fatalf("fallback = %+v", fallback.Digest)
	}
}

func TestLiteralAccessTokensMoveOutOfTheAuth(t *testing.T) {
	literal := NewAttributedAuth(authtype.OAuth2, []AuthAttribute{{Key: "accessToken", Value: "t0ken"}, {Key: "tokenType", Value: "Bearer"}, {Key: "useBrowser", Value: true}})
	stripped, token := literal.WithoutLiteralAccessToken()
	if token == nil || token.AccessToken != "t0ken" || token.TokenType != "Bearer" || token.UseToken != UseAccessToken {
		t.Fatalf("token = %+v", token)
	}
	if !reflect.DeepEqual(stripped.Attributes, []AuthAttribute{{Key: "useBrowser", Value: true}}) || len(literal.Attributes) != 3 {
		t.Fatalf("stripped = %+v, original = %+v", stripped.Attributes, literal.Attributes)
	}
	for _, kept := range []Auth{
		NewAttributedAuth(authtype.OAuth2, []AuthAttribute{{Key: "accessToken", Value: "{{token}}"}}),
		NewAttributedAuth(authtype.OAuth2, []AuthAttribute{{Key: "accessToken", Value: ""}}),
		NewAttributedAuth(authtype.Bearer, []AuthAttribute{{Key: "accessToken", Value: "t0ken"}}),
	} {
		if after, token := kept.WithoutLiteralAccessToken(); token != nil || !reflect.DeepEqual(after, kept) {
			t.Errorf("%+v must stay unchanged", kept.Attributes)
		}
	}
	request := NewRequest("r", "R")
	request.Auth = literal
	folder := NewFolder("f", "F")
	folder.Auth = literal
	collection := NewCollection("c", "C")
	collection.Auth = literal
	write := ImportWrite{Nodes: []ImportNode{NewContainerImportNode(folder, []ImportNode{NewRequestImportNode(request)})}, Collection: &collection}
	cleaned, tokens := write.WithoutLiteralTokens()
	owners := []string{}
	for _, owned := range tokens {
		owners = append(owners, owned.Owner)
	}
	if !reflect.DeepEqual(owners, []string{"f", "r", "c"}) {
		t.Fatalf("owners = %v", owners)
	}
	if cleaned.Collection.Auth.attributeText(attributeAccessToken) != "" || cleaned.Nodes[0].Container.Auth.attributeText(attributeAccessToken) != "" || cleaned.Nodes[0].Children[0].Request.Auth.attributeText(attributeAccessToken) != "" {
		t.Fatal("every literal token left the write")
	}
	if write.Nodes[0].Children[0].Request.Auth.attributeText(attributeAccessToken) != "t0ken" {
		t.Fatal("the original write is not changed")
	}
	if _, none := (ImportWrite{}).WithoutLiteralTokens(); len(none) != 0 {
		t.Fatal("no tokens")
	}
}

func TestEffectiveRequestAuthEdits(t *testing.T) {
	request := EffectiveRequest{
		URL:         "https://h/p?a=1&api_key=old&b&c%20d=2#f",
		Headers:     []KeyValue{NewKeyValue("x-key", "old"), NewKeyValue("Other", "o")},
		AutoHeaders: []AutoHeader{NewAutoHeader("X-Key", "auto"), NewAutoHeader("Accept", "*/*")},
	}
	edited := request.WithHeader("X-Key", "new")
	if !reflect.DeepEqual(edited.Headers, []KeyValue{NewKeyValue("Other", "o"), NewKeyValue("X-Key", "new")}) || len(edited.AutoHeaders) != 1 || edited.AutoHeaders[0].Key != "Accept" {
		t.Fatalf("WithHeader = %+v / %+v", edited.Headers, edited.AutoHeaders)
	}
	if len(request.Headers) != 2 || len(request.AutoHeaders) != 2 {
		t.Fatal("the original is not changed")
	}
	if got := request.WithQueryParam("api_key", "k y").URL; got != "https://h/p?a=1&b&c%20d=2&api_key=k+y#f" {
		t.Fatalf("WithQueryParam = %q", got)
	}
	if got := request.WithoutQueryParam("c d").URL; got != "https://h/p?a=1&api_key=old&b#f" {
		t.Fatalf("WithoutQueryParam = %q", got)
	}
	if got := (EffectiveRequest{URL: "https://h/p?x=1"}).WithoutQueryParam("x").URL; got != "https://h/p" {
		t.Fatalf("removing the last param drops the query: %q", got)
	}
	if got := (EffectiveRequest{URL: "https://h/p"}).WithQueryParam("access_token", "t").URL; got != "https://h/p?access_token=t" {
		t.Fatalf("WithQueryParam on no query = %q", got)
	}
	if got := (EffectiveRequest{URL: "https://h/p?%zz=1"}).WithoutQueryParam("%zz").URL; got != "https://h/p" {
		t.Fatalf("an undecodable key is compared raw: %q", got)
	}
	if request.WithURL("u").URL != "u" {
		t.Fatal("WithURL")
	}
	if request.WithAuthOwner("tab").AuthOwner != "tab" || (EffectiveRequest{AuthOwner: "r"}).WithAuthOwner("tab").AuthOwner != "r" {
		t.Fatal("WithAuthOwner only fills an empty owner")
	}
}

func TestResponseDigestChallenge(t *testing.T) {
	challenge := NewKeyValue("WWW-Authenticate", ` Digest realm="r"`)
	for status, want := range map[int]bool{401: true, 403: true, 200: false, 407: false} {
		response := Response{StatusCode: status, Headers: []KeyValue{NewKeyValue("WWW-Authenticate", "Basic realm=x"), challenge}}
		if value, found := response.DigestChallenge(); found != want || (found && value != challenge.Value) {
			t.Errorf("%d: %q %v", status, value, found)
		}
	}
	if _, found := (Response{StatusCode: 401, Headers: []KeyValue{NewKeyValue("WWW-Authenticate", "Basic realm=x")}}).DigestChallenge(); found {
		t.Fatal("a Basic challenge is not a Digest challenge")
	}
}

func TestVaultTokens(t *testing.T) {
	vault := Vault{State: "unlocked"}.WithToken(OAuth2Token{Owner: "o", AccessToken: "a"})
	if token, found := vault.TokenOf("o"); !found || token.AccessToken != "a" {
		t.Fatal("TokenOf after WithToken")
	}
	if _, found := vault.WithoutToken("o").TokenOf("o"); found {
		t.Fatal("WithoutToken")
	}
	if _, found := vault.TokenOf("o"); !found {
		t.Fatal("WithoutToken does not change the original")
	}
	locked := vault
	locked.State = "locked"
	if _, found := locked.TokenOf("o"); found {
		t.Fatal("a locked vault hands out no token")
	}
	stored := NewVaultFromStorage(VaultStorage{KeyFound: true, ContentFound: true, OAuth2Tokens: map[string]OAuth2Token{"o": {AccessToken: "a"}}})
	if token, found := stored.TokenOf("o"); !found || token.AccessToken != "a" {
		t.Fatal("NewVaultFromStorage keeps the tokens")
	}
}

func TestDigestKnownStatePrefersTheStoredChallenge(t *testing.T) {
	stored := &DigestState{Realm: "stored", Nonce: "s"}
	typed := DigestAuth{Realm: "typed", Nonce: "t", NonceCount: "0a"}
	if got := typed.KnownState(stored); got != stored {
		t.Fatalf("stored state = %+v", got)
	}
	if got := typed.KnownState(nil); got == nil || got.Realm != "typed" || got.NonceCount != 10 {
		t.Fatalf("typed state = %+v", got)
	}
	if got := (DigestAuth{}).KnownState(nil); got != nil {
		t.Fatalf("no state = %+v", got)
	}
}

func TestVaultWithTokensKeepsEveryOwner(t *testing.T) {
	vault := Vault{}.WithTokens([]OAuth2Token{{Owner: "a"}, {Owner: "b"}})
	if _, found := vault.OAuth2Tokens["a"]; !found {
		t.Fatal("token a missing")
	}
	if _, found := vault.OAuth2Tokens["b"]; !found {
		t.Fatal("token b missing")
	}
}

func TestCookieJarStateChangedSince(t *testing.T) {
	state := CookieJarState{Revision: 3}
	if state.ChangedSince(3) || !state.ChangedSince(2) {
		t.Fatalf("ChangedSince = %v / %v", state.ChangedSince(3), state.ChangedSince(2))
	}
}

func TestAuthIsBasic(t *testing.T) {
	if !(Auth{Type: authtype.Basic}).IsBasic() || (Auth{Type: authtype.Bearer}).IsBasic() {
		t.Fatal("IsBasic")
	}
}

func TestCookieMovedFrom(t *testing.T) {
	cookie := Cookie{Domain: "a.test", Path: "/", Name: "n"}
	if cookie.MovedFrom(nil) || cookie.MovedFrom(&Cookie{Domain: "A.test", Path: "/", Name: "n"}) || !cookie.MovedFrom(&Cookie{Domain: "a.test", Path: "/", Name: "old"}) {
		t.Fatal("MovedFrom")
	}
}
