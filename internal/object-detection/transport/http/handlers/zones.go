package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/object-detection/domain"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// ZoneHandlers handles HTTP requests for zone management
type ZoneHandlers struct {
	logger      *zap.Logger
	zoneService domain.ZoneService
}

// NewZoneHandlers creates new zone handlers
func NewZoneHandlers(logger *zap.Logger, zoneService domain.ZoneService) *ZoneHandlers {
	return &ZoneHandlers{
		logger:      logger.With(zap.String("component", "zone_handlers")),
		zoneService: zoneService,
	}
}

// CreateZone handles POST /api/v1/zones
func (h *ZoneHandlers) CreateZone(w http.ResponseWriter, r *http.Request) {
	var zone domain.DetectionZone
	if err := json.NewDecoder(r.Body).Decode(&zone); err != nil {
		h.logger.Error("Failed to decode zone request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.zoneService.CreateZone(&zone); err != nil {
		h.logger.Error("Failed to create zone", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to create zone: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"zone":    zone,
	})

	h.logger.Info("Zone created",
		zap.String("zone_id", zone.ID),
		zap.String("stream_id", zone.StreamID),
		zap.String("name", zone.Name))
}

// GetZone handles GET /api/v1/zones/{id}
func (h *ZoneHandlers) GetZone(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	zoneID := vars["id"]

	zone, err := h.zoneService.GetZone(zoneID)
	if err != nil {
		h.logger.Error("Failed to get zone", zap.String("zone_id", zoneID), zap.Error(err))
		http.Error(w, "Zone not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"zone":    zone,
	})
}

// GetZonesByStream handles GET /api/v1/streams/{streamId}/zones
func (h *ZoneHandlers) GetZonesByStream(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	streamID := vars["streamId"]

	zones, err := h.zoneService.GetZonesByStream(streamID)
	if err != nil {
		h.logger.Error("Failed to get zones for stream", zap.String("stream_id", streamID), zap.Error(err))
		http.Error(w, "Failed to get zones", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"zones":   zones,
		"count":   len(zones),
	})
}

// UpdateZone handles PUT /api/v1/zones/{id}
func (h *ZoneHandlers) UpdateZone(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	zoneID := vars["id"]

	var zone domain.DetectionZone
	if err := json.NewDecoder(r.Body).Decode(&zone); err != nil {
		h.logger.Error("Failed to decode zone update request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Ensure the ID matches the URL parameter
	zone.ID = zoneID

	if err := h.zoneService.UpdateZone(&zone); err != nil {
		h.logger.Error("Failed to update zone", zap.String("zone_id", zoneID), zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to update zone: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"zone":    zone,
	})

	h.logger.Info("Zone updated", zap.String("zone_id", zoneID))
}

// DeleteZone handles DELETE /api/v1/zones/{id}
func (h *ZoneHandlers) DeleteZone(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	zoneID := vars["id"]

	if err := h.zoneService.DeleteZone(zoneID); err != nil {
		h.logger.Error("Failed to delete zone", zap.String("zone_id", zoneID), zap.Error(err))
		http.Error(w, "Failed to delete zone", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Zone deleted successfully",
	})

	h.logger.Info("Zone deleted", zap.String("zone_id", zoneID))
}

// ListZones handles GET /api/v1/zones
func (h *ZoneHandlers) ListZones(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50 // Default limit
	offset := 0 // Default offset

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	zones, err := h.zoneService.ListZones(limit, offset)
	if err != nil {
		h.logger.Error("Failed to list zones", zap.Error(err))
		http.Error(w, "Failed to list zones", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"zones":   zones,
		"count":   len(zones),
		"limit":   limit,
		"offset":  offset,
	})
}

// GetZoneStatistics handles GET /api/v1/zones/{id}/statistics
func (h *ZoneHandlers) GetZoneStatistics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	zoneID := vars["id"]

	stats, err := h.zoneService.GetZoneStatistics(zoneID)
	if err != nil {
		h.logger.Error("Failed to get zone statistics", zap.String("zone_id", zoneID), zap.Error(err))
		http.Error(w, "Failed to get zone statistics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"statistics": stats,
	})
}

// GetZoneAnalytics handles GET /api/v1/zones/{id}/analytics
func (h *ZoneHandlers) GetZoneAnalytics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	zoneID := vars["id"]

	// Parse time range parameters
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	var timeRange domain.TimeRange
	var err error

	if startStr != "" {
		timeRange.Start, err = time.Parse(time.RFC3339, startStr)
		if err != nil {
			http.Error(w, "Invalid start time format", http.StatusBadRequest)
			return
		}
	} else {
		// Default to last 24 hours
		timeRange.Start = time.Now().Add(-24 * time.Hour)
	}

	if endStr != "" {
		timeRange.End, err = time.Parse(time.RFC3339, endStr)
		if err != nil {
			http.Error(w, "Invalid end time format", http.StatusBadRequest)
			return
		}
	} else {
		timeRange.End = time.Now()
	}

	analytics, err := h.zoneService.GetZoneAnalytics(zoneID, timeRange)
	if err != nil {
		h.logger.Error("Failed to get zone analytics", zap.String("zone_id", zoneID), zap.Error(err))
		http.Error(w, "Failed to get zone analytics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"analytics": analytics,
	})
}

// GenerateZoneReport handles POST /api/v1/zones/{id}/reports
func (h *ZoneHandlers) GenerateZoneReport(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	zoneID := vars["id"]

	var request struct {
		ReportType string             `json:"report_type"`
		TimeRange  domain.TimeRange   `json:"time_range"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error("Failed to decode report request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	reportType := domain.ReportType(request.ReportType)
	
	// Validate report type
	validTypes := []domain.ReportType{
		domain.ReportTypeOccupancy,
		domain.ReportTypeTraffic,
		domain.ReportTypeViolations,
		domain.ReportTypeDwellTime,
		domain.ReportTypeHeatMap,
		domain.ReportTypeSummary,
	}
	
	valid := false
	for _, validType := range validTypes {
		if reportType == validType {
			valid = true
			break
		}
	}
	
	if !valid {
		http.Error(w, "Invalid report type", http.StatusBadRequest)
		return
	}

	report, err := h.zoneService.GenerateZoneReport(zoneID, reportType, request.TimeRange)
	if err != nil {
		h.logger.Error("Failed to generate zone report", 
			zap.String("zone_id", zoneID),
			zap.String("report_type", string(reportType)),
			zap.Error(err))
		http.Error(w, "Failed to generate report", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"report":  report,
	})

	h.logger.Info("Zone report generated",
		zap.String("zone_id", zoneID),
		zap.String("report_type", string(reportType)),
		zap.String("report_id", report.ID))
}

// TestZonePoint handles POST /api/v1/zones/{id}/test-point
func (h *ZoneHandlers) TestZonePoint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	zoneID := vars["id"]

	var request struct {
		Point domain.Point `json:"point"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error("Failed to decode test point request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	zone, err := h.zoneService.GetZone(zoneID)
	if err != nil {
		h.logger.Error("Failed to get zone for point test", zap.String("zone_id", zoneID), zap.Error(err))
		http.Error(w, "Zone not found", http.StatusNotFound)
		return
	}

	isInside := zone.Polygon.Contains(request.Point)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"zone_id":   zoneID,
		"point":     request.Point,
		"is_inside": isInside,
	})
}

// RegisterZoneRoutes registers all zone-related routes
func (h *ZoneHandlers) RegisterRoutes(router *mux.Router) {
	// Zone management
	router.HandleFunc("/api/v1/zones", h.CreateZone).Methods("POST")
	router.HandleFunc("/api/v1/zones", h.ListZones).Methods("GET")
	router.HandleFunc("/api/v1/zones/{id}", h.GetZone).Methods("GET")
	router.HandleFunc("/api/v1/zones/{id}", h.UpdateZone).Methods("PUT")
	router.HandleFunc("/api/v1/zones/{id}", h.DeleteZone).Methods("DELETE")

	// Zone analytics and statistics
	router.HandleFunc("/api/v1/zones/{id}/statistics", h.GetZoneStatistics).Methods("GET")
	router.HandleFunc("/api/v1/zones/{id}/analytics", h.GetZoneAnalytics).Methods("GET")
	router.HandleFunc("/api/v1/zones/{id}/reports", h.GenerateZoneReport).Methods("POST")

	// Zone utilities
	router.HandleFunc("/api/v1/zones/{id}/test-point", h.TestZonePoint).Methods("POST")

	// Stream-specific zones
	router.HandleFunc("/api/v1/streams/{streamId}/zones", h.GetZonesByStream).Methods("GET")
}
