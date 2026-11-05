//go:build !windows

package singleinstance

type Mutex struct{}

// Acquire is a no-op fallback on non-Windows platforms.
func Acquire(name string) (*Mutex, bool, error) {
	return &Mutex{}, false, nil
}

// Release is a no-op fallback on non-Windows platforms.
func (m *Mutex) Release() {}
