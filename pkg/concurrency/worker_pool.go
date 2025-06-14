package concurrency

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// DynamicWorkerPool provides auto-scaling worker pool functionality
type DynamicWorkerPool struct {
	name            string
	logger          *zap.Logger
	config          *WorkerPoolConfig
	
	// Worker management
	workers         map[int]*Worker
	workersMu       sync.RWMutex
	nextWorkerID    int64
	
	// Queue management
	jobQueue        chan Job
	resultQueue     chan JobResult
	
	// Scaling management
	currentWorkers  int64
	targetWorkers   int64
	lastScaleTime   time.Time
	scaleMu         sync.RWMutex
	
	// Metrics
	metrics         *WorkerPoolMetrics
	
	// Lifecycle
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
	running         bool
	runningMu       sync.RWMutex
}

// WorkerPoolConfig contains worker pool configuration
type WorkerPoolConfig struct {
	MinWorkers          int           `json:"min_workers"`
	MaxWorkers          int           `json:"max_workers"`
	QueueSize           int           `json:"queue_size"`
	WorkerTimeout       time.Duration `json:"worker_timeout"`
	ScaleUpThreshold    float64       `json:"scale_up_threshold"`
	ScaleDownThreshold  float64       `json:"scale_down_threshold"`
	ScaleUpCooldown     time.Duration `json:"scale_up_cooldown"`
	ScaleDownCooldown   time.Duration `json:"scale_down_cooldown"`
	HealthCheckInterval time.Duration `json:"health_check_interval"`
	MetricsInterval     time.Duration `json:"metrics_interval"`
}

// Job represents a unit of work
type Job struct {
	ID       string
	Type     string
	Payload  interface{}
	Priority int
	Timeout  time.Duration
	Retry    int
	MaxRetry int
}

// JobResult represents the result of job execution
type JobResult struct {
	JobID     string
	Success   bool
	Result    interface{}
	Error     error
	Duration  time.Duration
	WorkerID  int
}

// Worker represents a single worker
type Worker struct {
	ID       int
	pool     *DynamicWorkerPool
	jobChan  chan Job
	quit     chan bool
	active   bool
	lastUsed time.Time
	metrics  *WorkerMetrics
}

// WorkerPoolMetrics tracks pool performance
type WorkerPoolMetrics struct {
	TotalJobs       int64         `json:"total_jobs"`
	CompletedJobs   int64         `json:"completed_jobs"`
	FailedJobs      int64         `json:"failed_jobs"`
	QueueDepth      int64         `json:"queue_depth"`
	ActiveWorkers   int64         `json:"active_workers"`
	IdleWorkers     int64         `json:"idle_workers"`
	AvgJobDuration  time.Duration `json:"avg_job_duration"`
	ThroughputPerSec float64      `json:"throughput_per_sec"`
	mu              sync.RWMutex
}

// WorkerMetrics tracks individual worker performance
type WorkerMetrics struct {
	JobsProcessed int64         `json:"jobs_processed"`
	TotalDuration time.Duration `json:"total_duration"`
	LastJobTime   time.Time     `json:"last_job_time"`
	ErrorCount    int64         `json:"error_count"`
}

// NewDynamicWorkerPool creates a new dynamic worker pool
func NewDynamicWorkerPool(name string, config *WorkerPoolConfig, logger *zap.Logger) *DynamicWorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	
	pool := &DynamicWorkerPool{
		name:           name,
		logger:         logger,
		config:         config,
		workers:        make(map[int]*Worker),
		jobQueue:       make(chan Job, config.QueueSize),
		resultQueue:    make(chan JobResult, config.QueueSize),
		currentWorkers: 0,
		targetWorkers:  int64(config.MinWorkers),
		lastScaleTime:  time.Now(),
		metrics:        &WorkerPoolMetrics{},
		ctx:            ctx,
		cancel:         cancel,
	}
	
	return pool
}

