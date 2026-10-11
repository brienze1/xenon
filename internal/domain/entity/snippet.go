package entity

import (
	"bytes"
	"encoding/json"
	"net/url"
	"slices"
	"strings"
)

const (
	NoOAuthTokenNotice     = "No OAuth 2.0 access token yet: get one in the Authorization tab."
	MacTokenNotice         = "OAuth 2.0 tokens of type mac are not sent, so the snippet has no auth."
	DigestChallengeNotice  = "Digest: send the request once so Xenon learns the server's challenge."
	unsupportedAuthPrefix  = "Auth type "
	unsupportedAuthSuffix  = " is not supported, so the snippet has no auth."
	awsSignatureMarker     = "Signature="
	awsSignatureQueryParam = "X-Amz-Signature"
	digestResponseMarker   = `response="`
)

type Snippet struct {
	Text    string
	Notices []string
}

func (s Snippet) LogValue() any {
	return map[string]any{"length": len(s.Text), "notices": s.Notices}
}

func UnsupportedAuthNotice(auth Auth) string {
	return unsupportedAuthPrefix + auth.Unsupported.Type + unsupportedAuthSuffix
}

func (a Auth) DerivedSecrets(request EffectiveRequest, secrets []string) []string {
	switch {
	case a.IsBasic() && usesSecret(secrets, a.Basic.Username, a.Basic.Password):
		return []string{strings.TrimPrefix(a.Basic.Header(), basicScheme)}
	case a.IsAwsV4() && usesSecret(secrets, a.AwsV4.SecretKey):
		return nonEmpty(awsSignature(request))
	case a.IsDigest() && usesSecret(secrets, a.Digest.Password):
		return nonEmpty(digestResponse(request))
	default:
		return []string{}
	}
}

func usesSecret(secrets []string, values ...string) bool {
	for _, secret := range secrets {
		for _, value := range values {
			if secret != "" && strings.Contains(value, secret) {
				return true
			}
		}
	}
	return false
}

func awsSignature(request EffectiveRequest) string {
	for _, header := range request.Headers {
		if _, signature, found := strings.Cut(header.Value, awsSignatureMarker); found && header.HasKey(HeaderAuthorization) {
			return signature
		}
	}
	return ParseRequestUrl(request.URL).QueryValue(awsSignatureQueryParam)
}

func digestResponse(request EffectiveRequest) string {
	for _, header := range request.Headers {
		if _, rest, found := strings.Cut(header.Value, digestResponseMarker); found && header.HasKey(HeaderAuthorization) {
			response, _, _ := strings.Cut(rest, `"`)
			return response
		}
	}
	return ""
}

func nonEmpty(value string) []string {
	if value == "" {
		return []string{}
	}
	return []string{value}
}

func (s Snippet) WithText(text string) Snippet {
	s.Text = text
	return s
}

func SnippetSecretForms(secrets []string) []string {
	forms := make([]string, 0, len(secrets)*4)
	for _, secret := range secrets {
		if secret == "" {
			continue
		}
		forms = append(forms, secret, url.QueryEscape(secret), url.PathEscape(secret), strings.ReplaceAll(url.QueryEscape(secret), "+", "%20"))
		forms = append(forms, escapedSecretForms(secret)...)
	}
	return slices.Compact(forms)
}

func escapedSecretForms(secret string) []string {
	var quoted bytes.Buffer
	encoder := json.NewEncoder(&quoted)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(secret)
	return []string{
		strings.Trim(strings.TrimSpace(quoted.String()), `"`),
		strings.ReplaceAll(secret, "'", `'\''`),
		strings.NewReplacer(`\`, `\\`, "'", `\'`).Replace(secret),
		strings.NewReplacer(`"`, `\"`, "$", `\$`, "`", "\\`", "!", `\!`, "#", `\#`).Replace(secret),
	}
}

func OAuthSnippetNotices(token *OAuth2Token) []string {
	switch {
	case token == nil || token.Value() == "":
		return []string{NoOAuthTokenNotice}
	case token.IsMac():
		return []string{MacTokenNotice}
	default:
		return nil
	}
}
