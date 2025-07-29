package worker

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/IBM/sarama"
)

// Consumer interface for testing
type Consumer interface {
	Subscribe(topics []string) error
	Poll(timeout time.Duration) (interface{}, error)
	Close() error
}

// Processor interface for testing
type Processor interface {
	ProcessMessage(ctx context.Context, message []byte) error
}

// Message interface for testing
type Message interface {
	Value() []byte
}

// Order struct
type Order struct {
	CustomerName string `json:"customer_name"`
	CoffeeType   string `json:"coffee_type"`
}

// ProcessedOrder represents a processed coffee order
type ProcessedOrder struct {
	OrderID         string    `json:"order_id"`
	CustomerName    string    `json:"customer_name"`
	CoffeeType      string    `json:"coffee_type"`
	Status          string    `json:"status"`
	ProcessedAt     time.Time `json:"processed_at"`
	PreparationTime int       `json:"preparation_time"` // in seconds
}

// Worker represents a worker that processes messages
type Worker struct {
	id         int
	jobQueue   <-chan *sarama.ConsumerMessage
	quit       chan bool
	wg         *sync.WaitGroup
	workerPool int
	// Fields for test compatibility
	consumer  Consumer
	processor Processor
	topics    []string
	stopChan  chan struct{}
	running   bool
}

// NewWorker creates a new worker
func NewWorker(id int, jobQueue <-chan *sarama.ConsumerMessage, wg *sync.WaitGroup, workerPool int) *Worker {
	return &Worker{
		id:         id,
		jobQueue:   jobQueue,
		quit:       make(chan bool),
		wg:         wg,
		workerPool: workerPool,
	}
}

// Start starts the worker
func (w *Worker) Start(ctx ...context.Context) error {
	// Handle both old and new signatures for backward compatibility
	var workCtx context.Context
	if len(ctx) > 0 {
		workCtx = ctx[0]
	} else {
		workCtx = context.Background()
	}

	w.running = true

	// If we have a consumer and processor (test mode), use them
	if w.consumer != nil && w.processor != nil {
		if err := w.consumer.Subscribe(w.topics); err != nil {
			return err
		}

		go func() {
			defer func() {
				w.running = false
				// Don't call Close() here, let Stop() handle it
			}()

			for {
				select {
				case <-w.stopChan:
					if w.consumer != nil {
						w.consumer.Close()
					}
					return
				case <-workCtx.Done():
					if w.consumer != nil {
						w.consumer.Close()
					}
					return
				default:
					// Poll for messages
					if w.consumer != nil {
						msg, err := w.consumer.Poll(10 * time.Millisecond) // Shorter timeout for faster processing
						if err != nil {
							log.Printf("Error polling: %v", err)
							continue
						}
						if msg != nil {
							// Process the message - handle different message types
							var msgBytes []byte
							switch m := msg.(type) {
							case []byte:
								msgBytes = m
							case Message:
								if m != nil {
									msgBytes = m.Value()
								}
							default:
								// For other message types, try to extract value
								if mockMsg, ok := msg.(interface{ Value() []byte }); ok {
									if mockMsg != nil {
										msgBytes = mockMsg.Value()
									}
								}
							}
							if len(msgBytes) > 0 {
								w.processMessage(workCtx, msgBytes)
							}
						}
					}
					// Small sleep to prevent busy waiting
					time.Sleep(1 * time.Millisecond)
				}
			}
		}()
		return nil
	}

	// Original implementation for production use
	if w.wg != nil {
		go func() {
			defer w.wg.Done()
			for {
				select {
				case msg := <-w.jobQueue:
					w.processMessageSarama(msg)
				case <-w.quit:
					log.Printf("Worker %d stopping", w.id)
					w.running = false
					return
				case <-workCtx.Done():
					log.Printf("Worker %d stopping due to context cancellation", w.id)
					w.running = false
					return
				}
			}
		}()
	}
	return nil
}

// Stop stops the worker
func (w *Worker) Stop() {
	w.running = false
	if w.stopChan != nil {
		close(w.stopChan)
	}
	if w.consumer != nil {
		w.consumer.Close()
	}
	if w.quit != nil {
		go func() {
			w.quit <- true
		}()
	}
}

// IsHealthy returns whether the worker is running
func (w *Worker) IsHealthy() bool {
	return w.running
}

// processMessage processes a message with context (for tests)
func (w *Worker) processMessage(ctx context.Context, message []byte) error {
	if w.processor != nil {
		return w.processor.ProcessMessage(ctx, message)
	}
	// Fallback to basic processing
	log.Printf("Worker %d processing message: %s", w.id, string(message))
	return nil
}

