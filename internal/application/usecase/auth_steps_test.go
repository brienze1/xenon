package usecase

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/authtype"
	"github.com/brienze1/xenon/internal/domain/enums/bodymode"
	"github.com/brienze1/xenon/internal/domain/enums/oauth2grant"
	"github.com/brienze1/xenon/internal/domain/enums/tokenplacement"
	"github.com/brienze1/xenon/pkg/errs"
)

func TestApplyApiKey(t *testing.T) {
	request := entity.EffectiveRequest{URL: "https://h/p?api_key=old", Headers: []entity.KeyValue{entity.NewKeyValue("x-api-key", "old")}}
	header := ApplyApiKey(request, entity.ApiKeyAuth{Key: "X-Api-Key", Value: "k", In: entity.ApiKeyInHeader})
	if len(header.Headers) != 1 || header.Headers[0] != entity.NewKeyValue("X-Api-Key", "k") || header.URL != request.URL {
		t.Fatalf("header = %+v", header)
	}
	query := ApplyApiKey(request, entity.ApiKeyAuth{Key: "api_key", Value: "k", In: entity.ApiKeyInQuery})
	if query.URL != "https://h/p?api_key=k" || len(query.Headers) != 1 || query.Headers[0].Value != "old" {
		t.Fatalf("query = %+v", query)
	}
	if nothing := ApplyApiKey(request, entity.ApiKeyAuth{In: entity.ApiKeyInQuery}); nothing.URL != request.URL || len(nothing.Headers) != 1 {
		t.Fatalf("empty = %+v", nothing)
	}
}

func TestApplyOAuthToken(t *testing.T) {
	request := entity.EffectiveRequest{URL: "https://h/p?access_token=old", Headers: []entity.KeyValue{entity.NewKeyValue("authorization", "Basic x")}}
	config := entity.DefaultOAuth2Config()
	token := &entity.OAuth2Token{AccessToken: "a", IDToken: "i", TokenType: "Bearer"}
	header := ApplyOAuthToken(request, config, token)
	if len(header.Headers) != 1 || header.Headers[0] != entity.NewKeyValue("Authorization", "Bearer a") || header.URL != request.URL {
		t.Fatalf("header = %+v", header)
	}
	config.AddTokenTo = tokenplacement.QueryParams
	query := ApplyOAuthToken(request, config, token)
	if query.URL != "https://h/p?access_token=a" || query.Headers[0].Value != "Basic x" {
		t.Fatalf("query = %+v", query)
	}
	for name, skipped := range map[string]*entity.OAuth2Token{"none": nil, "mac": {AccessToken: "a", TokenType: "mac"}, "empty": {UseToken: entity.UseIDToken, AccessToken: "a"}} {
		if got := ApplyOAuthToken(request, config, skipped); got.URL != request.URL || got.Headers[0].Value != "Basic x" {
			t.Errorf("%s token must not be sent: %+v", name, got)
		}
	}
}

func TestBuildAuthorizationUrl(t *testing.T) {
	config := entity.DefaultOAuth2Config()
	config.AuthURL, config.ClientID, config.Scope, config.Audience, config.Resource = "https://sso/authorize", "c", "openid api", "aud", "res"
	config.AuthRequestParams = []entity.OAuth2Param{{Key: "login_hint", Value: "ana", Enabled: true}, {Key: "off", Value: "x"}}
	raw := BuildAuthorizationUrl(config, "st", entity.Pkce{Challenge: "ch", Method: entity.PkceS256})
	parsed, _ := url.Parse(raw)
	query := parsed.Query()
	want := map[string]string{"response_type": "code", "client_id": "c", "redirect_uri": entity.DefaultOAuth2Callback, "state": "st", "scope": "openid api", "audience": "aud", "resource": "res", "login_hint": "ana"}
	for key, value := range want {
		if query.Get(key) != value {
			t.Errorf("%s = %q, want %q", key, query.Get(key), value)
		}
	}
	if query.Has("code_challenge") || query.Has("off") || !strings.HasPrefix(raw, "https://sso/authorize?") {
		t.Fatalf("authorization code without PKCE: %s", raw)
	}
	config.GrantType, config.AuthURL, config.Scope, config.Audience, config.Resource = oauth2grant.AuthorizationCodeWithPkce, "https://sso/authorize?prompt=login", "", "", ""
	pkce := BuildAuthorizationUrl(config, "st", entity.Pkce{Challenge: "ch", Method: entity.PkcePlain})
	if !strings.HasPrefix(pkce, "https://sso/authorize?prompt=login&") || !strings.Contains(pkce, "code_challenge=ch") || !strings.Contains(pkce, "code_challenge_method=plain") || strings.Contains(pkce, "scope=") || strings.Contains(pkce, "audience=") {
		t.Fatalf("pkce = %s", pkce)
	}
	config.GrantType = oauth2grant.Implicit
	if implicit := BuildAuthorizationUrl(config, "st", entity.Pkce{}); !strings.Contains(implicit, "response_type=token") || strings.Contains(implicit, "code_challenge") {
		t.Fatalf("implicit = %s", implicit)
	}
}

