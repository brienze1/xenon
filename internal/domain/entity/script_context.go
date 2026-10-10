package entity

// ScriptContext is what one script starts from; Cookies are pm.cookies (the cookies for the request's domain) and
// ResponseCookies are the ones the response set (pm.response.cookies).
type ScriptContext struct {
	Response        *Response
	Dynamic         DynamicValue
	Console         ScriptLogSink
	Vault           map[string]string
	Info            ScriptInfo
	Scopes          ScriptScopes
	Cookies         []ScriptCookie
	ResponseCookies []ScriptCookie
	Request         Request
}
