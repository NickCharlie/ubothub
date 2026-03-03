package ws

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Room groups clients subscribed to the same bot channel.
type Room struct {
	ID      string
	clients map[*Client]bool
	mu      sync.RWMutex
}

// Hub maintains the set of active clients and broadcasts messages to rooms.
// Designed for high concurrency with:
//   - configurable connection limits (global, per-room, per-user)
//   - lock-free atomic metrics
//   - buffered channels to prevent goroutine blocking
//   - graceful shutdown via context cancellation
type Hub struct {
	config     Config
	rooms      map[string]*Room
	clients    map[*Client]bool
	userConns  map[string]int // userID -> connection count
	register   chan *Client
	unregister chan *Client
	broadcast  chan *RoomMessage
	inbound    chan *ClientMessage
	mu         sync.RWMutex
	logger     *zap.Logger
	metrics    *Metrics
	onMessage  func(client *Client, msg *InboundMessage)
}

// NewHub creates a new WebSocket hub with the given configuration.
func NewHub(cfg Config, logger *zap.Logger) *Hub {
	return &Hub{
		config:     cfg,
		rooms:      make(map[string]*Room),
		clients:    make(map[*Client]bool),
		userConns:  make(map[string]int),
		register:   make(chan *Client, 64),
		unregister: make(chan *Client, 64),
		broadcast:  make(chan *RoomMessage, cfg.BroadcastChannelSize),
		inbound:    make(chan *ClientMessage, cfg.InboundChannelSize),
		logger:     logger,
		metrics:    &Metrics{},
	}
}

// SetMessageHandler registers a callback for inbound client messages.
func (h *Hub) SetMessageHandler(fn func(client *Client, msg *InboundMessage)) {
	h.onMessage = fn
}

// GetMetrics returns the hub's metrics instance for monitoring.
func (h *Hub) GetMetrics() *Metrics {
	return h.metrics
}

// GetConfig returns the hub's configuration.
func (h *Hub) GetConfig() Config {
	return h.config
}

// Run starts the hub's main event loop. It blocks until ctx is cancelled.
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			h.shutdown()
			return

		case client := <-h.register:
			h.addClient(client)

		case client := <-h.unregister:
			h.removeClient(client)

		case rm := <-h.broadcast:
			h.broadcastToRoom(rm)

		case cm := <-h.inbound:
			h.metrics.TotalMessagesRecv.Add(1)
			if h.onMessage != nil {
				h.onMessage(cm.Client, cm.Message)
			}
		}
	}
}

// Register enqueues a client for registration with the hub.
// Returns false if the registration channel is full (back-pressure).
func (h *Hub) Register(client *Client) bool {
	select {
	case h.register <- client:
		return true
	default:
		h.logger.Warn("register channel full, rejecting client",
			zap.String("user_id", client.userID),
			zap.String("room_id", client.roomID),
		)
		return false
	}
}

// Unregister enqueues a client for removal from the hub.
func (h *Hub) Unregister(client *Client) {
	select {
	case h.unregister <- client:
	default:
		// Channel full; spawn goroutine to avoid blocking the caller.
		go func() { h.unregister <- client }()
	}
}

// BroadcastToRoom sends a pre-built outbound message to all clients in a room.
func (h *Hub) BroadcastToRoom(roomID string, msg *OutboundMessage) {
	select {
	case h.broadcast <- &RoomMessage{RoomID: roomID, Message: msg}:
	default:
		h.logger.Warn("broadcast channel full, dropping message",
			zap.String("room_id", roomID),
			zap.String("type", msg.Type),
		)
	}
}

// SendToClient sends a message directly to a specific client.
func (h *Hub) SendToClient(client *Client, msg *OutboundMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		h.logger.Error("failed to marshal direct message", zap.Error(err))
		return
	}
	select {
	case client.send <- data:
		h.metrics.TotalMessagesSent.Add(1)
	default:
		h.logger.Debug("client send channel full, dropping direct message",
			zap.String("user_id", client.userID),
		)
	}
}

// OnlineCount returns the total number of connected clients.
func (h *Hub) OnlineCount() int64 {
	return h.metrics.TotalConnections.Load()
}

// RoomCount returns the number of active rooms.
func (h *Hub) RoomCount() int64 {
	return h.metrics.TotalRooms.Load()
}

