package ws

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// Client represents a single WebSocket connection.
type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	userID string
	roomID string
	logger *zap.Logger
	config Config
}

// NewClient creates a new WebSocket client bound to the given connection.
func NewClient(hub *Hub, conn *websocket.Conn, userID, roomID string, logger *zap.Logger) *Client {
	return &Client{
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, hub.config.SendChannelSize),
		userID: userID,
		roomID: roomID,
		logger: logger,
		config: hub.config,
	}
}

// UserID returns the authenticated user ID of this client.
func (c *Client) UserID() string {
	return c.userID
}

// RoomID returns the room (bot channel) this client is subscribed to.
func (c *Client) RoomID() string {
	return c.roomID
}

// Hub returns the hub this client belongs to.
func (c *Client) Hub() *Hub {
	return c.hub
}

// ReadPump pumps messages from the WebSocket connection to the hub.
// Must be run as a goroutine per client.
func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(c.config.MaxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(c.config.PongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(c.config.PongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseNormalClosure,
			) {
				c.logger.Warn("websocket read error",
					zap.String("user_id", c.userID),
					zap.Error(err),
				)
			}
			break
		}

		var inbound InboundMessage
		if err := json.Unmarshal(message, &inbound); err != nil {
			c.logger.Debug("invalid websocket message format",
				zap.String("user_id", c.userID),
				zap.Error(err),
			)
			continue
		}

		// Handle application-level ping/pong.
		if inbound.Type == TypePing {
			pong := OutboundMessage{Type: TypePong, Timestamp: time.Now().Unix()}
			data, _ := json.Marshal(pong)
			select {
			case c.send <- data:
			default:
				// Send channel full; skip pong.
			}
			continue
		}

		select {
		case c.hub.inbound <- &ClientMessage{Client: c, Message: &inbound}:
		default:
			c.logger.Warn("inbound channel full, dropping client message",
				zap.String("user_id", c.userID),
			)
		}
	}
}

// WritePump pumps messages from the hub to the WebSocket connection.
// Must be run as a goroutine per client.
func (c *Client) WritePump() {
	ticker := time.NewTicker(c.config.PingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(c.config.WriteWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Coalesce queued messages into a single write frame to reduce syscalls.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(c.config.WriteWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ClientMessage wraps an inbound message with its originating client.
type ClientMessage struct {
	Client  *Client
	Message *InboundMessage
}
