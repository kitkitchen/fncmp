package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var connPool = conns{
	pool: make(map[string]*Conn),
}

func (c *conns) Get(id string) (*Conn, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	conn, ok := c.pool[id]
	return conn, ok
}

func (c *conns) Set(id string, conn *Conn) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.pool[id] = conn
}

func (c *conns) Delete(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.pool, id)
}

type (
	conns struct {
		mu   sync.Mutex
		pool map[string]*Conn
	}
	Conn struct {
		websocket *websocket.Conn
		ID        string
		HandlerID string
		Messages  chan []byte
	}
)

// NewConn upgrades an http connection to a websocket connection and returns a Conn
// or an error if the upgrade fails.
func NewConn(w http.ResponseWriter, r *http.Request, handlerID string, ID string) (*Conn, error) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// TODO: check cookie for session id
			// Set as conn id
			return true
			// host := strings.Split(r.Host, ":")[0]
			// return host == "localhost"
		},
	}
	websocket, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, errors.New("failed to upgrade connection")
	}

	c := &Conn{
		websocket: websocket,
		ID:        ID,
		HandlerID: handlerID,
		Messages:  make(chan []byte, 16),
	}
	connPool.Set(c.ID, c)
	return c, nil
}

// Close closes the websocket connection and removes the Conn from the pool.
// It returns an error if the Conn is nil.
func (c *Conn) close() error {
	if c == nil {
		return errors.New("cannot close nil connection")
	}
	evtListeners.Remove(c)
	connPool.Delete(c.ID)
	c.websocket.Close()
	return nil
}

func (c *Conn) listen() {
	go func(c *Conn) {
		defer c.close()
		// Listen for messages on the Conn's Messages channel
		var dispatch Dispatch
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
			// Parse dispatch from websocket message
			err = json.Unmarshal(message, &dispatch)
			if err != nil {
				log.Printf("error: %v", err)
				continue
			}
			// Get handler from handler pool
			handler, ok := handlers.Get(dispatch.HandlerID)
			if !ok {
				log.Printf("error: handler '%s' not found", dispatch.HandlerID)
				continue
			}
			// Set Conn on dispatch
			dispatch.Conn = c
			// Dispatch to handler
			handler.in <- dispatch
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
