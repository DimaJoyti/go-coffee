package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/kitchen/application"
	"github.com/DimaJoyti/go-coffee/internal/kitchen/domain"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
	"github.com/gorilla/mux"
)

// Handler represents HTTP handlers for kitchen service
type Handler struct {
	kitchenService      application.KitchenService
	queueService        application.QueueService
	optimizerService    application.OptimizerService
	notificationService application.NotificationService
	logger              *logger.Logger
}

// NewHandler creates a new HTTP handler
func NewHandler(
	kitchenService application.KitchenService,
	queueService application.QueueService,
	optimizerService application.OptimizerService,
	notificationService application.NotificationService,
	logger *logger.Logger,
) *Handler {
	return &Handler{
		kitchenService:      kitchenService,
		queueService:        queueService,
		optimizerService:    optimizerService,
		notificationService: notificationService,
		logger:              logger,
	}
}

// RegisterRoutes registers HTTP routes
func (h *Handler) RegisterRoutes(router *mux.Router) {
	// Equipment routes
	router.HandleFunc("/api/v1/kitchen/equipment", h.CreateEquipment).Methods("POST")
	router.HandleFunc("/api/v1/kitchen/equipment", h.ListEquipment).Methods("GET")
	router.HandleFunc("/api/v1/kitchen/equipment/{id}", h.GetEquipment).Methods("GET")
	router.HandleFunc("/api/v1/kitchen/equipment/{id}/status", h.UpdateEquipmentStatus).Methods("PUT")
	router.HandleFunc("/api/v1/kitchen/equipment/{id}/maintenance", h.ScheduleEquipmentMaintenance).Methods("POST")

	// Staff routes
	router.HandleFunc("/api/v1/kitchen/staff", h.CreateStaff).Methods("POST")
	router.HandleFunc("/api/v1/kitchen/staff", h.ListStaff).Methods("GET")
	router.HandleFunc("/api/v1/kitchen/staff/{id}", h.GetStaff).Methods("GET")
	router.HandleFunc("/api/v1/kitchen/staff/{id}/availability", h.UpdateStaffAvailability).Methods("PUT")
	router.HandleFunc("/api/v1/kitchen/staff/{id}/skill", h.UpdateStaffSkill).Methods("PUT")

	// Order routes
	router.HandleFunc("/api/v1/kitchen/orders", h.AddOrderToQueue).Methods("POST")
	router.HandleFunc("/api/v1/kitchen/orders/{id}", h.GetOrder).Methods("GET")
	router.HandleFunc("/api/v1/kitchen/orders/{id}/status", h.UpdateOrderStatus).Methods("PUT")
	router.HandleFunc("/api/v1/kitchen/orders/{id}/priority", h.UpdateOrderPriority).Methods("PUT")
	router.HandleFunc("/api/v1/kitchen/orders/{id}/assign", h.AssignOrderToStaff).Methods("POST")
	router.HandleFunc("/api/v1/kitchen/orders/{id}/start", h.StartOrderProcessing).Methods("POST")
	router.HandleFunc("/api/v1/kitchen/orders/{id}/complete", h.CompleteOrder).Methods("POST")

	// Queue routes
	router.HandleFunc("/api/v1/kitchen/queue/status", h.GetQueueStatus).Methods("GET")
	router.HandleFunc("/api/v1/kitchen/queue/next", h.GetNextOrder).Methods("GET")
	router.HandleFunc("/api/v1/kitchen/queue/optimize", h.OptimizeQueue).Methods("POST")

	// Analytics routes
	router.HandleFunc("/api/v1/kitchen/metrics", h.GetKitchenMetrics).Methods("GET")
	router.HandleFunc("/api/v1/kitchen/performance", h.GetPerformanceReport).Methods("GET")

	// Health check
	router.HandleFunc("/api/v1/kitchen/health", h.HealthCheck).Methods("GET")
}

// Equipment Handlers

// CreateEquipment creates new kitchen equipment
func (h *Handler) CreateEquipment(w http.ResponseWriter, r *http.Request) {
	var req application.CreateEquipmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	equipment, err := h.kitchenService.CreateEquipment(r.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create equipment")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to create equipment")
		return
	}

	h.respondWithJSON(w, http.StatusCreated, equipment)
}

// GetEquipment retrieves equipment by ID
func (h *Handler) GetEquipment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	equipment, err := h.kitchenService.GetEquipment(r.Context(), id)
	if err != nil {
		h.logger.WithError(err).WithField("equipment_id", id).Error("Failed to get equipment")
		h.respondWithError(w, http.StatusNotFound, "Equipment not found")
		return
	}

	h.respondWithJSON(w, http.StatusOK, equipment)
}

