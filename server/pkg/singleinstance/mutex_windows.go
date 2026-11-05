//go:build windows

package singleinstance

import (
	"errors"
	"fmt"
	"syscall"

	"golang.org/x/sys/windows"
)

type Mutex struct {
	handle windows.Handle
}

// Acquire attempts to create or open a Windows named mutex.
// Returns (mutex, isAlreadyRunning, error).
func Acquire(name string) (*Mutex, bool, error) {
	namePtr, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return nil, false, fmt.Errorf("invalid mutex name: %w", err)
	}

	handle, err := windows.CreateMutex(nil, false, namePtr)
	if err != nil {
		if errors.Is(err, windows.ERROR_ALREADY_EXISTS) {
			return &Mutex{handle: handle}, true, nil
		}
		return nil, false, fmt.Errorf("failed to create named mutex: %w", err)
	}

	return &Mutex{handle: handle}, false, nil
}

// Release closes the Windows mutex handle.
func (m *Mutex) Release() {
	if m != nil && m.handle != 0 {
		_ = windows.CloseHandle(m.handle)
		m.handle = 0
	}
}
