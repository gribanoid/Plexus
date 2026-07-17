package websocket

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/google/uuid"
	"github.com/plexus/backend/internal/metrics"
	"github.com/redis/go-redis/v9"
)

const redisPubSubChannel = "plexus:events"

// EventType identifies the kind of real-time event.
type EventType string

const (
	EventIssueCreated  EventType = "issue.created"
	EventIssueUpdated  EventType = "issue.updated"
	EventIssueDeleted  EventType = "issue.deleted"
	EventCommentCreated EventType = "comment.created"
	EventSprintUpdated EventType = "sprint.updated"
	EventNotification  EventType = "notification"
)

// Event is the envelope sent over WebSocket to clients.
type Event struct {
	Type      EventType   `json:"type"`
	ProjectID *uuid.UUID  `json:"project_id,omitempty"`
	UserID    *uuid.UUID  `json:"user_id,omitempty"`
	Payload   interface{} `json:"payload"`
}

// Client represents a connected WebSocket client.
type Client struct {
	UserID    uuid.UUID
	ProjectID *uuid.UUID
	Send      chan []byte
	hub       *Hub
}

func (c *Client) Close() {
	c.hub.unregister <- c
}

// Hub manages all WebSocket connections and broadcasts via Redis pub/sub.
type Hub struct {
	clients    map[*Client]bool
	mu         sync.RWMutex
	broadcast  chan *Event
	register   chan *Client
	unregister chan *Client
	redis      *redis.Client
}

func NewHub(rdb *redis.Client) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan *Event, 256),
		register:   make(chan *Client, 64),
		unregister: make(chan *Client, 64),
		redis:      rdb,
	}
}

func (h *Hub) Run() {
	// Subscribe to Redis for multi-instance broadcasting
	go h.subscribeRedis()

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			n := len(h.clients)
			h.mu.Unlock()
			metrics.WSClients.Set(float64(n))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
			n := len(h.clients)
			h.mu.Unlock()
			metrics.WSClients.Set(float64(n))

		case event := <-h.broadcast:
			data, err := json.Marshal(event)
			if err != nil {
				slog.Error("ws hub: marshal event", "error", err)
				continue
			}
			h.deliverLocal(event, data)
			// Publish to Redis so other instances receive it too
			_ = h.redis.Publish(context.Background(), redisPubSubChannel, data).Err()
		}
	}
}

// Publish enqueues an event for broadcast. Safe to call from any goroutine.
func (h *Hub) Publish(event *Event) {
	h.broadcast <- event
}

// Register adds a new client to the hub.
func (h *Hub) Register(client *Client) {
	client.hub = h
	h.register <- client
}

func (h *Hub) deliverLocal(event *Event, data []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		// Project-scoped events must only reach clients subscribed to that project.
		if event.ProjectID != nil {
			if client.ProjectID == nil || *event.ProjectID != *client.ProjectID {
				continue
			}
		}
		// User-targeted events (e.g. notifications) require matching user.
		if event.UserID != nil && client.UserID != *event.UserID {
			continue
		}
		select {
		case client.Send <- data:
		default:
			// Slow client — drop message
			slog.Warn("ws hub: slow client, dropping message", "user_id", client.UserID.String())
		}
	}
}

func (h *Hub) subscribeRedis() {
	pubsub := h.redis.Subscribe(context.Background(), redisPubSubChannel)
	defer pubsub.Close()

	ch := pubsub.Channel()
	for msg := range ch {
		var event Event
		if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
			slog.Error("ws hub: unmarshal redis event", "error", err)
			continue
		}
		data := []byte(msg.Payload)
		h.deliverLocal(&event, data)
	}
}
