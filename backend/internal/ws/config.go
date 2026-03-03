package ws

import "time"

// Config holds tunable parameters for the WebSocket hub and clients.
type Config struct {
	// MaxConnections is the global maximum number of simultaneous WebSocket connections.
	// Zero means unlimited.
	MaxConnections int

	// MaxConnectionsPerRoom is the maximum number of clients per room (bot channel).
	// Zero means unlimited.
	MaxConnectionsPerRoom int

	// MaxConnectionsPerUser is the maximum number of WebSocket connections per user.
	// Zero means unlimited.
	MaxConnectionsPerUser int

	// WriteWait is the time allowed to write a message to the peer.
	WriteWait time.Duration

	// PongWait is the time allowed to read the next pong message from the peer.
	PongWait time.Duration

	// PingPeriod is the interval at which server sends ping frames.
	// Must be less than PongWait.
	PingPeriod time.Duration

	// MaxMessageSize is the maximum size in bytes for a single inbound message.
	MaxMessageSize int64

	// ReadBufferSize is the size of the per-connection read buffer.
	ReadBufferSize int

	// WriteBufferSize is the size of the per-connection write buffer.
	WriteBufferSize int

	// SendChannelSize is the capacity of each client's outbound message channel.
	SendChannelSize int

	// BroadcastChannelSize is the capacity of the hub's broadcast channel.
	BroadcastChannelSize int

	// InboundChannelSize is the capacity of the hub's inbound message channel.
	InboundChannelSize int
}

// DefaultConfig returns production-ready default WebSocket configuration.
func DefaultConfig() Config {
	return Config{
		MaxConnections:        10000,
		MaxConnectionsPerRoom: 500,
		MaxConnectionsPerUser: 5,
		WriteWait:             10 * time.Second,
		PongWait:              60 * time.Second,
		PingPeriod:            54 * time.Second, // 90% of PongWait
		MaxMessageSize:        8192,
		ReadBufferSize:        4096,
		WriteBufferSize:       4096,
		SendChannelSize:       256,
		BroadcastChannelSize:  4096,
		InboundChannelSize:    1024,
	}
}
