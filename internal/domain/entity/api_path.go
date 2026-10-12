package entity

import (
	"regexp"
	"slices"
	"strings"
)

var (
	xenonVariable = regexp.MustCompile(`\{\{([^{}]+)\}\}`)
	pathTemplate  = regexp.MustCompile(`\{([^{}]+)\}`)
)

// SplitRequestUrl cuts a request URL (without its query) into its origin ("{{baseUrl}}", "https://host") and path.
func SplitRequestUrl(raw string) (string, string) {
	base := ParseRequestUrl(strings.TrimSpace(raw)).base
	start := 0
	if strings.HasPrefix(base, "{{") {
		if end := strings.Index(base, "}}"); end >= 0 {
			start = end + len("}}")
		}
	} else if index := strings.Index(base, schemeSeparator); index >= 0 {
		start = index + len(schemeSeparator)
	}
	slash := strings.Index(base[start:], "/")
	if slash < 0 {
		return base, "/"
	}
	return base[:start+slash], base[start+slash:]
}

// DocumentedPath turns ":id" segments and "{{variables}}" into OpenAPI "{id}" templates with their parameters.
func DocumentedPath(path string, variables []PathVariable) (string, []ApiParameter) {
	parameters := make([]ApiParameter, 0)
	segments := strings.Split(path, "/")
	for index, segment := range segments {
		if name, isVariable := strings.CutPrefix(segment, pathVariableTag); isVariable && name != "" {
			segments[index] = "{" + name + "}"
			parameters = appendParameter(parameters, pathVariableParameter(name, variables))
			continue
		}
		segments[index] = xenonVariable.ReplaceAllStringFunc(segment, func(match string) string {
			name := xenonVariable.FindStringSubmatch(match)[1]
			parameters = appendParameter(parameters, ApiParameter{Name: name, Example: match, Required: true, Variable: true})
			return "{" + name + "}"
		})
	}
	return strings.Join(segments, "/"), parameters
}

// RequestPath turns an OpenAPI path back into a Xenon one: "{{name}}" for variables, ":name" for path variables.
func RequestPath(path string, parameters []ApiParameter) string {
	return pathTemplate.ReplaceAllStringFunc(path, func(match string) string {
		name := pathTemplate.FindStringSubmatch(match)[1]
		if slices.ContainsFunc(parameters, func(parameter ApiParameter) bool { return parameter.Name == name && parameter.Variable }) {
			return "{{" + name + "}}"
		}
		return pathVariableTag + name
	})
}

// commonPathPrefix is the run of plain leading segments every path shares before its last segment.
func commonPathPrefix(paths []string) string {
	var common []string
	for index, path := range paths {
		segments := strings.Split(strings.Trim(path, "/"), "/")
		candidates := segments[:max(len(segments)-1, 0)]
		if index == 0 {
			common = candidates
			continue
		}
		common = common[:sharedLength(common, candidates)]
	}
	prefix := ""
	for _, segment := range common {
		if segment == "" || strings.ContainsAny(segment, ":{}") {
			break
		}
		prefix += "/" + segment
	}
	return prefix
}

func sharedLength(left, right []string) int {
	length := 0
	for length < len(left) && length < len(right) && left[length] == right[length] {
		length++
	}
	return length
}

func pathVariableParameter(name string, variables []PathVariable) ApiParameter {
	parameter := ApiParameter{Name: name, Required: true}
	for _, variable := range variables {
		if variable.HasKey(name) {
			parameter.Description, parameter.Example = variable.Description, variable.Value
		}
	}
	return parameter
}

func appendParameter(parameters []ApiParameter, parameter ApiParameter) []ApiParameter {
	if slices.ContainsFunc(parameters, func(existing ApiParameter) bool { return existing.Name == parameter.Name }) {
		return parameters
	}
	return append(parameters, parameter)
}

func trimPathPrefix(path, prefix string) string {
	trimmed := strings.TrimPrefix(path, prefix)
	if trimmed == "" {
		return "/"
	}
	return trimmed
}
