package entity

import (
	"net/url"
	"strings"
)

const (
	defaultScheme   = "http://"
	schemeSeparator = "://"
	pathVariableTag = ":"
	upperHex        = "0123456789ABCDEF"
	hexDigits       = upperHex + "abcdef"
	subDelimiters   = "!$&'()*+,;="

	unreservedCharacters = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~"
)

type RequestUrl struct {
	base        string
	query       string
	fragment    string
	hasQuery    bool
	hasFragment bool
}

func ParseRequestUrl(raw string) RequestUrl {
	rest, fragment, hasFragment := strings.Cut(raw, "#")
	base, query, hasQuery := strings.Cut(rest, "?")
	return RequestUrl{base: base, query: query, fragment: fragment, hasQuery: hasQuery, hasFragment: hasFragment}
}

func (r RequestUrl) String() string {
	result := r.base
	if r.hasQuery {
		result += "?" + r.query
	}
	if r.hasFragment {
		result += "#" + r.fragment
	}
	return result
}

func (r RequestUrl) QueryParams() []KeyValue {
	params := make([]KeyValue, 0)
	for _, pair := range strings.Split(r.query, "&") {
		if pair == "" {
			continue
		}
		key, value, _ := strings.Cut(pair, "=")
		params = append(params, NewKeyValue(key, value))
	}
	return params
}

func (r RequestUrl) WithQuery(params []KeyValue) RequestUrl {
	pairs := make([]string, 0, len(params))
	for _, param := range params {
		if param.IsActive() {
			pairs = append(pairs, queryPair(param))
		}
	}
	r.query = strings.Join(pairs, "&")
	r.hasQuery = len(pairs) > 0
	return r
}

func (r RequestUrl) withoutQueryKey(name string) RequestUrl {
	kept := make([]string, 0)
	for _, pair := range strings.Split(r.query, "&") {
		key, _, _ := strings.Cut(pair, "=")
		if pair != "" && decodedQueryKey(key) != name {
			kept = append(kept, pair)
		}
	}
	r.query = strings.Join(kept, "&")
	r.hasQuery = len(kept) > 0
	return r
}

func (r RequestUrl) withQueryPair(key, value string) RequestUrl {
	pair := url.QueryEscape(key) + "=" + url.QueryEscape(value)
	if r.query != "" {
		pair = r.query + "&" + pair
	}
	r.query = pair
	r.hasQuery = true
	return r
}

func decodedQueryKey(raw string) string {
	decoded, err := url.QueryUnescape(raw)
	if err != nil {
		return raw
	}
	return decoded
}

func (r RequestUrl) PathVariableNames() []string {
	names := make([]string, 0)
	for _, segment := range strings.Split(r.base, "/") {
		name, isVariable := strings.CutPrefix(segment, pathVariableTag)
		if isVariable && name != "" && !contains(names, name) {
			names = append(names, name)
		}
	}
	return names
}

func (r RequestUrl) WithPathVariableNames(names []string) RequestUrl {
	renames := make(map[string]string)
	for index, current := range r.PathVariableNames() {
		if index < len(names) && strings.TrimSpace(names[index]) != "" {
			renames[current] = pathVariableTag + strings.TrimSpace(names[index])
		}
	}
	r.base = r.mapPathVariables(func(name string) (string, bool) {
		replacement, found := renames[name]
		return replacement, found
	})
	return r
}

func (r RequestUrl) WithPathValues(variables []PathVariable) RequestUrl {
	r.base = r.mapPathVariables(func(name string) (string, bool) {
		for _, variable := range variables {
			if variable.HasKey(name) && variable.Value != "" {
				return variable.Value, true
			}
		}
		return "", false
	})
	return r
}

func (r RequestUrl) WithDefaultScheme() RequestUrl {
	if strings.TrimSpace(r.base) != "" && !strings.Contains(r.base, schemeSeparator) {
		r.base = defaultScheme + r.base
	}
	return r
}

func (r RequestUrl) WithoutFragment() RequestUrl {
	r.fragment, r.hasFragment = "", false
	return r
}

func (r RequestUrl) Encoded() RequestUrl {
	scheme, rest, hasScheme := strings.Cut(r.base, schemeSeparator)
	if hasScheme {
		host, path, hasPath := strings.Cut(rest, "/")
		r.base = scheme + schemeSeparator + host
		if hasPath {
			r.base += "/" + percentEncode(path, "/:@")
		}
	}
	r.query = percentEncode(r.query, "/:@?")
	return r
}

func (r RequestUrl) Host() string {
	parsed, err := url.Parse(r.WithDefaultScheme().String())
	if err != nil {
		return ""
	}
	return parsed.Host
}

func (r RequestUrl) mapPathVariables(replace func(name string) (string, bool)) string {
	segments := strings.Split(r.base, "/")
	for index, segment := range segments {
		name, isVariable := strings.CutPrefix(segment, pathVariableTag)
		if replacement, found := replace(name); isVariable && found {
			segments[index] = replacement
		}
	}
	return strings.Join(segments, "/")
}

func queryPair(param KeyValue) string {
	if param.Value == "" {
		return param.Key
	}
	return param.Key + "=" + param.Value
}

func percentEncode(text, allowed string) string {
	safe := unreservedCharacters + subDelimiters + allowed
	var builder strings.Builder
	for index := 0; index < len(text); index++ {
		character := text[index]
		if strings.ContainsRune(safe, rune(character)) || isEscape(text, index) {
			builder.WriteByte(character)
			continue
		}
		builder.WriteByte('%')
		builder.WriteByte(upperHex[character>>4])
		builder.WriteByte(upperHex[character&15])
	}
	return builder.String()
}

func isEscape(text string, index int) bool {
	return text[index] == '%' && index+2 < len(text) && isHex(text[index+1]) && isHex(text[index+2])
}

func isHex(character byte) bool {
	return strings.ContainsRune(hexDigits, rune(character))
}

func contains(names []string, name string) bool {
	for _, existing := range names {
		if existing == name {
			return true
		}
	}
	return false
}
