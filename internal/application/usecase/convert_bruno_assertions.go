package usecase

import (
	"regexp"
	"slices"
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type chaiCheck func(value string) (string, bool)

var brunoAssertionChecks = map[string]chaiCheck{
	"eq":          literalCheck(".to.eql(%s)"),
	"neq":         literalCheck(".to.not.eql(%s)"),
	"gt":          literalCheck(".to.be.above(%s)"),
	"gte":         literalCheck(".to.be.at.least(%s)"),
	"lt":          literalCheck(".to.be.below(%s)"),
	"lte":         literalCheck(".to.be.at.most(%s)"),
	"in":          listCheck(".to.be.oneOf(%s)"),
	"notIn":       listCheck(".to.not.be.oneOf(%s)"),
	"contains":    literalCheck(".to.include(%s)"),
	"notContains": literalCheck(".to.not.include(%s)"),
	"length":      literalCheck(".to.have.lengthOf(%s)"),
	"matches":     patternCheck(".to.match(/%s/)", "", ""),
	"notMatches":  patternCheck(".to.not.match(/%s/)", "", ""),
	"startsWith":  patternCheck(".to.match(/%s/)", "^", ""),
	"endsWith":    patternCheck(".to.match(/%s/)", "", "$"),
	"between":     betweenCheck,
	"isEmpty":     fixedCheck(".to.be.empty"),
	"isNotEmpty":  fixedCheck(".to.not.be.empty"),
	"isNull":      fixedCheck(".to.be.null"),
	"isUndefined": fixedCheck(".to.be.undefined"),
	"isDefined":   fixedCheck(".to.not.be.undefined"),
	"isTruthy":    fixedCheck(".to.be.ok"),
	"isFalsy":     fixedCheck(".to.not.be.ok"),
	"isNumber":    fixedCheck(`.to.be.a("number")`),
	"isString":    fixedCheck(`.to.be.a("string")`),
	"isBoolean":   fixedCheck(`.to.be.a("boolean")`),
	"isArray":     fixedCheck(`.to.be.a("array")`),
	"isJson":      fixedCheck(`.to.be.an("object")`),
}

var (
	bareLiterals  = []string{"true", "false", "null", "undefined"}
	numberLiteral = regexp.MustCompile(`^-?(\d+\.?\d*|\.\d+)([eE][-+]?\d+)?$`)
)

func ConvertBrunoAssertions(assertions []entity.BrunoAssertion) entity.ScriptTranslation {
	translation := entity.ScriptTranslation{Lines: []string{}, Untranslated: []entity.UntranslatedLine{}}
	for _, assertion := range assertions {
		line, converted := assertionLine(assertion)
		if assertion.Enabled && !converted {
			translation.Lines = append(translation.Lines, untranslatedPrefix+assertion.Source())
			translation.Untranslated = append(translation.Untranslated, entity.UntranslatedLine{Line: assertion.Line, Source: assertion.Source()})
			continue
		}
		translation.Lines = append(translation.Lines, enabledLine(line, assertion.Enabled))
	}
	return translation
}

func assertionLine(assertion entity.BrunoAssertion) (string, bool) {
	check, known := brunoAssertionChecks[assertion.Operator]
	if !known {
		return assertion.Source(), false
	}
	chai, valid := check(strings.TrimSpace(assertion.Value))
	expression, translated := translateBrunoExpression(assertion.Expression)
	line := "pm.test(" + jsString(assertion.Title()) + ", () => pm.expect(" + expression + ")" + chai + ");"
	return line, valid && translated
}

func literalCheck(format string) chaiCheck {
	return func(value string) (string, bool) {
		return strings.Replace(format, "%s", jsLiteral(value), 1), value != ""
	}
}

func listCheck(format string) chaiCheck {
	return func(value string) (string, bool) {
		if strings.HasPrefix(value, "[") {
			return strings.Replace(format, "%s", value, 1), true
		}
		items := make([]string, 0)
		for _, item := range strings.Split(value, ",") {
			items = append(items, jsLiteral(strings.TrimSpace(item)))
		}
		return strings.Replace(format, "%s", "["+strings.Join(items, ", ")+"]", 1), value != ""
	}
}

func patternCheck(format, prefix, suffix string) chaiCheck {
	return func(value string) (string, bool) {
		pattern := unquoted(value)
		if prefix != "" || suffix != "" {
			pattern = prefix + regexp.QuoteMeta(pattern) + suffix
		}
		return strings.Replace(format, "%s", strings.ReplaceAll(pattern, "/", `\/`), 1), value != ""
	}
}

func betweenCheck(value string) (string, bool) {
	bounds := strings.Split(value, ",")
	if len(bounds) != 2 {
		return "", false
	}
	return ".to.be.within(" + jsLiteral(strings.TrimSpace(bounds[0])) + ", " + jsLiteral(strings.TrimSpace(bounds[1])) + ")", true
}

func fixedCheck(chai string) chaiCheck {
	return func(string) (string, bool) {
		return chai, true
	}
}

func jsLiteral(value string) string {
	if numberLiteral.MatchString(value) || slices.Contains(bareLiterals, value) || isQuoted(value) {
		return value
	}
	return jsString(value)
}

func unquoted(value string) string {
	if isQuoted(value) {
		return value[1 : len(value)-1]
	}
	return value
}

func isQuoted(value string) bool {
	return len(value) >= 2 && (value[0] == '"' || value[0] == '\'') && value[len(value)-1] == value[0]
}

func jsString(value string) string {
	escaper := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`, "\r", `\r`, "\t", `\t`)
	return `"` + escaper.Replace(value) + `"`
}
