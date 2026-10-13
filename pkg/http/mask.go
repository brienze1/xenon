package http

import (
	"net/http"
	"net/url"
	"strings"
)

const maskedValue = "******"

var sensitiveHeaders = map[string]bool{
	"Authorization":       true,
	"Cookie":              true,
	"Set-Cookie":          true,
	"Proxy-Authorization": true,
	"X-Api-Key":           true,
}

var sensitiveQueryKeys = map[string]bool{
	"access_token":  true,
	"refresh_token": true,
	"id_token":      true,
	"token":         true,
	"api_key":       true,
	"apikey":        true,
	"password":      true,
	"secret":        true,
	"client_secret": true,
}

func maskHeaders(headers http.Header) map[string][]string {
	masked := make(map[string][]string, len(headers))
	for key, values := range headers {
		if sensitiveHeaders[http.CanonicalHeaderKey(key)] {
			masked[key] = []string{maskedValue}
			continue
		}
		masked[key] = values
	}
	return masked
}

func maskUrl(target *url.URL) string {
	if target == nil {
		return ""
	}
	copied := *target
	copied.RawQuery = maskQuery(copied.RawQuery)
	return copied.Redacted()
}

func maskQuery(rawQuery string) string {
	if rawQuery == "" {
		return ""
	}
	pairs := strings.Split(rawQuery, "&")
	for index, pair := range pairs {
		key, _, hasValue := strings.Cut(pair, "=")
		name, err := url.QueryUnescape(key)
		if err != nil {
			name = key
		}
		if hasValue && sensitiveQueryKeys[strings.ToLower(name)] {
			pairs[index] = key + "=" + maskedValue
		}
	}
	return strings.Join(pairs, "&")
}
