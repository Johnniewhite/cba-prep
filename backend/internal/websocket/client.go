package websocket

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 * 1024
)

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.Hub.logger.WithError(err).Errorf("WebSocket error for client %s", c.ID)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			c.Hub.logger.WithError(err).Error("Failed to unmarshal message")
			continue
		}

		msg.UserID = c.UserID
		msg.Timestamp = time.Now()

		c.handleMessage(&msg)
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
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) handleMessage(msg *Message) {
	switch MessageType(msg.Type) {
	case MessageTypeChat:
		c.handleChatMessage(msg)
	case MessageTypeTaskUpdate:
		c.handleTaskUpdate(msg)
	case MessageTypeTyping:
		c.handleTypingIndicator(msg)
	default:
		c.Hub.logger.Warnf("Unknown message type: %s", msg.Type)
	}
}

func (c *Client) handleChatMessage(msg *Message) {
	if msg.Room == "" {
		msg.Room = "team:" + c.TeamID
	}
	c.Hub.broadcast <- msg
}

func (c *Client) handleTaskUpdate(msg *Message) {
	msg.Room = "team:" + c.TeamID
	c.Hub.broadcast <- msg
}

func (c *Client) handleTypingIndicator(msg *Message) {
	msg.Room = "team:" + c.TeamID
	c.Hub.broadcast <- msg
}

func (c *Client) JoinRoom(room string) {
	c.Hub.mu.Lock()
	defer c.Hub.mu.Unlock()
	c.Hub.joinRoom(c, room)
}

func (c *Client) LeaveRoom(room string) {
	c.Hub.mu.Lock()
	defer c.Hub.mu.Unlock()
	c.Hub.leaveRoom(c, room)
}

func (c *Client) SendMessage(message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	select {
	case c.Send <- data:
		return nil
	default:
		return websocket.ErrCloseSent
	}
}