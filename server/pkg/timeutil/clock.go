package timeutil

import "time"

// Clock provides an interface for getting the current time.
// This is essential for testing time-sensitive code like TOTP verification.
type Clock interface {
	Now() time.Time
}

// RealClock implements Clock using the system time.
type RealClock struct{}

func (RealClock) Now() time.Time {
	return time.Now()
}
