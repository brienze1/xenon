package sandbox

import (
	"encoding/json"
	"errors"

	"github.com/dop251/goja"

	"github.com/brienze1/xenon/internal/integration/webservice/dto"
	"github.com/brienze1/xenon/pkg/errs"
)

const (
	jarSet   = "set"
	jarUnset = "unset"
	jarClear = "clear"
)

func (e *execution) cookieJar(call goja.FunctionCall) goja.Value {
	if e.cookies == nil {
		panic(e.newError("pm.cookies.jar is not available in this run"))
	}
	result, err := e.jarOperation(call.Argument(0).String(), call.Argument(1).String(), call.Argument(2).String())
	if err != nil {
		panic(e.newError(errorDescription(err)))
	}
	encoded, _ := json.Marshal(result)
	return e.vm.ToValue(string(encoded))
}

func (e *execution) jarOperation(operation, url, argument string) (any, error) {
	switch operation {
	case jarSet:
		cookie, err := e.cookies.SetCookie(e.ctx, url, argument)
		if err != nil {
			return nil, err
		}
		return dto.NewScriptSandboxJarCookie(*cookie), nil
	case jarUnset:
		return nil, e.cookies.Unset(e.ctx, url, argument)
	case jarClear:
		return nil, e.cookies.Clear(e.ctx, url)
	default:
		cookies, err := e.cookies.Cookies(e.ctx, url)
		if err != nil {
			return nil, err
		}
		return dto.NewScriptSandboxJarCookies(cookies), nil
	}
}

func errorDescription(err error) string {
	var known errs.BaseError
	if errors.As(err, &known) {
		return known.Description()
	}
	return err.Error()
}
