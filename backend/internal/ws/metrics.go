package ws

import "sync/atomic"

// Metrics exposes real-time WebSocket connection statistics.
// All fields are updated atomically for lock-free reads.
type Metrics struct {
	TotalConnections   atomic.Int64
	TotalRooms         atomic.Int64
	TotalMessagesSent  atomic.Int64
	TotalMessagesRecv  atomic.Int64
	TotalBroadcasts    atomic.Int64
	RejectedConnLimit  atomic.Int64
	RejectedRoomLimit  atomic.Int64
	RejectedUserLimit  atomic.Int64
	DroppedSlowClients atomic.Int64
}

// Snapshot returns a point-in-time copy of all metrics.
func (m *Metrics) Snapshot() MetricsSnapshot {
	return MetricsSnapshot{
		TotalConnections:   m.TotalConnections.Load(),
		TotalRooms:         m.TotalRooms.Load(),
		TotalMessagesSent:  m.TotalMessagesSent.Load(),
		TotalMessagesRecv:  m.TotalMessagesRecv.Load(),
		TotalBroadcasts:    m.TotalBroadcasts.Load(),
		RejectedConnLimit:  m.RejectedConnLimit.Load(),
		RejectedRoomLimit:  m.RejectedRoomLimit.Load(),
		RejectedUserLimit:  m.RejectedUserLimit.Load(),
		DroppedSlowClients: m.DroppedSlowClients.Load(),
	}
}

// MetricsSnapshot is a serializable snapshot of WebSocket metrics.
type MetricsSnapshot struct {
	TotalConnections   int64 `json:"total_connections"`
	TotalRooms         int64 `json:"total_rooms"`
	TotalMessagesSent  int64 `json:"total_messages_sent"`
	TotalMessagesRecv  int64 `json:"total_messages_recv"`
	TotalBroadcasts    int64 `json:"total_broadcasts"`
	RejectedConnLimit  int64 `json:"rejected_conn_limit"`
	RejectedRoomLimit  int64 `json:"rejected_room_limit"`
	RejectedUserLimit  int64 `json:"rejected_user_limit"`
	DroppedSlowClients int64 `json:"dropped_slow_clients"`
}