// ListEquipment lists equipment with optional filtering
func (h *Handler) ListEquipment(w http.ResponseWriter, r *http.Request) {
	filter := &application.EquipmentFilter{}

	// Parse query parameters
	if stationType := r.URL.Query().Get("station_type"); stationType != "" {
		if st, err := strconv.Atoi(stationType); err == nil {
			stationTypeEnum := domain.StationType(st)
			filter.StationType = &stationTypeEnum
		}
	}

	if status := r.URL.Query().Get("status"); status != "" {
		if s, err := strconv.Atoi(status); err == nil {
			statusEnum := domain.EquipmentStatus(s)
			filter.Status = &statusEnum
		}
	}

	if available := r.URL.Query().Get("available"); available != "" {
		if a, err := strconv.ParseBool(available); err == nil {
			filter.Available = &a
		}
	}

	if limit := r.URL.Query().Get("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			filter.Limit = int32(l)
		}
	}

	if offset := r.URL.Query().Get("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil {
			filter.Offset = int32(o)
		}
	}

	equipment, err := h.kitchenService.ListEquipment(r.Context(), filter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list equipment")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to list equipment")
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"equipment": equipment,
		"total":     len(equipment),
	})
}

// UpdateEquipmentStatus updates equipment status
func (h *Handler) UpdateEquipmentStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req struct {
		Status domain.EquipmentStatus `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := h.kitchenService.UpdateEquipmentStatus(r.Context(), id, req.Status)
	if err != nil {
		h.logger.WithError(err).WithField("equipment_id", id).Error("Failed to update equipment status")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to update equipment status")
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Equipment status updated successfully",
	})
}

// ScheduleEquipmentMaintenance schedules equipment maintenance
func (h *Handler) ScheduleEquipmentMaintenance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	err := h.kitchenService.ScheduleEquipmentMaintenance(r.Context(), id)
	if err != nil {
		h.logger.WithError(err).WithField("equipment_id", id).Error("Failed to schedule equipment maintenance")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to schedule maintenance")
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Equipment maintenance scheduled successfully",
	})
}

// Staff Handlers

// CreateStaff creates new kitchen staff
func (h *Handler) CreateStaff(w http.ResponseWriter, r *http.Request) {
	var req application.CreateStaffRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	staff, err := h.kitchenService.CreateStaff(r.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create staff")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to create staff")
		return
	}

	h.respondWithJSON(w, http.StatusCreated, staff)
}

// GetStaff retrieves staff by ID
func (h *Handler) GetStaff(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	staff, err := h.kitchenService.GetStaff(r.Context(), id)
	if err != nil {
		h.logger.WithError(err).WithField("staff_id", id).Error("Failed to get staff")
		h.respondWithError(w, http.StatusNotFound, "Staff not found")
		return
	}

	h.respondWithJSON(w, http.StatusOK, staff)
}

// ListStaff lists staff with optional filtering
func (h *Handler) ListStaff(w http.ResponseWriter, r *http.Request) {
	filter := &application.StaffFilter{}

	// Parse query parameters
	if specialization := r.URL.Query().Get("specialization"); specialization != "" {
		if s, err := strconv.Atoi(specialization); err == nil {
			specEnum := domain.StationType(s)
			filter.Specialization = &specEnum
		}
	}

	if available := r.URL.Query().Get("available"); available != "" {
		if a, err := strconv.ParseBool(available); err == nil {
			filter.Available = &a
		}
	}

	if minSkillLevel := r.URL.Query().Get("min_skill_level"); minSkillLevel != "" {
		if msl, err := strconv.ParseFloat(minSkillLevel, 32); err == nil {
			skillLevel := float32(msl)
			filter.MinSkillLevel = &skillLevel
		}
	}

	if limit := r.URL.Query().Get("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			filter.Limit = int32(l)
		}
	}

	if offset := r.URL.Query().Get("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil {
			filter.Offset = int32(o)
		}
	}

	staff, err := h.kitchenService.ListStaff(r.Context(), filter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list staff")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to list staff")
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"staff": staff,
		"total": len(staff),
	})
}

// UpdateStaffAvailability updates staff availability
func (h *Handler) UpdateStaffAvailability(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req struct {
		Available bool `json:"available"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := h.kitchenService.UpdateStaffAvailability(r.Context(), id, req.Available)
	if err != nil {
		h.logger.WithError(err).WithField("staff_id", id).Error("Failed to update staff availability")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to update staff availability")
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Staff availability updated successfully",
	})
}

// UpdateStaffSkill updates staff skill level
func (h *Handler) UpdateStaffSkill(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req struct {
		SkillLevel float32 `json:"skill_level"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := h.kitchenService.UpdateStaffSkill(r.Context(), id, req.SkillLevel)
	if err != nil {
		h.logger.WithError(err).WithField("staff_id", id).Error("Failed to update staff skill")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to update staff skill")
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Staff skill updated successfully",
	})
}

// Order Handlers

// AddOrderToQueue adds an order to the kitchen queue
func (h *Handler) AddOrderToQueue(w http.ResponseWriter, r *http.Request) {
	var req application.AddOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	order, err := h.kitchenService.AddOrderToQueue(r.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to add order to queue")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to add order to queue")
		return
	}

	h.respondWithJSON(w, http.StatusCreated, order)
}

// GetOrder retrieves an order by ID
func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	order, err := h.kitchenService.GetOrder(r.Context(), id)
	if err != nil {
		h.logger.WithError(err).WithField("order_id", id).Error("Failed to get order")
		h.respondWithError(w, http.StatusNotFound, "Order not found")
		return
	}

	h.respondWithJSON(w, http.StatusOK, order)
}

// UpdateOrderStatus updates order status
func (h *Handler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req struct {
		Status domain.OrderStatus `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := h.kitchenService.UpdateOrderStatus(r.Context(), id, req.Status)
	if err != nil {
		h.logger.WithError(err).WithField("order_id", id).Error("Failed to update order status")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to update order status")
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Order status updated successfully",
	})
}

