package model

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/authtype"
	"github.com/brienze1/xenon/internal/domain/enums/clientauthentication"
	"github.com/brienze1/xenon/internal/domain/enums/digestalgorithm"
	"github.com/brienze1/xenon/internal/domain/enums/digestqop"
	"github.com/brienze1/xenon/internal/domain/enums/oauth2grant"
	"github.com/brienze1/xenon/internal/domain/enums/snippetlanguage"
	"github.com/brienze1/xenon/internal/domain/enums/tokenplacement"
	"github.com/brienze1/xenon/pkg/yamlnode"
)

func fullOAuth2() entity.OAuth2Config {
	return entity.OAuth2Config{
		GrantType: oauth2grant.AuthorizationCodeWithPkce, ClientAuthentication: clientauthentication.Body, AddTokenTo: tokenplacement.QueryParams,
		TokenName: "n", RedirectURI: "http://127.0.0.1:1/cb", AuthURL: "a", AccessTokenURL: "t", RefreshTokenURL: "r", ClientID: "c", ClientSecret: "s",
		Username: "u", Password: "p", Scope: "sc", State: "st", ChallengeAlgorithm: entity.PkcePlain, CodeVerifier: "v", HeaderPrefix: "", UseToken: entity.UseIDToken,
		Audience: "au", Resource: "re", AutoRefresh: false,
		AuthRequestParams:    []entity.OAuth2Param{{Key: "k", Value: "v", SendAs: entity.OAuth2SendAsUrl, Enabled: true}},
		TokenRequestParams:   []entity.OAuth2Param{{Key: "h", Value: "1", SendAs: entity.OAuth2SendAsHeader}},
		RefreshRequestParams: []entity.OAuth2Param{},
	}
}

func TestEveryAuthSubObjectRoundTrips(t *testing.T) {
	auth := entity.Auth{
		Type:   authtype.OAuth2,
		ApiKey: entity.ApiKeyAuth{Key: "k", Value: "v", In: entity.ApiKeyInQuery},
		OAuth2: fullOAuth2(),
		AwsV4:  entity.AwsV4Auth{AccessKey: "a", SecretKey: "s", SessionToken: "t", Region: "r", Service: "v", AddAuthDataToQuery: true},
		Digest: entity.DigestAuth{Username: "u", Password: "p", Algorithm: digestalgorithm.SHA512256Sess, Qop: digestqop.AuthInt, Realm: "r", Nonce: "n", Opaque: "o", NonceCount: "1", ClientNonce: "c", DisableRetryRequest: true},
	}
	stored := NewAuth(auth)
	text := write(t, []yamlnode.Field{stored.Field()})
	for _, fragment := range []string{"  apikey:\n    key: k\n    value: v\n    in: query\n", "    grant_type: authorization_code_with_pkce\n    tokenName: n\n", "    headerPrefix: \"\"\n    useToken: id_token\n    autoRefresh: false\n", "    authRequestParams:\n      - key: k\n        value: v\n        enabled: true\n        send_as: request_url\n", "    refreshRequestParams: []\n", "  awsv4:\n    accessKey: a\n", "    addAuthDataToQuery: true\n", "    algorithm: SHA-512-256-sess\n    qop: auth-int\n", "    disableRetryRequest: true\n"} {
		if !strings.Contains(text, fragment) {
			t.Errorf("file misses %q:\n%s", fragment, text)
		}
	}
	var decoded struct {
		Auth Auth `yaml:"auth"`
	}
	read(t, text, &decoded)
	if roundTrip := decoded.Auth.ToEntity(); !reflect.DeepEqual(roundTrip, auth) {
		t.Fatalf("ToEntity() =\n%+v\nwant\n%+v", roundTrip, auth)
	}
	encoded, _ := json.Marshal(stored)
	var draft Auth
	if err := json.Unmarshal(encoded, &draft); err != nil || !reflect.DeepEqual(draft.ToEntity(), auth) {
		t.Fatalf("a draft in state.json keeps the auth: %s", encoded)
	}
}

