// Package env adapts the process environment.
package env

import "os"

// System is the process environment.
type System struct{}

// Lookup returns the value of the variable named key, and whether it's set.
func (System) Lookup(key string) (string, bool) {
	return os.LookupEnv(key)
}
