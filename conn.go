package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var connPool = conns{
	pool: make(map[string]*Conn),
}

func UseConnPool() *conns {
	return &connPool
}

func (c *conns) Get(id string) (*Conn, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	conn, ok := c.pool[id]
	return conn, ok
}

type (
	// Conn is a websocket connection with a unique key, a pointer to a pool, and a channel for messages.
	conns struct {
		mu   sync.Mutex
		pool map[string]*Conn
	}
	Conn struct {
		websocket *websocket.Conn
		ID        string
		Messages  chan []byte
	}
	ConnectionInfo struct {
		ConnID string `json:"conn_id"`
		Config struct {
			BaseUrl   string `json:"base_url"`
			MainRoute string `json:"main_route"`
		}
	}
)

// NewConn upgrades an http connection to a websocket connection and returns a Conn
// or an error if the upgrade fails.
func NewConn(w http.ResponseWriter, r *http.Request) (*Conn, error) {
	connPool.mu.Lock()
	defer connPool.mu.Unlock()
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		}, // allow all connections
	}
	websocket, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, errors.New("failed to upgrade connection")
	}

	c := &Conn{
		websocket: websocket,
		ID:        uuid.New().String(),
		Messages:  make(chan []byte, 16),
	}
	connPool.pool[c.ID] = c
	return c, nil
}

// Close closes the websocket connection and removes the Conn from the pool.
// It returns an error if the Conn is nil.
func (c *Conn) close() error {
	if c == nil {
		return errors.New("cannot close nil connection")
	}
	c.websocket.Close()
	return nil
}

// Listen listens for messages on the Conn's Messages channel and writes them to the websocket connection.
func (c *Conn) listen() {
	go func(c *Conn) {
		for {
			_, message, err := c.websocket.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(
					err,
					websocket.CloseGoingAway,
					websocket.CloseAbnormalClosure,
					websocket.CloseNormalClosure,
				) {
					log.Printf("error: %v", err)
				}
				close(c.Messages)
				break
			}
			log.Printf("received: %s", message)
		}
	}(c)

	for {
		msg, ok := <-c.Messages
		if !ok {
			c.close()
			break
		}

		if err := c.websocket.WriteMessage(1, msg); err != nil {
			log.Printf("error: %v", err)
			c.close()
		}
	}
}

// Publish publishes a message to the Conn's Messages channel.
func (c *Conn) Publish(msg []byte) {
	// if msg is not json encodable, return
	_, err := json.Marshal(msg)
	if err != nil {
		log.Printf("error: %v", err)
		return
	}
	c.Messages <- msg
}

func (c *Conn) Write(p []byte) (n int, err error) {
	c.Messages <- p
	return len(p), nil
}
