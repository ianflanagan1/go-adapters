// Package random adapts the system's source of randomness.
package random

import (
	crand "crypto/rand"
	"fmt"
)

// Secure is the operating system's cryptographically secure source, for
// secrets, tokens and IDs.
type Secure struct{}

// Read fills b with random bytes. It never returns an error.
func (Secure) Read(b []byte) (int, error) {
	n, err := crand.Read(b)
	if err != nil {
		return n, fmt.Errorf("reading random bytes: %w", err)
	}
	return n, nil
}
