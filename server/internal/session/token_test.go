package session

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestSealOpen(t *testing.T) {
	s, err := NewSealer(bytes.Repeat([]byte{1}, 32), time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	tok, err := s.Seal(map[string]string{"ticket": "T"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(tok, "T\"") {
		t.Fatal("token must not be readable")
	}
	var got map[string]string
	if err := s.Open(tok, &got); err != nil || got["ticket"] != "T" {
		t.Fatalf("open: %v %v", err, got)
	}

	tampered := []byte(tok)
	tampered[len(tampered)/2] ^= 1
	if err := s.Open(string(tampered), &got); !errors.Is(err, ErrInvalid) {
		t.Errorf("tampered: %v", err)
	}

	other, _ := NewSealer(bytes.Repeat([]byte{2}, 32), time.Hour)
	if err := other.Open(tok, &got); !errors.Is(err, ErrInvalid) {
		t.Errorf("wrong key: %v", err)
	}

	s.now = func() time.Time { return time.Now().Add(2 * time.Hour) }
	if err := s.Open(tok, &got); !errors.Is(err, ErrInvalid) {
		t.Errorf("expired: %v", err)
	}
}
