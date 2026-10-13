package sandbox

import (
	"encoding/base64"
	"strings"
	"unicode/utf8"

	"github.com/dop251/goja"
)

const forgivingBase64Alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

func (e *execution) atob(call goja.FunctionCall) goja.Value {
	text := strings.Map(func(r rune) rune {
		if r == ' ' || r == '\t' || r == '\n' || r == '\f' || r == '\r' {
			return -1
		}
		return r
	}, call.Argument(0).String())
	if len(text)%4 == 0 {
		text = strings.TrimSuffix(strings.TrimSuffix(text, "="), "=")
	}
	if len(text)%4 == 1 || strings.IndexFunc(text, func(r rune) bool { return !strings.ContainsRune(forgivingBase64Alphabet, r) }) >= 0 {
		return goja.Null()
	}
	decoded, err := base64.RawStdEncoding.DecodeString(text)
	if err != nil {
		return goja.Null()
	}
	return e.vm.ToValue(latin1(decoded))
}

func (e *execution) btoa(call goja.FunctionCall) goja.Value {
	text := []rune(call.Argument(0).String())
	bytes := make([]byte, 0, len(text))
	for _, character := range text {
		if character > 0xFF {
			return goja.Null()
		}
		bytes = append(bytes, byte(character))
	}
	return e.vm.ToValue(base64.StdEncoding.EncodeToString(bytes))
}

func (e *execution) encodeUtf8(call goja.FunctionCall) goja.Value {
	return e.vm.ToValue(e.vm.NewArrayBuffer([]byte(utf16Text(call.Argument(0).String()))))
}

func (e *execution) decodeUtf8(call goja.FunctionCall) goja.Value {
	bytes := e.bytesOf(call.Argument(0))
	if call.Argument(1).ToBoolean() && !utf8.Valid(bytes) {
		return goja.Null()
	}
	return e.vm.ToValue(strings.ToValidUTF8(string(bytes), "�"))
}

func (e *execution) decodeBytes(call goja.FunctionCall) goja.Value {
	return e.vm.ToValue(strings.ToValidUTF8(string(e.bytesOf(call.Argument(0))), "�"))
}

func (e *execution) decodeBinary(call goja.FunctionCall) goja.Value {
	runes := []rune(call.Argument(0).String())
	bytes := make([]byte, 0, len(runes))
	for _, character := range runes {
		bytes = append(bytes, byte(character))
	}
	return e.vm.ToValue(strings.ToValidUTF8(string(bytes), "�"))
}

func (e *execution) byteLength(call goja.FunctionCall) goja.Value {
	return e.vm.ToValue(len(utf16Text(call.Argument(0).String())))
}

func (e *execution) base64(call goja.FunctionCall) goja.Value {
	return e.vm.ToValue(base64.StdEncoding.EncodeToString([]byte(utf16Text(call.Argument(0).String()))))
}

func (e *execution) bytesOf(value goja.Value) []byte {
	if buffer, isBuffer := value.Export().(goja.ArrayBuffer); isBuffer {
		return buffer.Bytes()
	}
	return []byte(value.String())
}

func latin1(bytes []byte) string {
	runes := make([]rune, 0, len(bytes))
	for _, b := range bytes {
		runes = append(runes, rune(b))
	}
	return string(runes)
}

func utf16Text(text string) string {
	return strings.ToValidUTF8(text, "�")
}
