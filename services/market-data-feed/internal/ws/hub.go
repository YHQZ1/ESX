package ws

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/YHQZ1/esx/packages/logger"
	"github.com/gorilla/websocket"
)

type Client struct {
	conn          *websocket.Conn
	send          chan []byte
	subscriptions map[string]bool
	mu            sync.Mutex
}

type Message struct {
	Channel string `json:"channel"`
	Data    any    `json:"data"`
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan BroadcastMessage
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	log        *logger.Logger
}

type BroadcastMessage struct {
	Channel string
	Data    any
}

func NewHub(log *logger.Logger) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan BroadcastMessage, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		log:        log,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			h.log.Debug("client connected")

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			h.log.Debug("client disconnected")

		case msg := <-h.broadcast:
			payload, err := json.Marshal(Message{Channel: msg.Channel, Data: msg.Data})
			if err != nil {
				continue
			}
			h.mu.RLock()
			for client := range h.clients {
				client.mu.Lock()
				subscribed := client.subscriptions[msg.Channel]
				client.mu.Unlock()
				if subscribed {
					select {
					case client.send <- payload:
					default:
						close(client.send)
						delete(h.clients, client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) Broadcast(channel string, data any) {
	h.broadcast <- BroadcastMessage{Channel: channel, Data: data}
}

func (h *Hub) Register(conn *websocket.Conn) *Client {
	client := &Client{
		conn:          conn,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}
	h.register <- client
	return client
}

func (c *Client) Subscribe(channel string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.subscriptions[channel] = true
}

func (c *Client) Unsubscribe(channel string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.subscriptions, channel)
}

func (c *Client) WritePump(hub *Hub) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
		hub.unregister <- c
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
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

func (c *Client) ReadPump(hub *Hub) {
	defer func() {
		hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		var sub struct {
			Action  string `json:"action"`
			Channel string `json:"channel"`
		}
		if err := json.Unmarshal(msg, &sub); err != nil {
			continue
		}

		switch sub.Action {
		case "subscribe":
			c.Subscribe(sub.Channel)
		case "unsubscribe":
			c.Unsubscribe(sub.Channel)
		}
	}
}
