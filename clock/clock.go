// Package clock adapts the system clock.
package clock

import "time"

// System is the system clock.
type System struct{}

// Now returns the current time.
func (System) Now() time.Time {
	return time.Now()
}
