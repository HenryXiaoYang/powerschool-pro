// Package session seals upstream credentials into an opaque bearer token so the server
// stays stateless: the browser holds the token, only the server can read or forge it.
package session

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"
)

// ErrInvalid covers malformed, forged, and expired tokens alike.
var ErrInvalid = errors.New("session: invalid token")

type Sealer struct {
	aead   cipher.AEAD
	maxAge time.Duration
	now    func() time.Time
}

// NewSealer takes a 32-byte AES-256 key. Tokens older than maxAge are rejected even if
// PowerSchool would still accept the ticket inside.
func NewSealer(key []byte, maxAge time.Duration) (*Sealer, error) {
	if len(key) != 32 {
		return nil, errors.New("session: key must be 32 bytes")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &Sealer{aead: aead, maxAge: maxAge, now: time.Now}, nil
}

type envelope struct {
	IssuedAt int64           `json:"iat"`
	Data     json.RawMessage `json:"d"`
}

func (s *Sealer) Seal(v any) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	plain, _ := json.Marshal(envelope{IssuedAt: s.now().Unix(), Data: data})
	nonce := make([]byte, s.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(s.aead.Seal(nonce, nonce, plain, nil)), nil
}

func (s *Sealer) Open(token string, v any) error {
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || len(raw) < s.aead.NonceSize() {
		return ErrInvalid
	}
	n := s.aead.NonceSize()
	plain, err := s.aead.Open(nil, raw[:n], raw[n:], nil)
	if err != nil {
		return ErrInvalid
	}
	var env envelope
	if json.Unmarshal(plain, &env) != nil {
		return ErrInvalid
	}
	if age := s.now().Sub(time.Unix(env.IssuedAt, 0)); age < 0 || age > s.maxAge {
		return ErrInvalid
	}
	if json.Unmarshal(env.Data, v) != nil {
		return ErrInvalid
	}
	return nil
}
