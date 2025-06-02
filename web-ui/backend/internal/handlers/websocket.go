package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/DimaJoyti/go-coffee/web-ui/backend/internal/websocket"
)

type WebSocketHandler struct {
	hub *websocket.Hub
}

func NewWebSocketHandler(hub *websocket.Hub) *WebSocketHandler {
	return &WebSocketHandler{
		hub: hub,
	}
}

func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	h.hub.ServeWS(c.Writer, c.Request)
}