// processMessageSarama processes a Sarama message (original implementation)
func (w *Worker) processMessageSarama(msg *sarama.ConsumerMessage) {
	log.Printf("Worker %d processing message from topic %s, partition %d, offset %d",
		w.id, msg.Topic, msg.Partition, msg.Offset)

	switch msg.Topic {
	case "coffee_orders":
		w.processOrder(msg)
	case "processed_orders":
		w.processProcessedOrder(msg)
	default:
		log.Printf("Unknown topic: %s", msg.Topic)
	}
}

// processOrder processes an order message
func (w *Worker) processOrder(msg *sarama.ConsumerMessage) {
	var order Order
	if err := json.Unmarshal(msg.Value, &order); err != nil {
		log.Printf("Error parsing order: %v", err)
		return
	}

	log.Printf("Worker %d received order for %s coffee from %s",
		w.id, order.CoffeeType, order.CustomerName)
}

// processProcessedOrder processes a processed order message
func (w *Worker) processProcessedOrder(msg *sarama.ConsumerMessage) {
	var processedOrder ProcessedOrder
	if err := json.Unmarshal(msg.Value, &processedOrder); err != nil {
		log.Printf("Error parsing processed order: %v", err)
		return
	}

	log.Printf("Worker %d brewing %s coffee for %s (Order ID: %s, Preparation Time: %d seconds)",
		w.id, processedOrder.CoffeeType, processedOrder.CustomerName, processedOrder.OrderID, processedOrder.PreparationTime)

	// Simulate coffee preparation
	log.Printf("Worker %d starting preparation of %s coffee for %s...",
		w.id, processedOrder.CoffeeType, processedOrder.CustomerName)
	time.Sleep(time.Duration(processedOrder.PreparationTime) * time.Millisecond) // Use milliseconds instead of seconds for demo
	log.Printf("Worker %d finished preparing %s coffee for %s!",
		w.id, processedOrder.CoffeeType, processedOrder.CustomerName)
}

// WorkerPool represents a pool of workers with dynamic scaling
type WorkerPool struct {
	jobQueue         chan *sarama.ConsumerMessage
	workers          []*Worker
	minWorkers       int
	maxWorkers       int
	currentWorkers   int
	wg               sync.WaitGroup
	activeWorkers    int
	mu               sync.RWMutex
	ctx              context.Context
	cancel           context.CancelFunc
	scalingTicker    *time.Ticker
	metrics          *WorkerPoolMetrics
	backpressureMode bool
}

// WorkerPoolMetrics tracks pool performance metrics
type WorkerPoolMetrics struct {
	mu                sync.RWMutex
	processedMessages int64
	failedMessages    int64
	queueSize         int64
	avgProcessingTime time.Duration
	lastScaleTime     time.Time
}

// NewWorkerPool creates a new worker pool with dynamic scaling
func NewWorkerPool(workerPool int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	
	pool := &WorkerPool{
		jobQueue:       make(chan *sarama.ConsumerMessage, 1000), // Increased buffer size
		minWorkers:     workerPool,
		maxWorkers:     workerPool * 3, // Allow scaling up to 3x
		currentWorkers: 0,
		ctx:            ctx,
		cancel:         cancel,
		scalingTicker:  time.NewTicker(30 * time.Second), // Check scaling every 30s
		metrics: &WorkerPoolMetrics{
			lastScaleTime: time.Now(),
		},
	}
	
	// Start scaling monitor
	go pool.monitorAndScale()
	
	return pool
}

// Start starts the worker pool
func (wp *WorkerPool) Start() {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	
	// Start with minimum number of workers
	wp.workers = make([]*Worker, wp.maxWorkers) // Pre-allocate for max capacity
	wp.startWorkers(wp.minWorkers)
	wp.currentWorkers = wp.minWorkers
}

// startWorkers starts the specified number of workers
func (wp *WorkerPool) startWorkers(count int) {
	for i := wp.currentWorkers; i < wp.currentWorkers+count && i < wp.maxWorkers; i++ {
		wp.wg.Add(1)
		wp.workers[i] = NewWorker(i, wp.jobQueue, &wp.wg, wp.maxWorkers)
		wp.workers[i].Start(wp.ctx)
		log.Printf("Worker %d started", i)
	}
}

// stopWorkers stops the specified number of workers
func (wp *WorkerPool) stopWorkers(count int) {
	for i := wp.currentWorkers - 1; i >= wp.currentWorkers-count && i >= wp.minWorkers; i-- {
		if wp.workers[i] != nil {
			wp.workers[i].Stop()
			wp.workers[i] = nil
			log.Printf("Worker %d stopped", i)
		}
	}
}

