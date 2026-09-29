package wormholescan

import "time"

// deref returns the value pointed to by p, or the zero value of T when p is nil.
func deref[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

// derefTime returns p in UTC, or the zero time when p is nil.
func derefTime(p *time.Time) time.Time {
	return deref(p).UTC()
}
