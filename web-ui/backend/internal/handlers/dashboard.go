package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/DimaJoyti/go-coffee/web-ui/backend/internal/services"
)

type DashboardHandler struct {
	service *services.DashboardService
}

func NewDashboardHandler(service *services.DashboardService) *DashboardHandler {
	return &DashboardHandler{
		service: service,
	}
}

func (h *DashboardHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.service.GetMetrics()
	if err != nil {
		response := map[string]interface{}{
			"error": "Failed to get dashboard metrics",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data":    metrics,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *DashboardHandler) GetActivity(w http.ResponseWriter, r *http.Request) {
	activity, err := h.service.GetActivity()
	if err != nil {
		response := map[string]interface{}{
			"error": "Failed to get activity feed",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data":    activity,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
