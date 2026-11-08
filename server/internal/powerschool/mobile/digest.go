package mobile

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"strings"
	"sync"
)

// digestAuth answers RFC 7616 HTTP Digest challenges with the app-level service
// credentials. The last challenge is reused so most calls need one round trip.
type digestAuth struct {
	user, pass string
	cnonce     func() string

	mu sync.Mutex
	ch *challenge
	nc uint32
}

type challenge struct {
	realm, nonce, opaque, algorithm string
	qopAuth                         bool
}

func newDigestAuth(user, pass string) *digestAuth {
	return &digestAuth{user: user, pass: pass, cnonce: randomCnonce}
}

func randomCnonce() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// setChallenge stores a WWW-Authenticate Digest challenge. It reports false when the
// header is not a usable Digest challenge.
func (d *digestAuth) setChallenge(header string) bool {
	scheme, params, ok := strings.Cut(strings.TrimSpace(header), " ")
	if !ok || !strings.EqualFold(scheme, "Digest") {
		return false
	}
	p := parseAuthParams(params)
	if p["nonce"] == "" {
		return false
	}
	ch := &challenge{realm: p["realm"], nonce: p["nonce"], opaque: p["opaque"], algorithm: p["algorithm"]}
	for _, q := range strings.Split(p["qop"], ",") {
		if strings.TrimSpace(q) == "auth" {
			ch.qopAuth = true
		}
	}
	if newHash(ch.algorithm) == nil {
		return false
	}
	d.mu.Lock()
	d.ch, d.nc = ch, 0
	d.mu.Unlock()
	return true
}

// authorization returns the Authorization header for a request, or "" before any challenge.
func (d *digestAuth) authorization(method, uri string) string {
	d.mu.Lock()
	ch := d.ch
	d.nc++
	nc := fmt.Sprintf("%08x", d.nc)
	d.mu.Unlock()
	if ch == nil {
		return ""
	}
	h := func(parts ...string) string {
		hh := newHash(ch.algorithm)
		hh.Write([]byte(strings.Join(parts, ":")))
		return hex.EncodeToString(hh.Sum(nil))
	}
	cnonce := d.cnonce()
	ha1 := h(d.user, ch.realm, d.pass)
	if strings.HasSuffix(strings.ToLower(ch.algorithm), "-sess") {
		ha1 = h(ha1, ch.nonce, cnonce)
	}
	ha2 := h(method, uri)

	var b strings.Builder
	fmt.Fprintf(&b, `Digest username=%q, realm=%q, nonce=%q, uri=%q`, d.user, ch.realm, ch.nonce, uri)
	if ch.qopAuth {
		fmt.Fprintf(&b, `, qop=auth, nc=%s, cnonce=%q, response=%q`, nc, cnonce, h(ha1, ch.nonce, nc, cnonce, "auth", ha2))
	} else {
		fmt.Fprintf(&b, `, response=%q`, h(ha1, ch.nonce, ha2))
	}
	if ch.algorithm != "" {
		fmt.Fprintf(&b, `, algorithm=%s`, ch.algorithm)
	}
	if ch.opaque != "" {
		fmt.Fprintf(&b, `, opaque=%q`, ch.opaque)
	}
	return b.String()
}

func newHash(algorithm string) hash.Hash {
	switch strings.ToUpper(strings.TrimSuffix(strings.ToLower(algorithm), "-sess")) {
	case "", "MD5":
		return md5.New()
	case "SHA-256":
		return sha256.New()
	}
	return nil
}

// parseAuthParams splits `k=v, k="quoted, value"` pairs.
func parseAuthParams(s string) map[string]string {
	out := map[string]string{}
	for len(s) > 0 {
		s = strings.TrimLeft(s, " ,")
		key, rest, ok := strings.Cut(s, "=")
		if !ok {
			break
		}
		key = strings.ToLower(strings.TrimSpace(key))
		var val string
		if strings.HasPrefix(rest, `"`) {
			end := 1
			for end < len(rest) && rest[end] != '"' {
				if rest[end] == '\\' {
					end++
				}
				end++
			}
			val = strings.ReplaceAll(rest[1:min(end, len(rest))], `\`, "")
			s = rest[min(end+1, len(rest)):]
		} else {
			val, s, _ = strings.Cut(rest, ",")
			val = strings.TrimSpace(val)
		}
		out[key] = val
	}
	return out
}
