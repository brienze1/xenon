package curl

import (
	"strings"

	"github.com/brienze1/xenon/internal/integration/webservice/curl/dto"
)

type option struct {
	apply      func(command *dto.Command, value string)
	takesValue bool
}

var longOptions = map[string]option{
	"--request":        {takesValue: true, apply: func(c *dto.Command, v string) { c.Method = v }},
	"--header":         {takesValue: true, apply: func(c *dto.Command, v string) { c.AddHeader(v) }},
	"--data":           {takesValue: true, apply: addData},
	"--data-ascii":     {takesValue: true, apply: addData},
	"--data-raw":       {takesValue: true, apply: func(c *dto.Command, v string) { c.Data = append(c.Data, dto.NewDataPart(v, true)) }},
	"--data-binary":    {takesValue: true, apply: addData},
	"--data-urlencode": {takesValue: true, apply: func(c *dto.Command, v string) { c.Data = append(c.Data, dto.NewUrlEncodedPart(v)) }},
	"--form":           {takesValue: true, apply: func(c *dto.Command, v string) { c.Forms = append(c.Forms, v) }},
	"--form-string":    {takesValue: true, apply: func(c *dto.Command, v string) { c.Forms = append(c.Forms, v) }},
	"--get":            {apply: func(c *dto.Command, _ string) { c.Get = true }},
	"--head":           {apply: func(c *dto.Command, _ string) { c.Head = true }},
	"--upload-file":    {takesValue: true, apply: func(c *dto.Command, v string) { c.UploadFile = v }},
	"--url":            {takesValue: true, apply: func(c *dto.Command, v string) { c.Urls = append(c.Urls, v) }},
	"--user":           {takesValue: true, apply: func(c *dto.Command, v string) { c.User = &v }},
	"--basic":          {apply: func(c *dto.Command, _ string) { c.Digest = false }},
	"--digest":         {apply: func(c *dto.Command, _ string) { c.Digest = true }},
	"--cookie":         {takesValue: true, apply: addCookie},
	"--user-agent":     {takesValue: true, apply: func(c *dto.Command, v string) { c.SetHeader("User-Agent", v) }},
	"--referer":        {takesValue: true, apply: func(c *dto.Command, v string) { c.SetHeader("Referer", v) }},
	"--json":           {takesValue: true, apply: func(c *dto.Command, v string) { c.Json = append(c.Json, v) }},
	"--insecure":       {apply: func(c *dto.Command, _ string) { c.Insecure = true }},
	"--location":       {apply: func(c *dto.Command, _ string) { c.Location = true }},
}

var shortOptions = map[byte]string{
	'X': "--request", 'H': "--header", 'd': "--data", 'F': "--form", 'G': "--get", 'I': "--head",
	'T': "--upload-file", 'u': "--user", 'b': "--cookie", 'A': "--user-agent", 'e': "--referer",
	'k': "--insecure", 'L': "--location", 's': "--silent", 'S': "--show-error", 'v': "--verbose",
	'i': "--include", 'f': "--fail", 'g': "--globoff", 'N': "--no-buffer", '#': "--progress-bar",
	'q': "--disable", 'o': "--output", 'm': "--max-time", 'w': "--write-out", 'c': "--cookie-jar",
	'D': "--dump-header", 'C': "--continue-at", 'Y': "--speed-limit", 'y': "--speed-time",
	'0': "--http1.0", '1': "--tlsv1", '2': "--sslv2", '3': "--sslv3", '4': "--ipv4", '6': "--ipv6",
	'x': "--proxy", 'E': "--cert", 'U': "--proxy-user", 'r': "--range", 'z': "--time-cond",
}

var ignoredFlags = []string{
	"--compressed", "--silent", "--show-error", "--verbose", "--include", "--fail", "--fail-with-body",
	"--globoff", "--no-buffer", "--progress-bar", "--disable", "--http1.0", "--http1.1", "--http2",
	"--http2-prior-knowledge", "--http3", "--tlsv1", "--tlsv1.0", "--tlsv1.1", "--tlsv1.2", "--tlsv1.3",
	"--sslv2", "--sslv3", "--ipv4", "--ipv6", "--no-keepalive", "--raw", "--tr-encoding", "--path-as-is",
	"--no-progress-meter", "--remote-name", "--remote-header-name", "--create-dirs", "--ssl-reqd",
	"--styled-output", "--no-styled-output",
}

var ignoredOptions = []string{
	"--output", "--max-time", "--write-out", "--cookie-jar", "--dump-header", "--continue-at",
	"--speed-limit", "--speed-time", "--connect-timeout", "--retry", "--retry-delay", "--retry-max-time",
	"--limit-rate", "--max-filesize", "--trace", "--trace-ascii", "--stderr", "--keepalive-time",
	"--expect100-timeout", "--max-redirs", "--interface", "--resolve", "--connect-to", "--cacert",
	"--capath", "--cert", "--cert-type", "--key", "--key-type", "--pass", "--proxy", "--proxy-user",
	"--noproxy", "--range", "--time-cond", "--config", "--netrc-file", "--unix-socket", "--abstract-unix-socket",
}

func parseOptions(arguments []string) *dto.Command {
	command := dto.NewCommand()
	for position := 0; position < len(arguments); position++ {
		argument := arguments[position]
		switch {
		case strings.HasPrefix(argument, "--") && len(argument) > 2:
			position = applyLong(command, argument, arguments, position)
		case strings.HasPrefix(argument, "-") && len(argument) > 1:
			position = applyShort(command, argument, arguments, position)
		default:
			command.Urls = append(command.Urls, argument)
		}
	}
	return command
}

func applyLong(command *dto.Command, name string, arguments []string, position int) int {
	if known, found := longOptions[name]; found {
		value, next := optionValue(known.takesValue, "", arguments, position)
		known.apply(command, value)
		return next
	}
	if contains(ignoredOptions, name) {
		return position + 1
	}
	if !contains(ignoredFlags, name) {
		command.Notice("Ignored unknown option " + name)
	}
	return position
}

func applyShort(command *dto.Command, argument string, arguments []string, position int) int {
	for index := 1; index < len(argument); index++ {
		name, found := shortOptions[argument[index]]
		if !found {
			command.Notice("Ignored unknown option -" + string(argument[index]))
			continue
		}
		if known, isApplied := longOptions[name]; isApplied {
			value, next := optionValue(known.takesValue, argument[index+1:], arguments, position)
			known.apply(command, value)
			if known.takesValue {
				return next
			}
			continue
		}
		if contains(ignoredOptions, name) {
			_, next := optionValue(true, argument[index+1:], arguments, position)
			return next
		}
	}
	return position
}

func optionValue(takesValue bool, attached string, arguments []string, position int) (string, int) {
	switch {
	case !takesValue:
		return "", position
	case attached != "":
		return attached, position
	case position+1 < len(arguments):
		return arguments[position+1], position + 1
	default:
		return "", position
	}
}

func addData(command *dto.Command, value string) {
	command.Data = append(command.Data, dto.NewDataPart(value, false))
}

func addCookie(command *dto.Command, value string) {
	if !strings.Contains(value, "=") {
		command.Notice("Ignored the cookie file " + value)
		return
	}
	command.AddCookie(value)
}

func contains(names []string, name string) bool {
	for _, candidate := range names {
		if candidate == name {
			return true
		}
	}
	return false
}
