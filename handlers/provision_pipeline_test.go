package handlers

import (
	"bytes"
	"testing"
)

// The preview pipeline never signs, with or without a key.
func TestNewPreviewPipeline_NoSigner(t *testing.T) {
	key := bytes.Repeat([]byte("k"), 32)
	for _, opts := range [][]WriteOption{nil, {WithConfirmKey(key)}} {
		p, err := NewPreviewPipeline(opts...)
		if err != nil || p.confirm != nil {
			t.Errorf("%d options: signer %v, err %v", len(opts), p != nil && p.confirm != nil, err)
		}
	}
}
