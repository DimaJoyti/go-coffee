package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// LLMWorkload represents a simplified LLM workload
type LLMWorkload struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	ModelName string            `json:"modelName"`
	ModelType string            `json:"modelType"`
	Resources ResourceSpec      `json:"resources"`
	Status    WorkloadStatus    `json:"status"`
	Metrics   WorkloadMetrics   `json:"metrics"`
	CreatedAt time.Time         `json:"createdAt"`
	UpdatedAt time.Time         `json:"updatedAt"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// ResourceSpec defines resource requirements
type ResourceSpec struct {
	CPU    string `json:"cpu"`    // e.g., "2000m"
	Memory string `json:"memory"` // e.g., "8Gi"
	GPU    int    `json:"gpu"`    // number of GPUs
}

// WorkloadStatus represents the current status
type WorkloadStatus struct {
	Phase         string    `json:"phase"` // pending, running, failed, completed
	Message       string    `json:"message,omitempty"`
	Replicas      int       `json:"replicas"`
	ReadyReplicas int       `json:"readyReplicas"`
	LastUpdated   time.Time `json:"lastUpdated"`
}

// WorkloadMetrics contains performance metrics
type WorkloadMetrics struct {
	RequestsPerSecond float64       `json:"requestsPerSecond"`
	AverageLatency    time.Duration `json:"averageLatency"`
	ErrorRate         float64       `json:"errorRate"`
	CPUUsage          float64       `json:"cpuUsage"`
	MemoryUsage       float64       `json:"memoryUsage"`
	GPUUsage          float64       `json:"gpuUsage"`
	LastUpdated       time.Time     `json:"lastUpdated"`
}

// LLMOrchestrator manages LLM workloads
type LLMOrchestrator struct {
	port      int
	workloads map[string]*LLMWorkload
	mutex     sync.RWMutex
	logger    *log.Logger
}

// NewLLMOrchestrator creates a new orchestrator instance
func NewLLMOrchestrator(port int) *LLMOrchestrator {
	return &LLMOrchestrator{
		port:      port,
		workloads: make(map[string]*LLMWorkload),
		logger:    log.New(os.Stdout, "[LLM-ORCHESTRATOR] ", log.LstdFlags),
	}
}

// Start starts the orchestrator
func (o *LLMOrchestrator) Start(ctx context.Context) error {
	o.logger.Printf("Starting LLM Orchestrator on port %d", o.port)

	// Start metrics monitoring
	go o.startMetricsCollection(ctx)

	// Setup HTTP server
	mux := http.NewServeMux()
	o.setupRoutes(mux)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", o.port),
		Handler: mux,
	}

	// Start server in goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			o.logger.Printf("Server failed: %v", err)
		}
	}()

	o.logger.Println("LLM Orchestrator started successfully")

	// Wait for context cancellation
	<-ctx.Done()

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return server.Shutdown(shutdownCtx)
}

// setupRoutes configures HTTP routes
func (o *LLMOrchestrator) setupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", o.handleHealth)
	mux.HandleFunc("/metrics", o.handleMetrics)
	mux.HandleFunc("/workloads", o.handleWorkloads)
	mux.HandleFunc("/workloads/", o.handleWorkloadByID)
	mux.HandleFunc("/schedule", o.handleSchedule)
	mux.HandleFunc("/status", o.handleStatus)
}

// HTTP Handlers
func (o *LLMOrchestrator) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   "1.0.0",
	})
}

func (o *LLMOrchestrator) handleMetrics(w http.ResponseWriter, r *http.Request) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	metrics := map[string]interface{}{
		"totalWorkloads":   len(o.workloads),
		"runningWorkloads": o.countWorkloadsByStatus("running"),
		"pendingWorkloads": o.countWorkloadsByStatus("pending"),
		"failedWorkloads":  o.countWorkloadsByStatus("failed"),
		"timestamp":        time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func (o *LLMOrchestrator) handleWorkloads(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		o.listWorkloads(w, r)
	case http.MethodPost:
		o.createWorkload(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (o *LLMOrchestrator) handleWorkloadByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/workloads/"):]
	if id == "" {
		http.Error(w, "Workload ID required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		o.getWorkload(w, r, id)
	case http.MethodDelete:
		o.deleteWorkload(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (o *LLMOrchestrator) handleSchedule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		WorkloadID string `json:"workloadId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result := map[string]interface{}{
		"workloadId":     req.WorkloadID,
		"scheduledNode":  "node-1",
		"schedulingTime": time.Now().Format(time.RFC3339),
		"status":         "scheduled",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (o *LLMOrchestrator) handleStatus(w http.ResponseWriter, r *http.Request) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	status := map[string]interface{}{
		"orchestrator": "running",
		"workloads":    len(o.workloads),
		"uptime":       "running", // Simplified
		"version":      "1.0.0",
		"timestamp":    time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// Workload management methods
func (o *LLMOrchestrator) listWorkloads(w http.ResponseWriter, r *http.Request) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	workloads := make([]*LLMWorkload, 0, len(o.workloads))
	for _, workload := range o.workloads {
		workloads = append(workloads, workload)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workloads)
}

