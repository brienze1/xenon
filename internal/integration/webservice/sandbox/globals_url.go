package sandbox

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dop251/goja"
)

const (
	zoneInfoMarker = "zoneinfo/"
	localTimeLink  = "/etc/localtime"
	defaultZone    = "UTC"
)

func (e *execution) domainToASCII(call goja.FunctionCall) goja.Value {
	toASCII, _ := goja.AssertFunction(e.require("punycode").Get("toASCII"))
	converted, err := toASCII(goja.Undefined(), call.Argument(0))
	if err != nil {
		return goja.Null()
	}
	return converted
}

func (e *execution) timeZone(goja.FunctionCall) goja.Value {
	return e.vm.ToValue(localZoneName())
}

func localZoneName() string {
	if zone := os.Getenv("TZ"); zone != "" {
		if _, err := time.LoadLocation(zone); err == nil {
			return zone
		}
	}
	if name := time.Local.String(); name != "" && name != "Local" {
		return name
	}
	if target, err := filepath.EvalSymlinks(localTimeLink); err == nil {
		if _, zone, found := strings.Cut(filepath.ToSlash(target), zoneInfoMarker); found {
			return zone
		}
	}
	return defaultZone
}
