package random_test

import (
	"bytes"
	"testing"

	"github.com/ianflanagan1/go-adapters/random"
)

func TestFastNumbersVary(t *testing.T) {
	seen := map[uint64]bool{}
	for range 4 {
		seen[random.Fast{}.Uint64()] = true
	}
	if len(seen) == 1 {
		t.Fatal("got the same number four times")
	}
}

func TestSecureFillsTheBuffer(t *testing.T) {
	b := make([]byte, 32)
	if n, err := (random.Secure{}).Read(b); err != nil || n != len(b) {
		t.Fatalf("got %d, %v; want %d, nil", n, err, len(b))
	}
	if bytes.Equal(b, make([]byte, 32)) {
		t.Fatal("got only zeros")
	}
}