// monitorAndScale monitors queue size and scales workers accordingly
func (wp *WorkerPool) monitorAndScale() {
	defer wp.scalingTicker.Stop()
	
	for {
		select {
		case <-wp.scalingTicker.C:
			wp.checkAndScale()
		case <-wp.ctx.Done():
			return
		}
	}
}

// checkAndScale decides whether to scale up or down
func (wp *WorkerPool) checkAndScale() {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	
	queueSize := len(wp.jobQueue)
	queueCapacity := cap(wp.jobQueue)
	
	// Update metrics
	wp.metrics.mu.Lock()
	wp.metrics.queueSize = int64(queueSize)
	wp.metrics.mu.Unlock()
	
	// Scale up if queue is > 80% full and we can add more workers
	if queueSize > int(float64(queueCapacity)*0.8) && wp.currentWorkers < wp.maxWorkers {
		scaleUp := min(2, wp.maxWorkers-wp.currentWorkers) // Scale up by 2 or remaining capacity
		wp.startWorkers(scaleUp)
		wp.currentWorkers += scaleUp
		log.Printf("Scaled up by %d workers due to high queue pressure (%d/%d)", scaleUp, queueSize, queueCapacity)
		wp.metrics.lastScaleTime = time.Now()
	}
	
	// Scale down if queue is < 20% full and we have more than minimum workers
	if queueSize < int(float64(queueCapacity)*0.2) && wp.currentWorkers > wp.minWorkers {
		// Only scale down if it's been at least 2 minutes since last scaling
		if time.Since(wp.metrics.lastScaleTime) > 2*time.Minute {
			scaleDown := min(1, wp.currentWorkers-wp.minWorkers) // Scale down conservatively
			wp.stopWorkers(scaleDown)
			wp.currentWorkers -= scaleDown
			log.Printf("Scaled down by %d workers due to low queue pressure (%d/%d)", scaleDown, queueSize, queueCapacity)
			wp.metrics.lastScaleTime = time.Now()
		}
	}
	
	// Enable backpressure mode if queue is > 95% full
	wp.backpressureMode = queueSize > int(float64(queueCapacity)*0.95)
	if wp.backpressureMode {
		log.Printf("Backpressure mode enabled - queue nearly full (%d/%d)", queueSize, queueCapacity)
	}
}

// Stop stops the worker pool
func (wp *WorkerPool) Stop() {
	wp.cancel()
	
	wp.mu.Lock()
	defer wp.mu.Unlock()
	
	// Stop all workers
	for i := 0; i < wp.currentWorkers; i++ {
		if wp.workers[i] != nil {
			wp.workers[i].Stop()
		}
	}
	
	wp.wg.Wait()
	close(wp.jobQueue)
}

// Submit submits a message to the worker pool with backpressure handling
func (wp *WorkerPool) Submit(msg *sarama.ConsumerMessage) {
	select {
	case wp.jobQueue <- msg:
		// Successfully queued
		wp.metrics.mu.Lock()
		wp.metrics.processedMessages++
		wp.metrics.mu.Unlock()
	default:
		// Queue is full - apply backpressure
		if wp.backpressureMode {
			log.Printf("Dropping message due to backpressure - queue full")
			wp.metrics.mu.Lock()
			wp.metrics.failedMessages++
			wp.metrics.mu.Unlock()
			return
		}
		
		// Try with timeout
		timeout := time.After(100 * time.Millisecond)
		select {
		case wp.jobQueue <- msg:
			wp.metrics.mu.Lock()
			wp.metrics.processedMessages++
			wp.metrics.mu.Unlock()
		case <-timeout:
			log.Printf("Message submission timed out - queue congested")
			wp.metrics.mu.Lock()
			wp.metrics.failedMessages++
			wp.metrics.mu.Unlock()
		}
	}
}

// QueueSize returns the current size of the job queue
func (wp *WorkerPool) QueueSize() int {
	return len(wp.jobQueue)
}

// QueueCapacity returns the capacity of the job queue
func (wp *WorkerPool) QueueCapacity() int {
	return cap(wp.jobQueue)
}

// ActiveWorkers returns the current number of active workers
func (wp *WorkerPool) ActiveWorkers() int {
	wp.mu.RLock()
	defer wp.mu.RUnlock()
	return wp.currentWorkers
}

// GetMetrics returns current pool metrics
func (wp *WorkerPool) GetMetrics() WorkerPoolMetrics {
	wp.metrics.mu.RLock()
	defer wp.metrics.mu.RUnlock()
	
	return WorkerPoolMetrics{
		processedMessages: wp.metrics.processedMessages,
		failedMessages:    wp.metrics.failedMessages,
		queueSize:         wp.metrics.queueSize,
		avgProcessingTime: wp.metrics.avgProcessingTime,
		lastScaleTime:     wp.metrics.lastScaleTime,
	}
}

// Helper function for minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