func (h *Hub) addClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Check global connection limit.
	if h.config.MaxConnections > 0 && len(h.clients) >= h.config.MaxConnections {
		h.metrics.RejectedConnLimit.Add(1)
		h.logger.Warn("global connection limit reached, rejecting client",
			zap.String("user_id", client.userID),
			zap.Int("limit", h.config.MaxConnections),
		)
		close(client.send)
		client.conn.Close()
		return
	}

	// Check per-user connection limit.
	if h.config.MaxConnectionsPerUser > 0 && h.userConns[client.userID] >= h.config.MaxConnectionsPerUser {
		h.metrics.RejectedUserLimit.Add(1)
		h.logger.Warn("per-user connection limit reached, rejecting client",
			zap.String("user_id", client.userID),
			zap.Int("limit", h.config.MaxConnectionsPerUser),
		)
		close(client.send)
		client.conn.Close()
		return
	}

	// Get or create room.
	room, ok := h.rooms[client.roomID]
	if !ok {
		room = &Room{
			ID:      client.roomID,
			clients: make(map[*Client]bool),
		}
		h.rooms[client.roomID] = room
		h.metrics.TotalRooms.Add(1)
	}

	// Check per-room connection limit.
	room.mu.RLock()
	roomSize := len(room.clients)
	room.mu.RUnlock()
	if h.config.MaxConnectionsPerRoom > 0 && roomSize >= h.config.MaxConnectionsPerRoom {
		h.metrics.RejectedRoomLimit.Add(1)
		h.logger.Warn("per-room connection limit reached, rejecting client",
			zap.String("room_id", client.roomID),
			zap.Int("limit", h.config.MaxConnectionsPerRoom),
		)
		// Clean up empty room if we just created it.
		if !ok {
			delete(h.rooms, client.roomID)
			h.metrics.TotalRooms.Add(-1)
		}
		close(client.send)
		client.conn.Close()
		return
	}

	// Register the client.
	h.clients[client] = true
	h.userConns[client.userID]++

	room.mu.Lock()
	room.clients[client] = true
	room.mu.Unlock()

	h.metrics.TotalConnections.Add(1)

	h.logger.Debug("client connected",
		zap.String("user_id", client.userID),
		zap.String("room_id", client.roomID),
		zap.Int64("total_connections", h.metrics.TotalConnections.Load()),
	)
}

func (h *Hub) removeClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client]; !ok {
		return
	}

	delete(h.clients, client)
	close(client.send)

	// Update per-user connection count.
	h.userConns[client.userID]--
	if h.userConns[client.userID] <= 0 {
		delete(h.userConns, client.userID)
	}

	if room, ok := h.rooms[client.roomID]; ok {
		room.mu.Lock()
		delete(room.clients, client)
		empty := len(room.clients) == 0
		room.mu.Unlock()

		if empty {
			delete(h.rooms, client.roomID)
			h.metrics.TotalRooms.Add(-1)
		}
	}

	h.metrics.TotalConnections.Add(-1)

	h.logger.Debug("client disconnected",
		zap.String("user_id", client.userID),
		zap.String("room_id", client.roomID),
		zap.Int64("total_connections", h.metrics.TotalConnections.Load()),
	)
}

func (h *Hub) broadcastToRoom(rm *RoomMessage) {
	h.mu.RLock()
	room, ok := h.rooms[rm.RoomID]
	h.mu.RUnlock()

	if !ok {
		return
	}

	data, err := json.Marshal(rm.Message)
	if err != nil {
		h.logger.Error("failed to marshal broadcast message", zap.Error(err))
		return
	}

	h.metrics.TotalBroadcasts.Add(1)

	room.mu.RLock()
	clients := make([]*Client, 0, len(room.clients))
	for client := range room.clients {
		clients = append(clients, client)
	}
	room.mu.RUnlock()

	// Fan out to clients outside the room lock to reduce contention.
	for _, client := range clients {
		select {
		case client.send <- data:
			h.metrics.TotalMessagesSent.Add(1)
		default:
			h.metrics.DroppedSlowClients.Add(1)
			// Slow client: schedule removal outside broadcast path.
			go func(c *Client) {
				h.Unregister(c)
			}(client)
		}
	}

	if rm.Message.Type == TypeBotReply || rm.Message.Type == TypeChat {
		h.logger.Debug("message broadcast to room",
			zap.String("room_id", rm.RoomID),
			zap.Int("recipients", len(clients)),
			zap.Int64("timestamp", time.Now().Unix()),
		)
	}
}

// shutdown gracefully closes all client connections.
func (h *Hub) shutdown() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.logger.Info("shutting down websocket hub",
		zap.Int("active_clients", len(h.clients)),
		zap.Int("active_rooms", len(h.rooms)),
	)

	for client := range h.clients {
		close(client.send)
		client.conn.Close()
		delete(h.clients, client)
	}

	h.rooms = make(map[string]*Room)
	h.userConns = make(map[string]int)
	h.metrics.TotalConnections.Store(0)
	h.metrics.TotalRooms.Store(0)
}
