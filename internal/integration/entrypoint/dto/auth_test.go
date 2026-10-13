package dto

import (
	"encoding/json"
	"reflect"
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
)

func TestAuthSubObjectsRoundTrip(t *testing.T) {
	auth := entity.Auth{
		Type:   authtype.Digest,
		ApiKey: entity.ApiKeyAuth{Key: "k", Value: "v", In: entity.ApiKeyInQuery},
		OAuth2: entity.OAuth2Config{
			GrantType: oauth2grant.Implicit, ClientAuthentication: clientauthentication.Body, AddTokenTo: tokenplacement.QueryParams, TokenName: "n",
			RedirectURI: "https://oauth.pstmn.io/v1/callback", AuthURL: "a", AccessTokenURL: "t", RefreshTokenURL: "r", ClientID: "c", ClientSecret: "s",
			Username: "u", Password: "p", Scope: "sc", State: "st", ChallengeAlgorithm: entity.PkcePlain, CodeVerifier: "v", HeaderPrefix: "", UseToken: entity.UseIDToken,
			Audience: "au", Resource: "re", AutoRefresh: true,
			AuthRequestParams:    []entity.OAuth2Param{{Key: "k", Value: "v", SendAs: entity.OAuth2SendAsUrl, Enabled: true}},
			TokenRequestParams:   []entity.OAuth2Param{},
			RefreshRequestParams: []entity.OAuth2Param{},
		},
		AwsV4:  entity.AwsV4Auth{AccessKey: "a", SecretKey: "s", SessionToken: "t", Region: "r", Service: "v", AddAuthDataToQuery: true},
		Digest: entity.DigestAuth{Username: "u", Password: "p", Algorithm: digestalgorithm.SHA256Sess, Qop: digestqop.Auth, Realm: "r", Nonce: "n", Opaque: "o", NonceCount: "1", ClientNonce: "c", DisableRetryRequest: true},
	}
	result := NewAuth(auth)
	if !result.OAuth2.Legacy || !result.OAuth2.UsesPostmanCallback || result.OAuth2.SuggestedCallback != entity.DefaultOAuth2Callback {
		t.Fatalf("oauth2 flags = %+v", result.OAuth2)
	}
	encoded, _ := json.Marshal(result)
	var decoded Auth
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if back := decoded.ToEntity(); !reflect.DeepEqual(back, auth) {
		t.Fatalf("ToEntity() =\n%+v\nwant\n%+v", back, auth)
	}
	if log := result.OAuth2.LogValue().(map[string]any); log["has_client_secret"] != true || log["client_secret"] != nil {
		t.Fatalf("LogValue = %v", log)
	}
	plain := NewAuth(entity.Auth{Type: authtype.Bearer, OAuth2: entity.DefaultOAuth2Config()})
	if plain.ApiKey != nil || plain.OAuth2 != nil || plain.AwsV4 != nil || plain.Digest != nil {
		t.Fatalf("blank sub-objects are left out: %+v", plain)
	}
	if NewAuthOAuth2(entity.DefaultOAuth2Config()).SuggestedCallback != "" {
		t.Fatal("no guidance for a loopback callback")
	}
}

func TestAbsentAuthSubObjectsAreZeroAndPresentOnesGetDefaults(t *testing.T) {
	auth := (&Auth{Type: "oauth2"}).ToEntity()
	if !reflect.DeepEqual(auth.OAuth2, entity.OAuth2Config{}) || auth.ApiKey != (entity.ApiKeyAuth{}) || auth.AwsV4 != (entity.AwsV4Auth{}) || auth.Digest != (entity.DigestAuth{}) {
		t.Fatalf("absent = %+v", auth)
	}
	var decoded Auth
	_ = json.Unmarshal([]byte(`{"type":"oauth2","apikey":{"in":"cookie"},"oauth2":{"grant_type":"x","use_token":"other","token_request_params":[{"key":"a","enabled":true}]},"digest":{"algorithm":"SHA-1","qop":"x"}}`), &decoded)
	present := decoded.ToEntity()
	if present.ApiKey != entity.DefaultApiKeyAuth() || present.OAuth2.GrantType != oauth2grant.AuthorizationCode || present.OAuth2.RedirectURI != entity.DefaultOAuth2Callback || present.OAuth2.UseToken != entity.UseAccessToken || present.OAuth2.TokenRequestParams[0].SendAs != entity.OAuth2SendAsBody || present.Digest.Algorithm != digestalgorithm.MD5 || present.Digest.Qop != digestqop.None {
		t.Fatalf("present = %+v", present)
	}
}

func TestOAuthDtos(t *testing.T) {
	request := OAuthTokenRequest{LocalId: "ws", OwnerId: "o"}
	if request.Draft() != nil || request.LogValue().(map[string]any)["auth"] != nil {
		t.Fatal("no draft")
	}
	request.Auth = &AuthOAuth2{ClientId: "c", ClientSecret: "s"}
	if draft := request.Draft(); draft == nil || draft.ClientID != "c" {
		t.Fatalf("draft = %+v", draft)
	}
	if log := request.LogValue().(map[string]any); log["auth"] == nil || log["owner_id"] != "o" {
		t.Fatalf("LogValue = %v", log)
	}
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	expires := now.Add(-time.Minute)
	info := NewTokenInfo(entity.OAuth2Token{Owner: "o", TokenType: "Bearer", Scope: "s", UseToken: entity.UseIDToken, RefreshToken: "r", IDToken: "i", ExpiresAt: &expires, ObtainedAt: now}, now)
	if info.OwnerId != "o" || !info.Expired || !info.HasRefreshToken || !info.HasIdToken || info.UseToken != entity.UseIDToken || info.Scope != "s" || info.TokenType != "Bearer" || !info.ObtainedAt.Equal(now) {
		t.Fatalf("info = %+v", info)
	}
	send := SendRequest{SendWithExpiredToken: true}
	if !send.ToAcknowledgement().AcceptsExpiredToken() {
		t.Fatal("send_with_expired_token")
	}
	settings := RequestSettings{DisableCookieJar: true}
	if !settings.ToEntity().DisableCookieJar || !NewRequestSettings(entity.RequestSettings{DisableCookieJar: true}).DisableCookieJar {
		t.Fatal("disable_cookie_jar")
	}
}

func TestUIStateCarriesClosedTabsAndSnippetOptions(t *testing.T) {
	state := entity.NewUIState()
	state.ClosedTabs = []entity.ClosedTab{{ClosedAt: time.Unix(5, 0).UTC(), Tab: entity.Tab{TabID: "t", RequestID: "r"}}}
	state.Snippets = []entity.SnippetOptions{{Language: snippetlanguage.GoNative, Options: map[string]string{"timeout": "5"}}}
	result := NewUIState("ws", state)
	back := result.ToEntity()
	if len(back.ClosedTabs) != 1 || back.ClosedTabs[0].Tab.TabID != "t" || !back.ClosedTabs[0].ClosedAt.Equal(time.Unix(5, 0)) || back.Snippets[0].Language != snippetlanguage.GoNative || back.Snippets[0].Options["timeout"] != "5" {
		t.Fatalf("ToEntity() = %+v", back)
	}
}
