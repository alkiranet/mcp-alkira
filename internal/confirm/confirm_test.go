package confirm

import (
	"bytes"
	"encoding/base64"
	"errors"
	"testing"
	"time"
)

func TestCode_CarriesNoRawStateHash(t *testing.T) {
	s, err := NewSigner(RandomKey())
	if err != nil {
		t.Fatal(err)
	}
	state, err := StateHash([]byte(`{"name":"w","psk":"s3cr3t"}`), nil)
	if err != nil {
		t.Fatal(err)
	}
	code, _ := s.Issue("t", "widget", "7", state)
	buf, _ := base64.RawURLEncoding.DecodeString(code)
	if bytes.Contains(buf, state[:]) {
		t.Fatal("code carries the raw state hash")
	}
	if err := s.CheckSignature(code, "t", "widget", "7"); err != nil {
		t.Fatalf("CheckSignature: %v", err)
	}
	if err := s.Check(code, "t", "widget", "7", state); err != nil {
		t.Fatalf("Check: %v", err)
	}
	other, _ := StateHash([]byte(`{"name":"w","psk":"other"}`), nil)
	if err := s.Check(code, "t", "widget", "7", other); !errors.Is(err, ErrStale) {
		t.Fatalf("Check on changed state = %v, want ErrStale", err)
	}
	otherSigner, _ := NewSigner(RandomKey())
	if err := otherSigner.Check(code, "t", "widget", "7", state); !errors.Is(err, ErrMismatch) {
		t.Fatalf("Check with another key = %v, want ErrMismatch", err)
	}
}

// A missing or short key is refused; a 32-byte key is accepted.
func TestNewSigner_RefusesShortKey(t *testing.T) {
	for _, key := range [][]byte{nil, {}, []byte("k"), bytes.Repeat([]byte("k"), MinKeyLen-1)} {
		if s, err := NewSigner(key); !errors.Is(err, ErrShortKey) || s != nil {
			t.Errorf("NewSigner(%d bytes) = %v, %v; want nil, ErrShortKey", len(key), s, err)
		}
	}
	if _, err := NewSigner(bytes.Repeat([]byte("k"), MinKeyLen)); err != nil {
		t.Fatalf("32-byte key: %v", err)
	}
	if _, err := NewSigner(RandomKey()); err != nil {
		t.Fatalf("RandomKey: %v", err)
	}
}

// WithClock sets the clock that issues and expires codes.
func TestWithClock_Expiry(t *testing.T) {
	now := time.Date(2026, 11, 19, 12, 0, 0, 0, time.UTC)
	s, err := NewSigner(RandomKey(), WithClock(func() time.Time { return now }))
	if err != nil {
		t.Fatal(err)
	}
	code, exp := s.Issue("t", "widget", "7", [hashLen]byte{})
	if want := now.Add(Lifetime); !exp.Equal(want) {
		t.Fatalf("expiry = %s, want %s", exp, want)
	}
	now = now.Add(Lifetime - time.Second)
	if err := s.CheckSignature(code, "t", "widget", "7"); err != nil {
		t.Fatalf("before expiry: %v", err)
	}
	now = now.Add(time.Second)
	if err := s.CheckSignature(code, "t", "widget", "7"); !errors.Is(err, ErrExpired) {
		t.Fatalf("at expiry: %v, want ErrExpired", err)
	}
}
