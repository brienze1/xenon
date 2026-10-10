package sandbox

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"hash"

	"github.com/dop251/goja"
)

var digests = map[string]func() hash.Hash{"SHA-1": sha1.New, "SHA-256": sha256.New, "SHA-384": sha512.New384, "SHA-512": sha512.New}

func (e *execution) randomUUID(goja.FunctionCall) goja.Value {
	return e.vm.ToValue(newUUID())
}

func newUUID() string {
	bytes := make([]byte, 16)
	_, _ = rand.Read(bytes)
	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80
	const hexDigits = "0123456789abcdef"
	text := make([]byte, 0, 36)
	for index, b := range bytes {
		if index == 4 || index == 6 || index == 8 || index == 10 {
			text = append(text, '-')
		}
		text = append(text, hexDigits[b>>4], hexDigits[b&0x0f])
	}
	return string(text)
}

func (e *execution) randomBytes(call goja.FunctionCall) goja.Value {
	bytes := make([]byte, call.Argument(0).ToInteger())
	_, _ = rand.Read(bytes)
	return e.vm.ToValue(e.vm.NewArrayBuffer(bytes))
}

func (e *execution) digest(call goja.FunctionCall) goja.Value {
	digest := digests[call.Argument(0).String()]()
	digest.Write(e.bytesOf(call.Argument(1)))
	return e.vm.ToValue(e.vm.NewArrayBuffer(digest.Sum(nil)))
}

func (e *execution) hmac(call goja.FunctionCall) goja.Value {
	return e.vm.ToValue(e.vm.NewArrayBuffer(e.sign(call.Argument(0).String(), e.bytesOf(call.Argument(1)), e.bytesOf(call.Argument(2)))))
}

func (e *execution) hmacEqual(call goja.FunctionCall) goja.Value {
	expected := e.sign(call.Argument(0).String(), e.bytesOf(call.Argument(1)), e.bytesOf(call.Argument(3)))
	return e.vm.ToValue(hmac.Equal(expected, e.bytesOf(call.Argument(2))))
}

func (e *execution) sign(algorithm string, key, data []byte) []byte {
	mac := hmac.New(digests[algorithm], key)
	mac.Write(data)
	return mac.Sum(nil)
}
