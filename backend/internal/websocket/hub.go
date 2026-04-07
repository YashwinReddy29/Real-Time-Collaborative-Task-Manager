package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"taskflow/internal/models"

	"github.com/gorilla/websocket"
)

type Client struct {
	ID      string
	UserID  string
	BoardID string
	Conn    *websocket.Conn
	Send    chan []byte
}

type Hub struct {
	mu      sync.RWMutex
	boards  map[string]map[string]*Client // boardID -> clientID -> client
	register   chan *Client
	unregister chan *Client
	broadcast  chan *models.WSEvent
}

func NewHub() *Hub {
	return &Hub{
		boards:     make(map[string]map[string]*Client),
		register:   make(chan *Client, 256),
		unregister: make(chan *Client, 256),
		broadcast:  make(chan *models.WSEvent, 256),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if _, ok := h.boards[client.BoardID]; !ok {
				h.boards[client.BoardID] = make(map[string]*Client)
			}
			h.boards[client.BoardID][client.ID] = client
			h.mu.Unlock()
			log.Printf("[WS] client %s joined board %s", client.UserID, client.BoardID)

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.boards[client.BoardID]; ok {
				if _, ok := clients[client.ID]; ok {
					delete(clients, client.ID)
					close(client.Send)
					if len(clients) == 0 {
						delete(h.boards, client.BoardID)
					}
				}
			}
			h.mu.Unlock()
			log.Printf("[WS] client %s left board %s", client.UserID, client.BoardID)

		case event := <-h.broadcast:
			data, err := json.Marshal(event)
			if err != nil {
				continue
			}
			h.mu.RLock()
			clients := h.boards[event.BoardID]
			for _, client := range clients {
				select {
				case client.Send <- data:
				default:
					close(client.Send)
					delete(clients, client.ID)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) Broadcast(event *models.WSEvent) {
	h.broadcast <- event
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

func (c *Client) WritePump() {
	defer c.Conn.Close()
	for msg := range c.Send {
		if err := c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}

func (c *Client) ReadPump(hub *Hub) {
	defer func() {
		hub.Unregister(c)
		c.Conn.Close()
	}()
	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (h *Hub) GetPresence(boardID string) []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	var users []string
	seen := map[string]bool{}
	for _, client := range h.boards[boardID] {
		if !seen[client.UserID] {
			users = append(users, client.UserID)
			seen[client.UserID] = true
		}
	}
	return users
}
