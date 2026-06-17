package signaling

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
)

type Message struct {
	Type    string          `json:"type"`
	Sender  string          `json:"sender"`
	Target  string          `json:"target"`
	Payload json.RawMessage `json:"payload"`
}

type Client struct {
	ID   string
	Conn *websocket.Conn
}

type Hub struct {
	clients map[string]*Client
	mu      sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]*Client),
	}
}

func (h *Hub) register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[c.ID] = c
	log.Printf("Signaling client registered: %s", c.ID)
}

func (h *Hub) unregister(id string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, id)
	log.Printf("Signaling client unregistered: %s", id)
}

func (h *Hub) relay(msg Message) {
	h.mu.RLock()
	targetClient, ok := h.clients[msg.Target]
	h.mu.RUnlock()

	if !ok {
		log.Printf("Signaling target not found: %s", msg.Target)
		return
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal relayed message: %v", err)
		return
	}

	if err := targetClient.Conn.Write(context.Background(), websocket.MessageText, data); err != nil {
		log.Printf("Failed to send message to %s: %v", msg.Target, err)
	}
}

func Handler(hub *Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		tsNodeRaw, exists := c.Get("ts_node")
		if !exists {
			log.Println("Signaling failed: ts_node not found in context")
			return
		}
		tsNode := tsNodeRaw.(string)

		conn, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{
			InsecureSkipVerify: true,
		})
		if err != nil {
			log.Printf("failed to accept websocket for signaling: %v", err)
			return
		}

		client := &Client{ID: tsNode, Conn: conn}
		hub.register(client)
		defer func() {
			hub.unregister(tsNode)
			conn.Close(websocket.StatusInternalError, "closing")
		}()

		ctx := c.Request.Context()
		for {
			_, data, err := conn.Read(ctx)
			if err != nil {
				break
			}

			var msg Message
			if err := json.Unmarshal(data, &msg); err != nil {
				log.Printf("Invalid signaling message from %s: %v", tsNode, err)
				continue
			}

			// Force sender to be the actual connected node
			msg.Sender = tsNode
			hub.relay(msg)
		}
	}
}
