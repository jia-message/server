package ws

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gofiber/contrib/v3/websocket"
	"github.com/google/uuid"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512 * 1024 // 512KB (large enough for metadata or key exchanges)
)

type Client struct {
	Hub    *Hub
	Conn   *websocket.Conn
	UserID uuid.UUID
	Send   chan []byte
}

type wsMessage struct {
	Type           string          `json:"type"`
	ConversationID string          `json:"conversation_id,omitempty"`
	MessageID      string          `json:"message_id,omitempty"`
	Body           string          `json:"body,omitempty"`
	Emoji          string          `json:"emoji,omitempty"`
	Payload        json.RawMessage `json:"payload,omitempty"`
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.unregister <- c
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}

		// Process incoming message
		c.handleIncomingMessage(message)
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, _ = w.Write(message)

			// Add queued chat messages to the current websocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				_, _ = w.Write([]byte{'\n'})
				_, _ = w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) Close() {
	close(c.Send)
	c.Conn.Close()
}

// Global hooks back into message service and conversation service initialized in main
var WSIncomingMessageCallback func(client *Client, rawMessage []byte)

func (c *Client) handleIncomingMessage(rawMessage []byte) {
	if WSIncomingMessageCallback != nil {
		WSIncomingMessageCallback(c, rawMessage)
	}
}
