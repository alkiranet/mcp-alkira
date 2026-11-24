package handlers

import (
	"bytes"
	"testing"
)

// The preview pipeline signs only with a configured key and never draws a random one.
func TestNewPreviewPipeline_Key(t *testing.T) {
	key := bytes.Repeat([]byte("k"), 32)
	for _, c := range []struct {
		opts   []WriteOption
		signed bool
	}{
		{nil, false},
		{[]WriteOption{WithConfirmKey(key)}, true},
		{[]WriteOption{WithConfirmKey(key), WithoutConfirm()}, false},
	} {
		p, err := NewPreviewPipeline(c.opts...)
		if err != nil || (p.confirm != nil) != c.signed {
			t.Errorf("%d options: signer %v, err %v; want signer %v", len(c.opts), p.confirm != nil, err, c.signed)
		}
	}
}
