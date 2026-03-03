package ws

import (
	"context"
	"time"

	"github.com/NickCharlie/ubothub/backend/internal/event"
	"go.uber.org/zap"
)

// RegisterEventSubscribers subscribes the WebSocket hub to relevant events
// from the application event bus. This bridges the gateway/adapter layer
// with real-time WebSocket delivery to connected clients.
func RegisterEventSubscribers(hub *Hub, bus *event.Bus, logger *zap.Logger) {
	bus.Subscribe(event.BotMessageReceived, func(_ context.Context, evt event.Event) error {
		msg, ok := evt.Payload.(event.BotMessageEvent)
		if !ok {
			logger.Warn("unexpected payload type for BotMessageReceived event")
			return nil
		}

		hub.BroadcastToRoom(msg.BotID, &OutboundMessage{
			Type:      TypeBotReply,
			BotID:     msg.BotID,
			Content:   msg.Content,
			Sender:    msg.Sender.Nickname,
			Timestamp: msg.Timestamp,
			Data:      msg.Metadata,
		})
		return nil
	})

	bus.Subscribe(event.BotStatusChanged, func(_ context.Context, evt event.Event) error {
		msg, ok := evt.Payload.(event.BotStatusEvent)
		if !ok {
			logger.Warn("unexpected payload type for BotStatusChanged event")
			return nil
		}

		hub.BroadcastToRoom(msg.BotID, &OutboundMessage{
			Type:      TypeBotStatus,
			BotID:     msg.BotID,
			Timestamp: time.Now().Unix(),
			Data:      map[string]interface{}{"status": msg.Status},
		})
		return nil
	})

	logger.Info("websocket event subscribers registered")
}
