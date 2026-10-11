package steps

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

const (
	postmanDefaultAttributeType = "string"
	postmanDefaultRawLanguage   = "text"
	postmanSecretType           = "secret"
	postmanDiffLimit            = 25
)

var postmanUncomparedKeys = []string{"_exporter_id", "_postman_exported_at", "_postman_exported_using"}

// normalisePostman applies the XEN-006 export-formats §4 rules so two Postman documents compare equal when Postman would treat them the same.
func normalisePostman(value any) any {
	return normaliseValue("", value)
}

func normaliseValue(key string, value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return normaliseObject(key, typed)
	case []any:
		return normaliseList(key, typed)
	default:
		return value
	}
}

func normaliseObject(key string, object map[string]any) any {
	if key == "auth" {
		object = normaliseAuth(object)
	}
	if key == "request" {
		object = withUpperMethod(object)
	}
	if isSecretVariable(object) {
		object = withBlankValue(object)
	}
	result := map[string]any{}
	for name, child := range object {
		if isUncomparedKey(name) {
			continue
		}
		if name == "url" {
			result[name] = normaliseUrl(child)
			continue
		}
		if name == "request" {
			child = requestObject(child)
		}
		normalised := normaliseValue(name, child)
		if isPostmanDefault(name, normalised) {
			continue
		}
		result[name] = normalised
	}
	return result
}

func normaliseList(key string, list []any) []any {
	result := make([]any, 0, len(list))
	for _, item := range list {
		result = append(result, normaliseValue(key, item))
	}
	return result
}

func isUncomparedKey(name string) bool {
	for _, uncompared := range postmanUncomparedKeys {
		if name == uncompared {
			return true
		}
	}
	return false
}

func isPostmanDefault(name string, value any) bool {
	switch typed := value.(type) {
	case nil:
		return name == "description"
	case bool:
		return name == "disabled" && !typed
	case string:
		return (name == "description" && typed == "") || (name == "language" && typed == postmanDefaultRawLanguage)
	case []any:
		return len(typed) == 0
	case map[string]any:
		return len(typed) == 0 || (name == "description" && len(typed) == 1 && typed["content"] == "")
	default:
		return false
	}
}

func isSecretVariable(object map[string]any) bool {
	_, hasKey := object["key"]
	return hasKey && object["type"] == postmanSecretType
}

func withBlankValue(object map[string]any) map[string]any {
	copied := map[string]any{}
	for name, child := range object {
		copied[name] = child
	}
	copied["value"] = ""
	return copied
}

func withUpperMethod(object map[string]any) map[string]any {
	method, isText := object["method"].(string)
	if !isText {
		return object
	}
	copied := map[string]any{}
	for name, child := range object {
		copied[name] = child
	}
	copied["method"] = strings.ToUpper(method)
	return copied
}

func requestObject(request any) any {
	if text, isText := request.(string); isText {
		return map[string]any{"method": "GET", "url": text}
	}
	return request
}

func normaliseUrl(url any) any {
	switch typed := url.(type) {
	case string:
		return urlComparison(typed, queryRows(typed), nil)
	case map[string]any:
		raw, hasRaw := typed["raw"].(string)
		if !hasRaw {
			raw = rebuiltUrl(typed)
		}
		return urlComparison(raw, typed["query"], typed["variable"])
	default:
		return url
	}
}

func urlComparison(raw string, query, variable any) map[string]any {
	comparison := map[string]any{"raw": raw}
	if normalised := normaliseValue("query", query); !isPostmanDefault("query", normalised) && normalised != nil {
		comparison["query"] = normalised
	}
	if normalised := normaliseValue("variable", variable); !isPostmanDefault("variable", normalised) && normalised != nil {
		comparison["variable"] = normalised
	}
	return comparison
}

func queryRows(raw string) any {
	_, query, found := strings.Cut(raw, "?")
	if !found {
		return nil
	}
	query, _, _ = strings.Cut(query, "#")
	rows := make([]any, 0)
	for _, pair := range strings.Split(query, "&") {
		if pair == "" {
			continue
		}
		key, value, _ := strings.Cut(pair, "=")
		rows = append(rows, map[string]any{"key": key, "value": value})
	}
	return rows
}

