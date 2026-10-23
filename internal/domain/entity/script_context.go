package entity

import "maps"

// ScriptContext is what one script starts from; Cookies are pm.cookies (the cookies for the request's domain) and
// ResponseCookies are the ones the response set (pm.response.cookies).
type ScriptContext struct {
	Response        *Response
	Dynamic         DynamicValue
	Console         ScriptLogSink
	Vault           map[string]string
	IterationData   map[string]any
	Info            ScriptInfo
	Scopes          ScriptScopes
	Cookies         []ScriptCookie
	ResponseCookies []ScriptCookie
	Request         Request
}

func NewScriptContext(info ScriptInfo, iterationData map[string]any, cookies, responseCookies []ScriptCookie, dynamic DynamicValue, console ScriptLogSink) ScriptContext {
	return ScriptContext{Info: info, IterationData: iterationData, Cookies: cookies, ResponseCookies: responseCookies, Dynamic: dynamic, Console: console}
}

func (s ScriptContext) Over(request Request, response *Response, scopes ScriptScopes, vault map[string]string) ScriptContext {
	s.Request = request
	s.Response = response
	s.Scopes = scopes
	s.Vault = maps.Clone(vault)
	return s
}