// Start starts the worker pool
func (p *DynamicWorkerPool) Start() error {
	p.runningMu.Lock()
	defer p.runningMu.Unlock()
	
	if p.running {
		return fmt.Errorf("worker pool %s is already running", p.name)
	}
	
	p.logger.Info("Starting dynamic worker pool",
		zap.String("pool", p.name),
		zap.Int("min_workers", p.config.MinWorkers),
		zap.Int("max_workers", p.config.MaxWorkers))
	
	// Start initial workers
	for i := 0; i < p.config.MinWorkers; i++ {
		p.addWorker()
	}
	
	// Start scaling manager
	p.wg.Add(1)
	go p.scalingManager()
	
	// Start metrics collector
	p.wg.Add(1)
	go p.metricsCollector()
	
	// Start result processor
	p.wg.Add(1)
	go p.resultProcessor()
	
	p.running = true
	p.logger.Info("Dynamic worker pool started", zap.String("pool", p.name))
	
	return nil
}

// Stop stops the worker pool gracefully
func (p *DynamicWorkerPool) Stop() error {
	p.runningMu.Lock()
	defer p.runningMu.Unlock()
	
	if !p.running {
		return nil
	}
	
	p.logger.Info("Stopping dynamic worker pool", zap.String("pool", p.name))
	
	// Cancel context to signal shutdown
	p.cancel()
	
	// Close job queue
	close(p.jobQueue)
	
	// Stop all workers
	p.workersMu.Lock()
	for _, worker := range p.workers {
		close(worker.quit)
	}
	p.workersMu.Unlock()
	
	// Wait for all goroutines to finish
	p.wg.Wait()
	
	// Close result queue
	close(p.resultQueue)
	
	p.running = false
	p.logger.Info("Dynamic worker pool stopped", zap.String("pool", p.name))
	
	return nil
}

// SubmitJob submits a job to the worker pool
func (p *DynamicWorkerPool) SubmitJob(job Job) error {
	p.runningMu.RLock()
	defer p.runningMu.RUnlock()
	
	if !p.running {
		return fmt.Errorf("worker pool %s is not running", p.name)
	}
	
	select {
	case p.jobQueue <- job:
		atomic.AddInt64(&p.metrics.TotalJobs, 1)
		atomic.StoreInt64(&p.metrics.QueueDepth, int64(len(p.jobQueue)))
		return nil
	case <-p.ctx.Done():
		return fmt.Errorf("worker pool %s is shutting down", p.name)
	default:
		return fmt.Errorf("job queue is full for pool %s", p.name)
	}
}

// GetResults returns the result channel
func (p *DynamicWorkerPool) GetResults() <-chan JobResult {
	return p.resultQueue
}

// GetMetrics returns current pool metrics
func (p *DynamicWorkerPool) GetMetrics() *WorkerPoolMetrics {
	p.metrics.mu.RLock()
	defer p.metrics.mu.RUnlock()
	
	// Create a copy to avoid race conditions
	metrics := &WorkerPoolMetrics{
		TotalJobs:       atomic.LoadInt64(&p.metrics.TotalJobs),
		CompletedJobs:   atomic.LoadInt64(&p.metrics.CompletedJobs),
		FailedJobs:      atomic.LoadInt64(&p.metrics.FailedJobs),
		QueueDepth:      atomic.LoadInt64(&p.metrics.QueueDepth),
		ActiveWorkers:   atomic.LoadInt64(&p.currentWorkers),
		IdleWorkers:     atomic.LoadInt64(&p.currentWorkers) - p.getActiveWorkerCount(),
		AvgJobDuration:  p.metrics.AvgJobDuration,
		ThroughputPerSec: p.metrics.ThroughputPerSec,
	}
	
	return metrics
}

