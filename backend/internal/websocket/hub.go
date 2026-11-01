package websocket

import (
	"encoding/json"
	"log"
	"sort"
	"sync"
	"taskflow/internal/models"

	"github.com/gorilla/websocket"
)

const sendQueueSize = 256

type Client struct {
	ID      string
	UserID  string
	BoardID string
	Conn    *websocket.Conn
	Send    chan []byte
}

type Hub struct {
	mu     sync.RWMutex
	boards map[string]map[string]*Client
}

func NewHub() *Hub {
	return &Hub{
		boards: make(map[string]map[string]*Client),
	}
}

// Register adds a client synchronously.
// Presence is immediately visible after this method returns.
func (h *Hub) Register(client *Client) {
	if client == nil || client.ID == "" || client.BoardID == "" {
		return
	}

	if client.Send == nil {
		client.Send = make(chan []byte, sendQueueSize)
	}

	var replaced *Client

	h.mu.Lock()

	clients, exists := h.boards[client.BoardID]
	if !exists {
		clients = make(map[string]*Client)
		h.boards[client.BoardID] = clients
	}

	if old, exists := clients[client.ID]; exists {
		if old == client {
			h.mu.Unlock()
			return
		}
		replaced = old
		close(old.Send)
	}

	clients[client.ID] = client
	h.mu.Unlock()

	if replaced != nil && replaced.Conn != nil {
		_ = replaced.Conn.Close()
	}

	log.Printf(
		"[WS] registered client=%s board=%s",
		client.ID,
		client.BoardID,
	)
}

// Unregister removes only the matching client.
// Repeated unregister calls are safe.
func (h *Hub) Unregister(client *Client) {
	if client == nil {
		return
	}

	removed := false

	h.mu.Lock()

	if clients, exists := h.boards[client.BoardID]; exists {
		if current, exists := clients[client.ID]; exists && current == client {
			delete(clients, client.ID)
			close(client.Send)
			removed = true

			if len(clients) == 0 {
				delete(h.boards, client.BoardID)
			}
		}
	}

	h.mu.Unlock()

	if removed {
		if client.Conn != nil {
			_ = client.Conn.Close()
		}

		log.Printf(
			"[WS] unregistered client=%s board=%s",
			client.ID,
			client.BoardID,
		)
	}
}

// Broadcast sends an event only to clients on its board.
// Slow clients are removed instead of blocking other clients.
func (h *Hub) Broadcast(event *models.WSEvent) {
	if event == nil || event.BoardID == "" {
		return
	}

	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("[WS] event serialization failed: %v", err)
		return
	}

	var disconnected []*Client

	h.mu.Lock()

	clients := h.boards[event.BoardID]

	for id, client := range clients {
		select {
		case client.Send <- data:
		default:
			delete(clients, id)
			close(client.Send)
			disconnected = append(disconnected, client)
		}
	}

	if len(clients) == 0 {
		delete(h.boards, event.BoardID)
	}

	h.mu.Unlock()

	for _, client := range disconnected {
		if client.Conn != nil {
			_ = client.Conn.Close()
		}
		log.Printf(
			"[WS] disconnected slow client=%s board=%s",
			client.ID,
			client.BoardID,
		)
	}
}

// GetPresence returns each connected user once.
// Sorting makes the result deterministic for clients and tests.
func (h *Hub) GetPresence(boardID string) []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	seen := make(map[string]struct{})
	users := make([]string, 0)

	for _, client := range h.boards[boardID] {
		if _, exists := seen[client.UserID]; exists {
			continue
		}
		seen[client.UserID] = struct{}{}
		users = append(users, client.UserID)
	}

	sort.Strings(users)
	return users
}

// WritePump forwards outgoing messages to the WebSocket.
func (c *Client) WritePump() {
	if c.Conn == nil {
		return
	}

	defer c.Conn.Close()

	for msg := range c.Send {
		if err := c.Conn.WriteMessage(
			websocket.TextMessage,
			msg,
		); err != nil {
			return
		}
	}
}

// ReadPump keeps the connection alive while reading.
// Disconnects are cleaned up even after read failures.
func (c *Client) ReadPump(hub *Hub) {
	if c.Conn == nil {
		hub.Unregister(c)
		return
	}

	defer func() {
		hub.Unregister(c)
		_ = c.Conn.Close()
	}()

	for {
		if _, _, err := c.Conn.ReadMessage(); err != nil {
			return
		}
	}
}