func (o *LLMOrchestrator) createWorkload(w http.ResponseWriter, r *http.Request) {
	var workload LLMWorkload
	if err := json.NewDecoder(r.Body).Decode(&workload); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set defaults
	if workload.ID == "" {
		workload.ID = fmt.Sprintf("workload-%d", time.Now().Unix())
	}
	if workload.Resources.CPU == "" {
		workload.Resources.CPU = "1000m"
	}
	if workload.Resources.Memory == "" {
		workload.Resources.Memory = "2Gi"
	}

	workload.CreatedAt = time.Now()
	workload.UpdatedAt = time.Now()
	workload.Status = WorkloadStatus{
		Phase:         "pending",
		Replicas:      1,
		ReadyReplicas: 0,
		LastUpdated:   time.Now(),
	}

	o.mutex.Lock()
	o.workloads[workload.ID] = &workload
	o.mutex.Unlock()

	o.logger.Printf("Created workload: %s (%s)", workload.ID, workload.Name)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(&workload)
}

func (o *LLMOrchestrator) getWorkload(w http.ResponseWriter, r *http.Request, id string) {
	o.mutex.RLock()
	workload, exists := o.workloads[id]
	o.mutex.RUnlock()

	if !exists {
		http.Error(w, "Workload not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workload)
}

func (o *LLMOrchestrator) deleteWorkload(w http.ResponseWriter, r *http.Request, id string) {
	o.mutex.Lock()
	_, exists := o.workloads[id]
	if exists {
		delete(o.workloads, id)
	}
	o.mutex.Unlock()

	if !exists {
		http.Error(w, "Workload not found", http.StatusNotFound)
		return
	}

	o.logger.Printf("Deleted workload: %s", id)
	w.WriteHeader(http.StatusNoContent)
}

// Helper methods
func (o *LLMOrchestrator) countWorkloadsByStatus(status string) int {
	count := 0
	for _, workload := range o.workloads {
		if workload.Status.Phase == status {
			count++
		}
	}
	return count
}

// startMetricsCollection starts metrics collection
func (o *LLMOrchestrator) startMetricsCollection(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			o.collectMetrics()
		}
	}
}

// collectMetrics collects current metrics from all workloads
func (o *LLMOrchestrator) collectMetrics() {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	// Update workload metrics (simulated)
	for _, workload := range o.workloads {
		workload.Metrics = WorkloadMetrics{
			RequestsPerSecond: 100.0 + float64(time.Now().Unix()%50),
			AverageLatency:    time.Duration(50+time.Now().Unix()%100) * time.Millisecond,
			ErrorRate:         0.01,
			CPUUsage:          0.5 + float64(time.Now().Unix()%30)/100,
			MemoryUsage:       0.6 + float64(time.Now().Unix()%20)/100,
			GPUUsage:          0.8 + float64(time.Now().Unix()%15)/100,
			LastUpdated:       time.Now(),
		}

		// Simulate status updates
		if workload.Status.Phase == "pending" && time.Since(workload.CreatedAt) > 10*time.Second {
			workload.Status.Phase = "running"
			workload.Status.ReadyReplicas = workload.Status.Replicas
			workload.Status.LastUpdated = time.Now()
		}
	}
}

func main() {
	var (
		port = flag.Int("port", 8080, "Port to listen on")
	)
	flag.Parse()

	// Create orchestrator
	orchestrator := NewLLMOrchestrator(*port)

	// Setup signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Println("Received shutdown signal")
		cancel()
	}()

	// Start orchestrator
	if err := orchestrator.Start(ctx); err != nil {
		log.Fatalf("Orchestrator failed: %v", err)
	}

	log.Println("LLM Orchestrator shut down successfully")
}