// addWorker adds a new worker to the pool
func (p *DynamicWorkerPool) addWorker() {
	workerID := int(atomic.AddInt64(&p.nextWorkerID, 1))
	
	worker := &Worker{
		ID:       workerID,
		pool:     p,
		jobChan:  make(chan Job, 1),
		quit:     make(chan bool),
		active:   false,
		lastUsed: time.Now(),
		metrics:  &WorkerMetrics{},
	}
	
	p.workersMu.Lock()
	p.workers[workerID] = worker
	p.workersMu.Unlock()
	
	atomic.AddInt64(&p.currentWorkers, 1)
	
	// Start worker goroutine
	p.wg.Add(1)
	go worker.start()
	
	p.logger.Debug("Added worker to pool",
		zap.String("pool", p.name),
		zap.Int("worker_id", workerID),
		zap.Int64("total_workers", atomic.LoadInt64(&p.currentWorkers)))
}

// removeWorker removes idle workers from the pool until the minimum worker count is reached or no more idle workers are available
func (p *DynamicWorkerPool) removeWorker() {
	p.workersMu.Lock()
	defer p.workersMu.Unlock()

	removed := 0
	for {
		if int64(len(p.workers)) <= int64(p.config.MinWorkers) {
			break
		}
		idleFound := false
		for id, worker := range p.workers {
			if !worker.active && time.Since(worker.lastUsed) > p.config.ScaleDownCooldown {
				close(worker.quit)
				delete(p.workers, id)
				atomic.AddInt64(&p.currentWorkers, -1)
				removed++
				p.logger.Debug("Removed worker from pool",
					zap.String("pool", p.name),
					zap.Int("worker_id", id),
					zap.Int64("total_workers", atomic.LoadInt64(&p.currentWorkers)))
				idleFound = true
				break // break inner loop to re-check worker count and continue removing if possible
			}
		}
		if !idleFound {
			break // no more idle workers to remove
		}
	}
}

// scalingManager manages automatic scaling of the worker pool
func (p *DynamicWorkerPool) scalingManager() {
	defer p.wg.Done()
	
	ticker := time.NewTicker(p.config.HealthCheckInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			p.evaluateScaling()
		case <-p.ctx.Done():
			return
		}
	}
}

// evaluateScaling evaluates whether to scale up or down
func (p *DynamicWorkerPool) evaluateScaling() {
	p.scaleMu.Lock()
	defer p.scaleMu.Unlock()
	
	queueDepth := float64(len(p.jobQueue))
	queueCapacity := float64(p.config.QueueSize)
	currentWorkers := atomic.LoadInt64(&p.currentWorkers)
	
	queueUtilization := queueDepth / queueCapacity
	
	now := time.Now()
	
	// Scale up if queue utilization is high
	if queueUtilization > p.config.ScaleUpThreshold &&
		currentWorkers < int64(p.config.MaxWorkers) &&
		now.Sub(p.lastScaleTime) > p.config.ScaleUpCooldown {
		
		p.addWorker()
		p.lastScaleTime = now
		
		p.logger.Info("Scaled up worker pool",
			zap.String("pool", p.name),
			zap.Float64("queue_utilization", queueUtilization),
			zap.Int64("workers", atomic.LoadInt64(&p.currentWorkers)))
	}
	
	// Scale down if queue utilization is low
	if queueUtilization < p.config.ScaleDownThreshold &&
		currentWorkers > int64(p.config.MinWorkers) &&
		now.Sub(p.lastScaleTime) > p.config.ScaleDownCooldown {
		
		p.removeWorker()
		p.lastScaleTime = now
		
		p.logger.Info("Scaled down worker pool",
			zap.String("pool", p.name),
			zap.Float64("queue_utilization", queueUtilization),
			zap.Int64("workers", atomic.LoadInt64(&p.currentWorkers)))
	}
}

