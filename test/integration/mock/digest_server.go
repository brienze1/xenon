package mock

import (
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
)

const digestNonce = "7ypf/xlj9XXwfDPEoM4URrv/xwf94BcCAzFZH4GiTo0v"

type DigestServerConfig struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	Realm     string `json:"realm"`
	Algorithm string `json:"algorithm"`
	Qop       string `json:"qop"`
	Opaque    string `json:"opaque"`
}

type DigestServer struct {
	server         *httptest.Server
	config         DigestServerConfig
	authorizations []string
	lastCount      int64
	requests       int
	accepted       int
	mu             sync.Mutex
}

var digestServerOnce sync.Once
var digestServerInstance *DigestServer

func NewDigestServer() *DigestServer {
	digestServerOnce.Do(func() {
		digestServerInstance = &DigestServer{}
		digestServerInstance.server = httptest.NewServer(http.HandlerFunc(digestServerInstance.handle))
		digestServerInstance.Reset()
	})
	return digestServerInstance
}

func (d *DigestServer) URL() string {
	return d.server.URL
}

func (d *DigestServer) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.config = DigestServerConfig{Username: "ana", Password: "s3cret", Realm: "xenon@test", Algorithm: "MD5", Qop: "auth", Opaque: "FQhe/qaU925kfnzjCev0ciny7QMkPqMAFRtzCUYo5tdS"}
	d.authorizations = nil
	d.lastCount = 0
	d.requests = 0
	d.accepted = 0
}

func (d *DigestServer) Configure(raw string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	return json.Unmarshal([]byte(raw), &d.config)
}

func (d *DigestServer) Counts() (int, int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.requests, d.accepted
}

func (d *DigestServer) Authorization(index int) (string, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if index < 1 || index > len(d.authorizations) {
		return "", false
	}
	return d.authorizations[index-1], true
}

func (d *DigestServer) handle(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	authorization := r.Header.Get("Authorization")
	d.mu.Lock()
	defer d.mu.Unlock()
	d.requests++
	d.authorizations = append(d.authorizations, authorization)
	if d.verify(r.Method, r.URL.RequestURI(), authorization, body) {
		d.accepted++
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"authenticated":true,"user":%q}`, d.config.Username)
		return
	}
	challenge := fmt.Sprintf(`Digest realm=%q, nonce=%q, opaque=%q, algorithm=%s`, d.config.Realm, digestNonce, d.config.Opaque, d.config.Algorithm)
	if d.config.Qop != "" {
		challenge += fmt.Sprintf(`, qop=%q`, d.config.Qop)
	}
	w.Header().Set("WWW-Authenticate", challenge)
	w.WriteHeader(http.StatusUnauthorized)
}

func (d *DigestServer) verify(method, uri, authorization string, body []byte) bool {
	scheme, raw, found := strings.Cut(authorization, " ")
	if !found || !strings.EqualFold(scheme, "Digest") {
		return false
	}
	params := digestParams(raw)
	if params["username"] != d.config.Username || params["realm"] != d.config.Realm || params["nonce"] != digestNonce || params["uri"] != uri || params["opaque"] != d.config.Opaque || !strings.EqualFold(params["algorithm"], d.config.Algorithm) {
		return false
	}
	newHash := digestHash(d.config.Algorithm)
	ha1 := digestHex(newHash, d.config.Username+":"+d.config.Realm+":"+d.config.Password)
	if strings.HasSuffix(strings.ToLower(d.config.Algorithm), "-sess") {
		ha1 = digestHex(newHash, ha1+":"+digestNonce+":"+params["cnonce"])
	}
	a2 := method + ":" + uri
	if d.config.Qop == "auth-int" {
		a2 += ":" + digestHex(newHash, string(body))
	}
	ha2 := digestHex(newHash, a2)
	if d.config.Qop == "" {
		return params["response"] == digestHex(newHash, ha1+":"+digestNonce+":"+ha2)
	}
	count, err := strconv.ParseInt(params["nc"], 16, 64)
	if err != nil || len(params["nc"]) != 8 || count <= d.lastCount || params["qop"] != d.config.Qop {
		return false
	}
	if params["response"] != digestHex(newHash, ha1+":"+digestNonce+":"+params["nc"]+":"+params["cnonce"]+":"+params["qop"]+":"+ha2) {
		return false
	}
	d.lastCount = count
	return true
}

func digestParams(raw string) map[string]string {
	params := map[string]string{}
	for _, part := range splitDigestParams(raw) {
		key, value, found := strings.Cut(strings.TrimSpace(part), "=")
		if found {
			params[strings.ToLower(key)] = strings.Trim(value, `"`)
		}
	}
	return params
}

func splitDigestParams(raw string) []string {
	parts := make([]string, 0)
	quoted, start := false, 0
	for index, char := range raw {
		switch {
		case char == '"':
			quoted = !quoted
		case char == ',' && !quoted:
			parts = append(parts, raw[start:index])
			start = index + 1
		}
	}
	return append(parts, raw[start:])
}

func digestHash(algorithm string) func() hash.Hash {
	switch strings.TrimSuffix(strings.ToUpper(algorithm), "-SESS") {
	case "SHA-256":
		return sha256.New
	case "SHA-512-256":
		return sha512.New512_256
	default:
		return md5.New
	}
}

func digestHex(newHash func() hash.Hash, text string) string {
	sum := newHash()
	sum.Write([]byte(text))
	return hex.EncodeToString(sum.Sum(nil))
}
