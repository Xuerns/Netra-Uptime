package websocket

import (
	"encoding/json"
	"sync"
)

type Hub struct {
	// Map OrganizationID -> Set of Clients (Rooms)
	Rooms map[string]map[*Client]bool

	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan Message

	mu sync.RWMutex
}

type Message struct {
	Room    string
	Type    string
	Payload interface{}
}

func NewHub() *Hub {
	return &Hub{
		Rooms:      make(map[string]map[*Client]bool),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan Message),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			for _, orgID := range client.OrgIDs {
				if h.Rooms[orgID] == nil {
					h.Rooms[orgID] = make(map[*Client]bool)
				}
				h.Rooms[orgID][client] = true
			}
			h.mu.Unlock()
		case client := <-h.Unregister:
			h.mu.Lock()
			closed := false
			for _, orgID := range client.OrgIDs {
				if _, ok := h.Rooms[orgID][client]; ok {
					delete(h.Rooms[orgID], client)
					if !closed {
						close(client.Send)
						closed = true
					}
				}
			}
			h.mu.Unlock()
		case message := <-h.Broadcast:
			h.mu.Lock()
			roomClients := h.Rooms[message.Room]
			for client := range roomClients {
				msgBytes, err := json.Marshal(message)
				if err != nil {
					continue
				}
				select {
				case client.Send <- msgBytes:
				default:
					closed := false
					for _, orgID := range client.OrgIDs {
						if _, ok := h.Rooms[orgID][client]; ok {
							delete(h.Rooms[orgID], client)
							if !closed {
								close(client.Send)
								closed = true
							}
						}
					}
				}
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) BroadcastToOrg(orgID, msgType string, payload interface{}) {
	h.Broadcast <- Message{
		Room:    orgID,
		Type:    msgType,
		Payload: payload,
	}
}
