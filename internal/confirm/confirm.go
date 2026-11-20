// Package confirm issues and checks the one-time codes of the two-step delete.
//
// A code carries its expiry and a keyed digest of the state it was issued for,
// signed with HMAC-SHA256 over tenant, family, object id, digest and expiry.
// The digest is an HMAC, so a code reveals nothing that could be checked
// offline against guesses of the object's secret fields. No server
// state is kept: the second call re-reads the object, and a code is void once
// the object it was issued for has changed or is gone.
package confirm

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"time"
)

// Lifetime is how long a code stays valid.
const Lifetime = 5 * time.Minute

var (
	// ErrMismatch: the code is malformed, tampered, or was issued for another tenant, family or id.
	ErrMismatch = errors.New("confirmation code does not match this object")
	// ErrExpired: the code is genuine but past its expiry.
	ErrExpired = errors.New("confirmation code expired")
	// ErrStale: the code is genuine but the object changed since it was issued.
	ErrStale = errors.New("the object changed since the preview")
)

const (
	hashLen = sha256.Size
	macLen  = sha256.Size
	codeLen = 8 + hashLen + macLen
)

// Signer issues and checks codes with one key.
type Signer struct {
	key []byte
	now func() time.Time
}

// Option configures NewSigner.
type Option func(*Signer)

// WithClock makes the signer read now instead of the wall clock for code
// expiry.
func WithClock(now func() time.Time) Option {
	return func(s *Signer) { s.now = now }
}

// MinKeyLen is the shortest key NewSigner accepts: the HMAC-SHA256 output size.
const MinKeyLen = 32

// ErrShortKey: the signing key is missing or shorter than MinKeyLen.
var ErrShortKey = errors.New("confirm: signing key shorter than 32 bytes")

// NewSigner returns a Signer for key, or ErrShortKey for a key shorter than
// MinKeyLen. A local process draws its own key with RandomKey, so its codes
// do not survive a restart; replicas that share a key accept each other's codes.
func NewSigner(key []byte, opts ...Option) (*Signer, error) {
	if len(key) < MinKeyLen {
		return nil, ErrShortKey
	}
	s := &Signer{key: append([]byte(nil), key...), now: time.Now}
	for _, o := range opts {
		o(s)
	}
	return s, nil
}

// RandomKey returns a fresh 32-byte key.
func RandomKey() []byte {
	k := make([]byte, 32)
	if _, err := rand.Read(k); err != nil {
		panic("confirm: no random source: " + err.Error())
	}
	return k
}

// Issue returns a code for the object and its expiry.
func (s *Signer) Issue(tenant, family, id string, state [hashLen]byte) (string, time.Time) {
	exp := s.now().Add(Lifetime).Truncate(time.Second)
	buf := make([]byte, 0, codeLen)
	buf = binary.BigEndian.AppendUint64(buf, uint64(exp.Unix()))
	digest := s.digest(state)
	buf = append(buf, digest...)
	buf = append(buf, s.mac(tenant, family, id, digest, exp.Unix())...)
	return base64.RawURLEncoding.EncodeToString(buf), exp
}

// Check verifies code for the object in its current state. The signature is
// checked first, so an altered expiry or digest reads as ErrMismatch.
func (s *Signer) Check(code, tenant, family, id string, current [hashLen]byte) error {
	digest, err := s.verify(code, tenant, family, id)
	if err != nil {
		return err
	}
	if !hmac.Equal(digest, s.digest(current)) {
		return ErrStale
	}
	return nil
}

// CheckSignature verifies the signature and expiry of code without the
// object's current state, so a bad code is refused before any request.
func (s *Signer) CheckSignature(code, tenant, family, id string) error {
	_, err := s.verify(code, tenant, family, id)
	return err
}

// verify checks the signature and expiry of code and returns its state digest.
func (s *Signer) verify(code, tenant, family, id string) ([]byte, error) {
	buf, err := base64.RawURLEncoding.DecodeString(code)
	if err != nil || len(buf) != codeLen {
		return nil, ErrMismatch
	}
	exp := int64(binary.BigEndian.Uint64(buf[:8]))
	digest := buf[8 : 8+hashLen]
	if !hmac.Equal(buf[8+hashLen:], s.mac(tenant, family, id, digest, exp)) {
		return nil, ErrMismatch
	}
	if !s.now().Before(time.Unix(exp, 0)) {
		return nil, ErrExpired
	}
	return digest, nil
}

func (s *Signer) digest(state [hashLen]byte) []byte {
	m := hmac.New(sha256.New, s.key)
	m.Write([]byte("state"))
	m.Write(state[:])
	return m.Sum(nil)
}

func (s *Signer) mac(tenant, family, id string, state []byte, exp int64) []byte {
	m := hmac.New(sha256.New, s.key)
	// Length-prefixed fields keep ("ab","c") and ("a","bc") apart.
	for _, f := range [][]byte{[]byte("delete"), []byte(tenant), []byte(family), []byte(id), state} {
		m.Write(binary.BigEndian.AppendUint32(nil, uint32(len(f))))
		m.Write(f)
	}
	m.Write(binary.BigEndian.AppendUint64(nil, uint64(exp)))
	return m.Sum(nil)
}

// StateHash hashes a JSON object body in canonical form (sorted keys, numbers
// as written) without the top-level fields in volatile.
func StateHash(body []byte, volatile []string) ([hashLen]byte, error) {
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return [hashLen]byte{}, err
	}
	if m, ok := v.(map[string]any); ok {
		for _, k := range volatile {
			delete(m, k)
		}
	}
	// encoding/json writes map keys sorted, which makes the encoding canonical.
	canon, err := json.Marshal(v)
	if err != nil {
		return [hashLen]byte{}, err
	}
	return sha256.Sum256(canon), nil
}
