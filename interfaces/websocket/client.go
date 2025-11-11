// interfaces/websocket/client.go
package websocket

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/gofiber/contrib/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512 * 1024 // 512KB
)

// ReadPump pumps messages from the websocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		c.IsAlive = true
		c.LastPingTime = time.Now()
		return nil
	})

	// Initialize rate limiting
	c.lastReset = time.Now()
	c.messageCount = 0

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("websocket error: %v", err)
			}
			break
		}

		// Simple rate limiting - 60 messages per minute
		if time.Since(c.lastReset) > time.Minute {
			c.messageCount = 0
			c.lastReset = time.Now()
		}

		if c.messageCount > 60 {
			c.sendError("Rate limit exceeded. Max 60 messages per minute", "")
			continue
		}
		c.messageCount++

		// Parse message
		var wsMsg WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			c.sendError("Invalid message format", wsMsg.RequestID)
			continue
		}

		// Validate message size
		if len(message) > maxMessageSize {
			c.sendError("Message too large", wsMsg.RequestID)
			continue
		}

		// Handle message
		if handler, ok := c.Hub.handlers[string(wsMsg.Type)]; ok {
			// Increment message counter
			c.Hub.IncrementMessageCount()

			if err := handler.Handle(context.Background(), c, wsMsg.Data); err != nil {
				c.sendError(err.Error(), wsMsg.RequestID)
			}
		} else {
			c.sendError("Unknown message type", wsMsg.RequestID)
		}
	}
}

// WritePump pumps messages from the hub to the websocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.Conn.WriteMessage(websocket.TextMessage, message)

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// sendError sends an error message to the client
func (c *Client) sendError(errorMsg string, requestID string) {
	response := WSResponse{
		Type:      "error",
		Success:   false,
		Error:     errorMsg,
		Timestamp: time.Now(),
		RequestID: requestID,
	}

	if data, err := json.Marshal(response); err == nil {
		select {
		case c.Send <- data:
		default:
			// Channel is full, close connection
			close(c.Send)
		}
	}
}
