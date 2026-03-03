package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/NickCharlie/ubothub/backend/internal/ws"
	"github.com/NickCharlie/ubothub/backend/pkg/token"
	"go.uber.org/zap"
)

// WSHandler handles WebSocket upgrade and connection management.
type WSHandler struct {
	hub      *ws.Hub
	tokenMgr *token.Manager
	logger   *zap.Logger
	upgrader websocket.Upgrader
}

// NewWSHandler creates a new WebSocket handler with the given allowed origins.
// If allowedOrigins is empty or nil, all origins are permitted (development mode).
func NewWSHandler(hub *ws.Hub, tokenMgr *token.Manager, allowedOrigins []string, logger *zap.Logger) *WSHandler {
	cfg := hub.GetConfig()

	originSet := make(map[string]bool, len(allowedOrigins))
	for _, o := range allowedOrigins {
		originSet[o] = true
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  cfg.ReadBufferSize,
		WriteBufferSize: cfg.WriteBufferSize,
		CheckOrigin: func(r *http.Request) bool {
			if len(originSet) == 0 {
				return true
			}
			origin := r.Header.Get("Origin")
			return originSet[origin]
		},
		EnableCompression: true,
	}

	return &WSHandler{
		hub:      hub,
		tokenMgr: tokenMgr,
		logger:   logger,
		upgrader: upgrader,
	}
}

// Connect handles GET /api/v1/ws.
// Upgrades the HTTP connection to WebSocket after JWT validation.
// The client must provide the JWT token as a query parameter: ?token=xxx&bot_id=xxx.
func (h *WSHandler) Connect(c *gin.Context) {
	tokenStr := c.Query("token")
	if tokenStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 10003, "message": "missing token"})
		return
	}

	claims, err := h.tokenMgr.ParseToken(tokenStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 11003, "message": "invalid or expired token"})
		return
	}

	botID := c.Query("bot_id")
	if botID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 10002, "message": "bot_id is required"})
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("websocket upgrade failed",
			zap.String("user_id", claims.UserID),
			zap.Error(err),
		)
		return
	}

	client := ws.NewClient(h.hub, conn, claims.UserID, botID, h.logger)

	if !h.hub.Register(client) {
		h.logger.Warn("failed to register websocket client, hub channel full",
			zap.String("user_id", claims.UserID),
			zap.String("bot_id", botID),
		)
		conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseTryAgainLater, "server busy"))
		conn.Close()
		return
	}

	go client.WritePump()
	go client.ReadPump()
}

// Metrics handles GET /api/v1/ws/metrics.
// Returns real-time WebSocket connection statistics.
func (h *WSHandler) Metrics(c *gin.Context) {
	snapshot := h.hub.GetMetrics().Snapshot()
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data":    snapshot,
	})
}
