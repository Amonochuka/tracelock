package access

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type Hub struct {
	clients   map[*websocket.Conn]bool
	broadcast chan interface{}
	mu        sync.Mutex
	allowedOrigin string
}

func NewHub() *Hub {
	return &Hub{
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan any, 256),
	}
}

// Run listens for payloads and sends them to all connected clients.
// Must be started as a goroutine: go hub.Run()
func (h *Hub) Run() {
	for {
		payload := <-h.broadcast
		h.mu.Lock()
		for client := range h.clients {
			err := client.WriteJSON(payload)
			if err != nil {
				// client disconnected — remove and close
				client.Close()
				delete(h.clients, client)
			}
		}
		h.mu.Unlock()
	}
}

// BroadcastPayload sends a payload to all connected WebSocket clients.
func (h *Hub) BroadcastPayload(payload any) {
	h.broadcast <- payload
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // tighten this for production
	},
}

// HandleWebSocket upgrades the HTTP connection to WebSocket and registers the client.
func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade failed: %v", err)
		return
	}

	// register client
	h.mu.Lock()
	h.clients[conn] = true
	h.mu.Unlock()

	// keep alive — detect disconnections
	go func() {
		defer func() {
			h.mu.Lock()
			delete(h.clients, conn)
			conn.Close()
			h.mu.Unlock()
		}()
		for {
			if _, _, err := conn.NextReader(); err != nil {
				break
			}
		}
	}()
}