func TestPartialSubObjectsGetDefaults(t *testing.T) {
	var decoded struct {
		Auth Auth `yaml:"auth"`
	}
	read(t, "auth:\n  type: apikey\n  apikey:\n    in: cookie\n  oauth2:\n    grant_type: device\n    challengeAlgorithm: S512\n    client_authentication: other\n    useToken: refresh\n    tokenRequestParams:\n      - key: a\n  digest:\n    algorithm: SHA-1\n    qop: weird\n", &decoded)
	auth := decoded.Auth.ToEntity()
	defaults := entity.DefaultOAuth2Config()
	if auth.ApiKey != entity.DefaultApiKeyAuth() || auth.OAuth2.GrantType != defaults.GrantType || auth.OAuth2.ChallengeAlgorithm != entity.PkceS256 || auth.OAuth2.ClientAuthentication != clientauthentication.Header || auth.OAuth2.UseToken != entity.UseAccessToken || auth.OAuth2.RedirectURI != entity.DefaultOAuth2Callback || auth.OAuth2.HeaderPrefix != "Bearer" || !auth.OAuth2.AutoRefresh {
		t.Fatalf("defaults = %+v", auth)
	}
	if auth.OAuth2.TokenRequestParams[0] != (entity.OAuth2Param{Key: "a", SendAs: entity.OAuth2SendAsBody, Enabled: true}) {
		t.Fatalf("param defaults = %+v", auth.OAuth2.TokenRequestParams)
	}
	if auth.Digest.Algorithm != digestalgorithm.MD5 || auth.Digest.Qop != digestqop.None {
		t.Fatalf("digest defaults = %+v", auth.Digest)
	}
	if NewAuth(entity.Auth{Type: authtype.Bearer}).OAuth2 != nil || NewAuth(entity.Auth{Type: authtype.AwsV4}).AwsV4 != nil {
		t.Fatal("blank sub-objects are not written")
	}
}

func TestVaultContentKeepsTokens(t *testing.T) {
	expires := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	vault := entity.Vault{Secrets: map[string]map[string]string{"globals": {"a": "1"}, "empty": {}}, OAuth2Tokens: map[string]entity.OAuth2Token{"o": {Owner: "o", AccessToken: "a", RefreshToken: "r", IDToken: "i", TokenType: "Bearer", Scope: "s", UseToken: entity.UseAccessToken, ExpiresAt: &expires, ObtainedAt: expires}}}
	content := NewVaultContent(vault)
	encoded, _ := json.Marshal(content)
	var decoded VaultContent
	_ = json.Unmarshal(encoded, &decoded)
	storage := decoded.ToEntity(entity.VaultStorage{LocalID: "ws"})
	if !reflect.DeepEqual(storage.OAuth2Tokens, vault.OAuth2Tokens) || len(storage.Secrets) != 1 || storage.LocalID != "ws" {
		t.Fatalf("storage = %+v from %s", storage, encoded)
	}
}

func TestDigestStateClosedTabsAndSnippetModels(t *testing.T) {
	state := entity.DigestState{Realm: "r", Nonce: "n", Opaque: "o", Algorithm: digestalgorithm.MD5Sess, Qop: digestqop.Auth, NonceCount: 4}
	stored := NewDigestState(state)
	if stored.ToEntity() != state {
		t.Fatalf("digest state = %+v", stored.ToEntity())
	}
	closed := entity.ClosedTab{ClosedAt: time.Unix(1, 0).UTC(), Tab: entity.Tab{TabID: "t", RequestID: "r"}}
	closedModel := NewClosedTab(closed)
	if back := closedModel.ToEntity(); back.ClosedAt != closed.ClosedAt || back.Tab.TabID != "t" || back.Tab.RequestID != "r" {
		t.Fatalf("closed tab = %+v", back)
	}
	options := entity.SnippetOptions{Language: snippetlanguage.PythonRequests, Options: map[string]string{"a": "b"}}
	optionsModel := NewSnippetOptions(options)
	if !reflect.DeepEqual(optionsModel.ToEntity(), options) {
		t.Fatalf("snippet options = %+v", optionsModel.ToEntity())
	}
	settings := NewSettings(entity.RequestSettings{DisableCookieJar: true})
	if !settings.ToEntity().DisableCookieJar || !strings.Contains(write(t, []yamlnode.Field{settings.Field()}), "disableCookieJar: true") {
		t.Fatal("disableCookieJar round-trips")
	}
}
