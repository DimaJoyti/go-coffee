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

// WorkerPool represents a pool of workers
type WorkerPool struct {
	jobQueue      chan *sarama.ConsumerMessage
	workers       []*Worker
	workerPool    int
	wg            sync.WaitGroup
	activeWorkers int
	mu            sync.RWMutex
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(workerPool int) *WorkerPool {
	return &WorkerPool{
		jobQueue:   make(chan *sarama.ConsumerMessage, 100),
		workerPool: workerPool,
	}
}

// Start starts the worker pool
func (wp *WorkerPool) Start() {
	// Create and start workers
	wp.workers = make([]*Worker, wp.workerPool)
	for i := 0; i < wp.workerPool; i++ {
		wp.wg.Add(1)
		wp.workers[i] = NewWorker(i, wp.jobQueue, &wp.wg, wp.workerPool)
		wp.workers[i].Start()
		log.Printf("Worker %d started", i)
	}
}

// Stop stops the worker pool
func (wp *WorkerPool) Stop() {
	// Stop workers
	for i := 0; i < wp.workerPool; i++ {
		wp.workers[i].Stop()
	}
	wp.wg.Wait()
	close(wp.jobQueue)
}

// Submit submits a message to the worker pool
func (wp *WorkerPool) Submit(msg *sarama.ConsumerMessage) {
	wp.jobQueue <- msg
}

// QueueSize returns the current size of the job queue
func (wp *WorkerPool) QueueSize() int {
	return len(wp.jobQueue)
}

// ActiveWorkers returns the current number of active workers
func (wp *WorkerPool) ActiveWorkers() int {
	wp.mu.RLock()
	defer wp.mu.RUnlock()
	return wp.activeWorkers
}
