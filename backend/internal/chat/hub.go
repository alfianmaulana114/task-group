package chat

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/alfianhamzah/task-group/backend/internal/org"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
}

type Hub struct {
	Redis *redis.Client
	mu    sync.RWMutex
	clients map[string]map[*Client]bool // key = orgID:projectID
	subs map[string]bool // channels already subscribed
	allowedOrigins map[string]struct{}
}

type Client struct {
	hub       *Hub
	conn      *websocket.Conn
	send      chan []byte
	orgID     string
	projectID string
}

func NewHub(redis *redis.Client) *Hub {
	h := &Hub{
		Redis: redis,
		clients: make(map[string]map[*Client]bool),
		subs: make(map[string]bool),
		allowedOrigins: make(map[string]struct{}),
	}
	upgrader.CheckOrigin = h.checkOrigin
	return h
}

func NewHubWithOrigins(redis *redis.Client, allowedOrigins []string) *Hub {
	h := NewHub(redis)
	for _, o := range allowedOrigins {
		h.allowedOrigins[o] = struct{}{}
	}
	// setup CheckOrigin based on allowed origins
	upgrader.CheckOrigin = h.checkOrigin
	return h
}

func (h *Hub) checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true // non-browser or same-origin
	}
	if len(h.allowedOrigins) == 0 {
		return false // fail-closed: deny if no origins configured
	}
	_, ok := h.allowedOrigins[origin]
	return ok
}

func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	claims := getClaims(r.Context())
	if claims == nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	orgID := org.GetOrgID(r.Context())
	projectID := chi.URLParam(r, "project_id")
	if orgID == "" || projectID == "" {
		http.Error(w, `{"error":"organization and project ID required"}`, http.StatusBadRequest)
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade: %v", err)
		return
	}
	client := &Client{
		hub: h,
		conn: conn,
		send: make(chan []byte, 256),
		orgID: orgID,
		projectID: projectID,
	}
	key := orgID + ":" + projectID
	h.mu.Lock()
	if h.clients[key] == nil {
		h.clients[key] = make(map[*Client]bool)
	}
	h.clients[key][client] = true
	h.mu.Unlock()

	go client.writePump()
	go client.readPump(h)
	// start redis subscriber for this key once
	h.ensureRedisSub(context.Background(), orgID, projectID)
}

func (h *Hub) ensureRedisSub(ctx context.Context, orgID, projectID string) {
	key := orgID + ":" + projectID
	channel := "chat:" + key
	h.mu.Lock()
	if h.subs[channel] {
		h.mu.Unlock()
		return
	}
	h.subs[channel] = true
	h.mu.Unlock()
	go func() {
		pubsub := h.Redis.Subscribe(ctx, channel)
		defer func() {
			pubsub.Close()
			h.mu.Lock()
			delete(h.subs, channel)
			h.mu.Unlock()
		}()
		ch := pubsub.Channel()
		for msg := range ch {
			payload := []byte(msg.Payload)
			h.mu.RLock()
			for client := range h.clients[key] {
				select {
				case client.send <- payload:
				default:
					// drop if full
				}
			}
			h.mu.RUnlock()
		}
		if err := pubsub.Close(); err != nil {
			log.Printf("redis pubsub close: %v", err)
		}
	}()
}

func (h *Hub) broadcast(orgID, projectID string, payload []byte) {
	key := orgID + ":" + projectID
	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.clients[key] {
		select {
		case client.send <- payload:
		default:
		}
	}
}

// also allow direct publish via hub (called from handler fallback)
func (h *Hub) PublishDirect(orgID, projectID string, payload interface{}) {
	b, _ := json.Marshal(payload)
	h.broadcast(orgID, projectID, b)
}

func (c *Client) readPump(h *Hub) {
	defer func() {
		h.mu.Lock()
		key := c.orgID + ":" + c.projectID
		delete(h.clients[key], c)
		if len(h.clients[key]) == 0 {
			delete(h.clients, key)
		}
		h.mu.Unlock()
		c.conn.Close()
	}()
	c.conn.SetReadLimit(512 * 1024)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			// flush queued
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
