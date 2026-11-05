package singleinstance

import (
	"runtime"
	"testing"
)

func TestSingleInstanceMutex(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Named mutex tests are specific to Windows")
	}

	mutexName := "Local\\NetraUptimeTestMutex_12345"

	// First acquire should succeed and indicate not running
	m1, alreadyRunning1, err1 := Acquire(mutexName)
	if err1 != nil {
		t.Fatalf("first acquire failed: %v", err1)
	}
	defer m1.Release()

	if alreadyRunning1 {
		t.Fatal("expected first acquire to report alreadyRunning=false")
	}

	// Second acquire on the same name should indicate already running
	m2, alreadyRunning2, err2 := Acquire(mutexName)
	if err2 != nil {
		t.Fatalf("second acquire failed: %v", err2)
	}
	defer m2.Release()

	if !alreadyRunning2 {
		t.Fatal("expected second acquire to report alreadyRunning=true")
	}
}
