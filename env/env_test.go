package env_test

import (
	"testing"

	"github.com/ianflanagan1/go-adapters/env"
)

func TestLookupReadsTheEnvironment(t *testing.T) {
	t.Setenv("GO_ADAPTERS_TEST", "set")
	if got, ok := (env.System{}).Lookup("GO_ADAPTERS_TEST"); !ok || got != "set" {
		t.Fatalf("got %q, %v; want \"set\", true", got, ok)
	}
	if _, ok := (env.System{}).Lookup("GO_ADAPTERS_UNSET"); ok {
		t.Fatal("got an unset variable")
	}
}
