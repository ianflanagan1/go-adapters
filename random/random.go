// Package random adapts the system's sources of randomness.
package random

import (
	crand "crypto/rand"
	mrand "math/rand/v2"
)

// Fast is a fast source of random numbers, not for secrets.
type Fast struct{}

// Uint64 returns a random number.
func (Fast) Uint64() uint64 {
	return mrand.Uint64()
}

// Secure is the operating system's cryptographically secure source, for
// secrets, tokens and IDs.
type Secure struct{}

// Read fills b with random bytes. It never returns an error.
func (Secure) Read(b []byte) (int, error) {
	return crand.Read(b)
}
