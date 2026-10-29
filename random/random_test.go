package random_test

import (
	"bytes"
	"testing"

	"github.com/ianflanagan1/go-adapters/random"
)

func TestSecureFillsTheBuffer(t *testing.T) {
	b := make([]byte, 32)
	if n, err := (random.Secure{}).Read(b); err != nil || n != len(b) {
		t.Fatalf("got %d, %v; want %d, nil", n, err, len(b))
	}
	if bytes.Equal(b, make([]byte, 32)) {
		t.Fatal("got only zeros")
	}
}