func rebuiltUrl(url map[string]any) string {
	built := joinedParts(url["host"], ".")
	if protocol, isText := url["protocol"].(string); isText && protocol != "" {
		built = protocol + "://" + built
	}
	if port := stringify(url["port"]); url["port"] != nil && port != "" {
		built += ":" + port
	}
	if path := joinedParts(url["path"], "/"); path != "" {
		built += "/" + path
	}
	if rows, isList := url["query"].([]any); isList && len(rows) > 0 {
		pairs := make([]string, 0, len(rows))
		for _, row := range rows {
			entry, _ := row.(map[string]any)
			pairs = append(pairs, stringify(entry["key"])+"="+stringify(entry["value"]))
		}
		built += "?" + strings.Join(pairs, "&")
	}
	if hash, isText := url["hash"].(string); isText && hash != "" {
		built += "#" + hash
	}
	return built
}

func joinedParts(value any, separator string) string {
	parts, isList := value.([]any)
	if !isList {
		if text, isText := value.(string); isText {
			return text
		}
		return ""
	}
	segments := make([]string, 0, len(parts))
	for _, part := range parts {
		segments = append(segments, stringify(part))
	}
	return strings.Join(segments, separator)
}

func normaliseAuth(auth map[string]any) map[string]any {
	authType, _ := auth["type"].(string)
	result := map[string]any{}
	for name, value := range auth {
		result[name] = value
	}
	if attributes, found := auth[authType]; found {
		result[authType] = authAttributeSet(authType, attributes)
	}
	return result
}

func authAttributeSet(authType string, attributes any) any {
	set := map[string]any{}
	switch typed := attributes.(type) {
	case []any:
		for _, attribute := range typed {
			entry, isMap := attribute.(map[string]any)
			if !isMap {
				return attributes
			}
			set[stringify(entry["key"])] = attributeComparison(entry)
		}
	case map[string]any:
		for key, value := range typed {
			set[key] = map[string]any{"value": value}
		}
	default:
		return attributes
	}
	if authType == "oauth2" {
		withoutLiteralToken(set)
	}
	return set
}

func attributeComparison(entry map[string]any) map[string]any {
	comparison := map[string]any{}
	for name, value := range entry {
		if name == "key" || (name == "type" && value == postmanDefaultAttributeType) {
			continue
		}
		comparison[name] = value
	}
	return comparison
}

func withoutLiteralToken(set map[string]any) {
	token, found := set["accessToken"].(map[string]any)
	if !found {
		return
	}
	if value, isText := token["value"].(string); isText && !strings.Contains(value, "{{") {
		delete(set, "accessToken")
		delete(set, "tokenType")
	}
}

func postmanDifferences(path string, expected, actual any) []string {
	switch typed := expected.(type) {
	case map[string]any:
		other, isMap := actual.(map[string]any)
		if !isMap {
			return []string{fmt.Sprintf("%s: expected an object, got %s", pathOrRoot(path), stringify(actual))}
		}
		return objectDifferences(path, typed, other)
	case []any:
		other, isList := actual.([]any)
		if !isList || len(other) != len(typed) {
			return []string{fmt.Sprintf("%s: expected %s, got %s", pathOrRoot(path), stringify(expected), stringify(actual))}
		}
		differences := make([]string, 0)
		for index := range typed {
			differences = append(differences, postmanDifferences(path+"."+strconv.Itoa(index), typed[index], other[index])...)
		}
		return differences
	default:
		if !reflect.DeepEqual(expected, actual) {
			return []string{fmt.Sprintf("%s: expected %s, got %s", pathOrRoot(path), stringify(expected), stringify(actual))}
		}
		return nil
	}
}

func objectDifferences(path string, expected, actual map[string]any) []string {
	keys := map[string]bool{}
	for key := range expected {
		keys[key] = true
	}
	for key := range actual {
		keys[key] = true
	}
	names := make([]string, 0, len(keys))
	for key := range keys {
		names = append(names, key)
	}
	sort.Strings(names)
	differences := make([]string, 0)
	for _, name := range names {
		child := strings.TrimPrefix(path+"."+name, ".")
		expectedValue, inExpected := expected[name]
		actualValue, inActual := actual[name]
		switch {
		case !inActual:
			differences = append(differences, fmt.Sprintf("%s: missing, expected %s", child, stringify(expectedValue)))
		case !inExpected:
			differences = append(differences, fmt.Sprintf("%s: unexpected %s", child, stringify(actualValue)))
		default:
			differences = append(differences, postmanDifferences(child, expectedValue, actualValue)...)
		}
	}
	return differences
}

func pathOrRoot(path string) string {
	if path == "" {
		return "(root)"
	}
	return strings.TrimPrefix(path, ".")
}

func limitedDifferences(differences []string) string {
	if len(differences) > postmanDiffLimit {
		differences = append(differences[:postmanDiffLimit], fmt.Sprintf("… and %d more", len(differences)-postmanDiffLimit))
	}
	return strings.Join(differences, "\n")
}
