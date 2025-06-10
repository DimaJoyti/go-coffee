package handlers

import (
	"net/http"

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

func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	h.hub.ServeWS(w, r)
}
