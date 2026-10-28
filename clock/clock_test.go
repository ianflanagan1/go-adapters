package clock_test

import (
	"testing"
	"time"

	"github.com/ianflanagan1/go-adapters/clock"
)

func TestNowIsTheCurrentTime(t *testing.T) {
	before := time.Now()
	got := clock.System{}.Now()
	if got.Before(before) || got.After(time.Now()) {
		t.Fatalf("got %v, want a time during the call", got)
	}
}