func TestGeneratePkceAndState(t *testing.T) {
	config := entity.DefaultOAuth2Config()
	generated := GeneratePkce(config)
	if !regexp.MustCompile(`^[A-Za-z0-9._~-]{64}$`).MatchString(generated.Verifier) || generated.Method != entity.PkceS256 {
		t.Fatalf("generated = %+v", generated)
	}
	sum := sha256.Sum256([]byte(generated.Verifier))
	if generated.Challenge != base64.RawURLEncoding.EncodeToString(sum[:]) {
		t.Fatal("S256 challenge")
	}
	if GeneratePkce(config).Verifier == generated.Verifier {
		t.Fatal("verifiers are random")
	}
	config.CodeVerifier, config.ChallengeAlgorithm = "typed", entity.PkcePlain
	if plain := GeneratePkce(config); plain != (entity.Pkce{Verifier: "typed", Challenge: "typed", Method: entity.PkcePlain}) {
		t.Fatalf("plain = %+v", plain)
	}
	state := GenerateOAuthState(entity.OAuth2Config{})
	if !regexp.MustCompile(`^[0-9a-f]{32}$`).MatchString(state) || GenerateOAuthState(entity.OAuth2Config{}) == state {
		t.Fatalf("state = %q", state)
	}
	if GenerateOAuthState(entity.OAuth2Config{State: "mine"}) != "mine" {
		t.Fatal("a typed state is used")
	}
}

func TestComputePayloadHash(t *testing.T) {
	empty := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	if got := ComputePayloadHash(entity.Body{Mode: bodymode.Raw}, []byte("hello")); got != "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824" {
		t.Fatalf("raw = %s", got)
	}
	if got := ComputePayloadHash(entity.Body{Mode: bodymode.FormData}, []byte("hello")); got != empty {
		t.Fatalf("form-data = %s", got)
	}
}

func TestOAuthRejections(t *testing.T) {
	cases := map[string]error{
		"OAU-01400": RejectInvalidOAuthConfig("why"),
		"OAU-04401": RejectOAuthProviderError("access_denied", "no"),
		"OAU-05400": RejectOAuthStateMismatch(),
		"OAU-06404": RejectMissingRefreshToken(),
		"OAU-10404": RejectMissingOAuthToken(),
		"SND-07409": RejectExpiredToken("owner"),
	}
	for code, err := range cases {
		var custom errs.BaseError
		if !errors.As(err, &custom) || custom.Code() != code {
			t.Errorf("%s: got %v", code, err)
		}
	}
	var provider errs.BaseError
	errors.As(cases["OAU-04401"], &provider)
	if provider.StatusCode() != 401 || !strings.Contains(provider.Description(), "access_denied") {
		t.Fatalf("provider error = %d %q", provider.StatusCode(), provider.Description())
	}
	var invalid errs.BaseError
	if errors.As(cases["OAU-01400"], &invalid); !strings.Contains(invalid.Description(), "why") {
		t.Fatal("the reason is in the message")
	}
}

func TestResolveEffectiveRequestNamesTheAuthOwner(t *testing.T) {
	request := entity.NewRequest("r", "R")
	request.Auth = entity.Auth{Type: authtype.Inherit}
	collection := entity.NewCollection("c", "C")
	collection.Auth = entity.Auth{Type: authtype.ApiKey}
	folder := entity.NewFolder("f", "F")
	if owner := ResolveEffectiveRequest(request, []entity.Collection{collection, folder}, "/w").AuthOwner; owner != "c" {
		t.Fatalf("inherited owner = %q", owner)
	}
	if owner := ResolveEffectiveRequest(request, nil, "/w").AuthOwner; owner != "" {
		t.Fatalf("no owner = %q", owner)
	}
	request.Auth = entity.Auth{Type: authtype.OAuth2}
	if owner := ResolveEffectiveRequest(request, []entity.Collection{collection}, "/w").AuthOwner; owner != "r" {
		t.Fatalf("own owner = %q", owner)
	}
}
