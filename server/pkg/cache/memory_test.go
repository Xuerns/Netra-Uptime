package cache

import (
	"testing"
	"time"
)

func TestMemoryCache_SetGetDel(t *testing.T) {
	mc := NewMemoryCache()

	mc.Set("key1", "hello", 5*time.Second)
	val, ok := mc.Get("key1")
	if !ok || val != "hello" {
		t.Fatalf("expected ('hello', true), got (%q, %v)", val, ok)
	}

	mc.Del("key1")
	_, ok = mc.Get("key1")
	if ok {
		t.Fatal("expected key1 to be deleted")
	}
}

func TestMemoryCache_Expiration(t *testing.T) {
	mc := NewMemoryCache()

	mc.Set("exp_key", "temp", 20*time.Millisecond)
	val, ok := mc.Get("exp_key")
	if !ok || val != "temp" {
		t.Fatalf("expected ('temp', true), got (%q, %v)", val, ok)
	}

	time.Sleep(35 * time.Millisecond)
	_, ok = mc.Get("exp_key")
	if ok {
		t.Fatal("expected exp_key to be expired")
	}
}

func TestMemoryCache_Incr(t *testing.T) {
	mc := NewMemoryCache()

	c1 := mc.Incr("counter", 50*time.Millisecond)
	if c1 != 1 {
		t.Fatalf("expected count 1, got %d", c1)
	}

	c2 := mc.Incr("counter", 50*time.Millisecond)
	if c2 != 2 {
		t.Fatalf("expected count 2, got %d", c2)
	}

	val, ok := mc.Get("counter")
	if !ok || val != "2" {
		t.Fatalf("expected Get('counter') = ('2', true), got (%q, %v)", val, ok)
	}

	time.Sleep(65 * time.Millisecond)

	c3 := mc.Incr("counter", 50*time.Millisecond)
	if c3 != 1 {
		t.Fatalf("expected count to reset to 1 after TTL expiration, got %d", c3)
	}
}
