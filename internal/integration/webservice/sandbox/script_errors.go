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
	text := moreErrors.ReplaceAllString(match[3], "")
	return entity.NewScriptError(scripterrorkind.Syntax, "SyntaxError: "+text, script).At(userLine(match[1]), number(match[2]))
}

func runtimeError(err error, script entity.ScriptRef) entity.ScriptError {
	var overflow *goja.StackOverflowError
	if errors.As(err, &overflow) {
		return located(entity.NewScriptError(scripterrorkind.StackOverflow, stackOverflow, script), overflow.String())
	}
	var exception *goja.Exception
	if errors.As(err, &exception) {
		return thrownError(exception.Value(), script)
	}
	return entity.NewScriptError(scripterrorkind.Runtime, err.Error(), script)
}

func thrownError(value goja.Value, script entity.ScriptRef) entity.ScriptError {
	scriptError := entity.NewScriptError(scripterrorkind.Runtime, describe(value), script)
	if object, ok := value.(*goja.Object); ok {
		return located(scriptError, valueText(object.Get("stack")))
	}
	return scriptError
}

func rejectionError(value goja.Value, script entity.ScriptRef) entity.ScriptError {
	scriptError := thrownError(value, script)
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