// metricsCollector collects and updates pool metrics
func (p *DynamicWorkerPool) metricsCollector() {
	defer p.wg.Done()
	
	ticker := time.NewTicker(p.config.MetricsInterval)
	defer ticker.Stop()
	
	var lastCompletedJobs int64
	var lastTime time.Time = time.Now()
	
	for {
		select {
		case <-ticker.C:
			now := time.Now()
			currentCompletedJobs := atomic.LoadInt64(&p.metrics.CompletedJobs)
			
			if !lastTime.IsZero() {
				duration := now.Sub(lastTime).Seconds()
				jobsDiff := currentCompletedJobs - lastCompletedJobs
				
				p.metrics.mu.Lock()
				p.metrics.ThroughputPerSec = float64(jobsDiff) / duration
				p.metrics.mu.Unlock()
			}
			
			lastCompletedJobs = currentCompletedJobs
			lastTime = now
			
		case <-p.ctx.Done():
			return
		}
	}
}

// resultProcessor processes job results
func (p *DynamicWorkerPool) resultProcessor() {
	defer p.wg.Done()
	
	for {
		select {
		case result := <-p.resultQueue:
			if result.Success {
				atomic.AddInt64(&p.metrics.CompletedJobs, 1)
			} else {
				atomic.AddInt64(&p.metrics.FailedJobs, 1)
			}
			
			// Update average job duration
			p.metrics.mu.Lock()
			totalJobs := atomic.LoadInt64(&p.metrics.CompletedJobs) + atomic.LoadInt64(&p.metrics.FailedJobs)
			if totalJobs > 0 {
				p.metrics.AvgJobDuration = (p.metrics.AvgJobDuration*time.Duration(totalJobs-1) + result.Duration) / time.Duration(totalJobs)
			}
			p.metrics.mu.Unlock()
			
		case <-p.ctx.Done():
			return
		}
	}
}

// getActiveWorkerCount returns the number of currently active workers
func (p *DynamicWorkerPool) getActiveWorkerCount() int64 {
	p.workersMu.RLock()
	defer p.workersMu.RUnlock()
	
	var activeCount int64
	for _, worker := range p.workers {
		if worker.active {
			activeCount++
		}
	}
	
	return activeCount
}

// Worker methods

// start starts the worker
func (w *Worker) start() {
	defer w.pool.wg.Done()
	
	w.pool.logger.Debug("Worker started",
		zap.String("pool", w.pool.name),
		zap.Int("worker_id", w.ID))
	
	for {
		select {
		case job := <-w.pool.jobQueue:
			w.processJob(job)
		case <-w.quit:
			w.pool.logger.Debug("Worker stopped",
				zap.String("pool", w.pool.name),
				zap.Int("worker_id", w.ID))
			return
		case <-w.pool.ctx.Done():
			return
		}
	}
}

// processJob processes a single job
func (w *Worker) processJob(job Job) {
	w.active = true
	w.lastUsed = time.Now()
	start := time.Now()
	
	defer func() {
		w.active = false
		duration := time.Since(start)
		
		w.metrics.JobsProcessed++
		w.metrics.TotalDuration += duration
		w.metrics.LastJobTime = time.Now()
	}()
	
	// Create job context with timeout
	ctx := w.pool.ctx
	if job.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, job.Timeout)
		defer cancel()
	}
	
	// Process the job
	result := JobResult{
		JobID:    job.ID,
		WorkerID: w.ID,
		Duration: time.Since(start),
	}
	
	// Simulate job processing (replace with actual job logic)
	select {
	case <-time.After(time.Millisecond * time.Duration(50+w.ID%100)): // Simulate work
		result.Success = true
		result.Result = fmt.Sprintf("Job %s completed by worker %d", job.ID, w.ID)
	case <-ctx.Done():
		result.Success = false
		result.Error = ctx.Err()
		w.metrics.ErrorCount++
	}
	
	result.Duration = time.Since(start)
	
	// Send result
	select {
	case w.pool.resultQueue <- result:
	case <-w.pool.ctx.Done():
	}
}
