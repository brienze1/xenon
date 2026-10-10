package sandbox

import (
	"errors"
	"regexp"
	"strconv"

	"github.com/dop251/goja"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/scripterrorkind"
)

const (
	scriptFile     = "script.js"
	wrapperPrefix  = ";(async()=>{\n"
	wrapperSuffix  = "\n})()"
	wrapperLines   = 1
	stackOverflow  = "RangeError: Maximum call stack size exceeded"
	timeoutMessage = "Script execution timed out"
	memoryMessage  = "Script stopped: it used too much memory"
	unhandled      = "Unhandled promise rejection: "
)

var (
	syntaxPosition = regexp.MustCompile(`Line (\d+):(\d+) (.*)$`)
	stackPosition  = regexp.MustCompile(regexp.QuoteMeta(scriptFile) + `:(\d+):(\d+)`)
	moreErrors     = regexp.MustCompile(` \(and \d+ more errors\)$`)
	bareToken      = regexp.MustCompile(`^Unexpected token ([^'\s][^\s]*)$`)
)

func wrap(text string) string {
	return wrapperPrefix + text + wrapperSuffix
}

func syntaxError(err error, script entity.ScriptRef) entity.ScriptError {
	var compileErr *goja.CompilerSyntaxError
	message := err.Error()
	if errors.As(err, &compileErr) {
		message = compileErr.Message
	}
	match := syntaxPosition.FindStringSubmatch(message)
	if match == nil {
		return entity.NewScriptError(scripterrorkind.Syntax, "SyntaxError: "+message, script)
	}
	text := bareToken.ReplaceAllString(moreErrors.ReplaceAllString(match[3], ""), "Unexpected token '$1'")
	return entity.NewScriptError(scripterrorkind.Syntax, "SyntaxError: "+text, script).At(userLine(match[1]), number(match[2]))
}

func (e *execution) runtimeError(err error) entity.ScriptError {
	var overflow *goja.StackOverflowError
	if errors.As(err, &overflow) {
		return located(entity.NewScriptError(scripterrorkind.StackOverflow, stackOverflow, e.script), overflow.String())
	}
	var exception *goja.Exception
	if errors.As(err, &exception) {
		return e.thrownError(exception.Value())
	}
	return entity.NewScriptError(scripterrorkind.Runtime, err.Error(), e.script)
}

func (e *execution) thrownError(value goja.Value) entity.ScriptError {
	kind := scripterrorkind.Runtime
	object, isObject := value.(*goja.Object)
	if isObject && e.unsupported[object] {
		kind = scripterrorkind.UnsupportedApi
	}
	scriptError := entity.NewScriptError(kind, v8Message(describe(value), e.script.Text), e.script)
	if isObject {
		return located(scriptError, valueText(object.Get("stack")))
	}
	return scriptError
}

func (e *execution) rejectionError(value goja.Value) entity.ScriptError {
	scriptError := e.thrownError(value)
	scriptError.Kind = scripterrorkind.UnhandledRejection
	return scriptError.WithMessage(unhandled + scriptError.Message)
}

func limitError(kind string, script entity.ScriptRef) entity.ScriptError {
	if kind == scripterrorkind.Memory.String() {
		return entity.NewScriptError(scripterrorkind.Memory, memoryMessage, script)
	}
	return entity.NewScriptError(scripterrorkind.Timeout, timeoutMessage, script)
}

func located(scriptError entity.ScriptError, stack string) entity.ScriptError {
	match := stackPosition.FindStringSubmatch(stack)
	if match == nil || userLine(match[1]) < 1 {
		return scriptError
	}
	return scriptError.At(userLine(match[1]), number(match[2]))
}

func describe(value goja.Value) string {
	if object, ok := value.(*goja.Object); ok && object.ClassName() == "Error" {
		return object.String()
	}
	return valueText(value)
}

func valueText(value goja.Value) string {
	if value == nil {
		return "undefined"
	}
	return value.String()
}

func userLine(raw string) int {
	return number(raw) - wrapperLines
}

func number(raw string) int {
	parsed, _ := strconv.Atoi(raw)
	return parsed
}