// UpdateOrderPriority updates order priority
func (h *Handler) UpdateOrderPriority(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req struct {
		Priority domain.OrderPriority `json:"priority"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := h.kitchenService.UpdateOrderPriority(r.Context(), id, req.Priority)
	if err != nil {
		h.logger.WithError(err).WithField("order_id", id).Error("Failed to update order priority")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to update order priority")
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Order priority updated successfully",
	})
}

// AssignOrderToStaff assigns an order to a staff member
func (h *Handler) AssignOrderToStaff(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["id"]

	var req struct {
		StaffID string `json:"staff_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := h.kitchenService.AssignOrderToStaff(r.Context(), orderID, req.StaffID)
	if err != nil {
		h.logger.WithError(err).WithFields(map[string]interface{}{
			"order_id": orderID,
			"staff_id": req.StaffID,
		}).Error("Failed to assign order to staff")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to assign order to staff")
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Order assigned to staff successfully",
	})
}

// StartOrderProcessing starts processing an order
func (h *Handler) StartOrderProcessing(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	err := h.kitchenService.StartOrderProcessing(r.Context(), id)
	if err != nil {
		h.logger.WithError(err).WithField("order_id", id).Error("Failed to start order processing")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to start order processing")
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Order processing started successfully",
	})
}

// CompleteOrder completes an order
func (h *Handler) CompleteOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	err := h.kitchenService.CompleteOrder(r.Context(), id)
	if err != nil {
		h.logger.WithError(err).WithField("order_id", id).Error("Failed to complete order")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to complete order")
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Order completed successfully",
	})
}

// Queue Handlers

// GetQueueStatus returns current queue status
func (h *Handler) GetQueueStatus(w http.ResponseWriter, r *http.Request) {
	status, err := h.kitchenService.GetQueueStatus(r.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get queue status")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get queue status")
		return
	}

	h.respondWithJSON(w, http.StatusOK, status)
}

// GetNextOrder returns the next order to be processed
func (h *Handler) GetNextOrder(w http.ResponseWriter, r *http.Request) {
	order, err := h.kitchenService.GetNextOrder(r.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get next order")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get next order")
		return
	}

	response := map[string]interface{}{
		"has_order": order != nil,
	}

	if order != nil {
		response["order"] = order
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// OptimizeQueue optimizes the current queue
func (h *Handler) OptimizeQueue(w http.ResponseWriter, r *http.Request) {
	optimization, err := h.kitchenService.OptimizeQueue(r.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to optimize queue")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to optimize queue")
		return
	}

	h.respondWithJSON(w, http.StatusOK, optimization)
}

// Analytics Handlers

// GetKitchenMetrics returns kitchen performance metrics
func (h *Handler) GetKitchenMetrics(w http.ResponseWriter, r *http.Request) {
	var period *application.TimePeriod

	// Parse time period from query parameters
	if startStr := r.URL.Query().Get("start"); startStr != "" {
		if start, err := time.Parse(time.RFC3339, startStr); err == nil {
			if period == nil {
				period = &application.TimePeriod{}
			}
			period.Start = start
		}
	}

	if endStr := r.URL.Query().Get("end"); endStr != "" {
		if end, err := time.Parse(time.RFC3339, endStr); err == nil {
			if period == nil {
				period = &application.TimePeriod{}
			}
			period.End = end
		}
	}

	metrics, err := h.kitchenService.GetKitchenMetrics(r.Context(), period)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get kitchen metrics")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get kitchen metrics")
		return
	}

	h.respondWithJSON(w, http.StatusOK, metrics)
}

// GetPerformanceReport returns a comprehensive performance report
func (h *Handler) GetPerformanceReport(w http.ResponseWriter, r *http.Request) {
	var period *application.TimePeriod

	// Parse time period from query parameters
	if startStr := r.URL.Query().Get("start"); startStr != "" {
		if start, err := time.Parse(time.RFC3339, startStr); err == nil {
			if period == nil {
				period = &application.TimePeriod{}
			}
			period.Start = start
		}
	}

	if endStr := r.URL.Query().Get("end"); endStr != "" {
		if end, err := time.Parse(time.RFC3339, endStr); err == nil {
			if period == nil {
				period = &application.TimePeriod{}
			}
			period.End = end
		}
	}

	report, err := h.kitchenService.GetPerformanceReport(r.Context(), period)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get performance report")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get performance report")
		return
	}

	h.respondWithJSON(w, http.StatusOK, report)
}

// HealthCheck returns service health status
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"service":   "kitchen-service",
		"timestamp": time.Now(),
	})
}

// Helper methods

func (h *Handler) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, map[string]string{"error": message})
}

func (h *Handler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
