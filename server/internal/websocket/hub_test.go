package websocket

import (
	"testing"
)

func TestHub_EvictUserFromOrg_RemovesClientAndClosesSend(t *testing.T) {
	hub := NewHub()

	c1 := &Client{
		Hub:    hub,
		Send:   make(chan []byte, 256),
		UserID: "user-1",
		OrgIDs: []string{"org-1"},
	}
	c2 := &Client{
		Hub:    hub,
		Send:   make(chan []byte, 256),
		UserID: "user-2",
		OrgIDs: []string{"org-1"},
	}

	hub.Rooms["org-1"] = map[*Client]bool{
		c1: true,
		c2: true,
	}

	hub.EvictUserFromOrg("org-1", "user-1")

	// c1 should be evicted from the room
	if hub.Rooms["org-1"][c1] {
		t.Fatal("expected c1 to be removed from org-1 room")
	}

	// c1's send channel should be closed
	select {
	case _, ok := <-c1.Send:
		if ok {
			t.Fatal("expected c1.Send to be closed")
		}
	default:
		t.Fatal("expected c1.Send to receive closed signal")
	}

	// c2 should remain in room
	if !hub.Rooms["org-1"][c2] {
		t.Fatal("expected c2 to still be present in org-1 room")
	}

	// c2's send channel should still be open
	select {
	case <-c2.Send:
		t.Fatal("expected c2.Send to be open, not receiving data")
	default:
		// Send is still open and empty
	}
}

func TestHub_EvictUserFromOrg_NonExistentOrgOrUser(t *testing.T) {
	hub := NewHub()

	// Should not panic on non-existent room
	hub.EvictUserFromOrg("non-existent-org", "user-1")

	c1 := &Client{
		Hub:    hub,
		Send:   make(chan []byte, 256),
		UserID: "user-1",
		OrgIDs: []string{"org-1"},
	}
	hub.Rooms["org-1"] = map[*Client]bool{c1: true}

	// Should not panic on non-existent user in existing room
	hub.EvictUserFromOrg("org-1", "user-999")
	if !hub.Rooms["org-1"][c1] {
		t.Fatal("expected c1 to remain untouched")
	}

	// Nil hub should not panic
	var nilHub *Hub
	nilHub.EvictUserFromOrg("org-1", "user-1")
}

func TestWsHandler_EvictUserFromOrg(t *testing.T) {
	hub := NewHub()
	handler := NewWsHandler(hub, nil)

	c1 := &Client{
		Hub:    hub,
		Send:   make(chan []byte, 256),
		UserID: "user-1",
		OrgIDs: []string{"org-1"},
	}
	hub.Rooms["org-1"] = map[*Client]bool{c1: true}

	handler.EvictUserFromOrg("org-1", "user-1")

	if hub.Rooms["org-1"][c1] {
		t.Fatal("expected c1 to be evicted via handler")
	}

	var nilHandler *WsHandler
	nilHandler.EvictUserFromOrg("org-1", "user-1")
}
