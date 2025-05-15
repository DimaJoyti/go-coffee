package worker

import (
	"encoding/json"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/IBM/sarama"

	"kafka_worker/metrics"
)

// Worker represents a worker that processes messages
type Worker struct {
	id         int
	jobQueue   <-chan *sarama.ConsumerMessage
	quit       chan bool
	wg         *sync.WaitGroup
	workerPool int
	active     int32 // Atomic counter for active state
}

// NewWorker creates a new worker
func NewWorker(id int, jobQueue <-chan *sarama.ConsumerMessage, wg *sync.WaitGroup, workerPool int) *Worker {
	return &Worker{
		id:         id,
		jobQueue:   jobQueue,
		quit:       make(chan bool),
		wg:         wg,
		workerPool: workerPool,
		active:     0,
	}
}

// Start starts the worker
func (w *Worker) Start() {
	go func() {
		defer w.wg.Done()
		for {
			select {
			case msg := <-w.jobQueue:
				atomic.StoreInt32(&w.active, 1)
				w.processMessage(msg)
				atomic.StoreInt32(&w.active, 0)
			case <-w.quit:
				log.Printf("Worker %d stopping", w.id)
				return
			}
		}
	}()
}

// Stop stops the worker
func (w *Worker) Stop() {
	go func() {
		w.quit <- true
	}()
}

// IsActive returns whether the worker is active
func (w *Worker) IsActive() bool {
	return atomic.LoadInt32(&w.active) == 1
}

// processMessage processes a message
func (w *Worker) processMessage(msg *sarama.ConsumerMessage) {
	startTime := time.Now()
	log.Printf("Worker %d processing message from topic %s, partition %d, offset %d",
		w.id, msg.Topic, msg.Partition, msg.Offset)

	metrics.KafkaMessagesProcessedTotal.WithLabelValues(msg.Topic).Inc()

	switch msg.Topic {
	case "coffee_orders":
		w.processOrder(msg)
	case "processed_orders":
		w.processProcessedOrder(msg)
	default:
		log.Printf("Unknown topic: %s", msg.Topic)
		metrics.KafkaMessagesFailedTotal.WithLabelValues(msg.Topic).Inc()
	}

	metrics.WorkerProcessingTime.WithLabelValues(
		string(w.id), msg.Topic).Observe(time.Since(startTime).Seconds())
}

// processOrder processes an order message
func (w *Worker) processOrder(msg *sarama.ConsumerMessage) {
	var order Order
	if err := json.Unmarshal(msg.Value, &order); err != nil {
		log.Printf("Error parsing order: %v", err)
		metrics.KafkaMessagesFailedTotal.WithLabelValues(msg.Topic).Inc()
		metrics.OrdersProcessedFailedTotal.Inc()
		return
	}

	log.Printf("Worker %d received order for %s coffee from %s",
		w.id, order.CoffeeType, order.CustomerName)

	metrics.OrdersProcessedTotal.Inc()
	metrics.OrdersProcessedSuccessTotal.Inc()
}

// processProcessedOrder processes a processed order message
func (w *Worker) processProcessedOrder(msg *sarama.ConsumerMessage) {
	startTime := time.Now()
	var processedOrder ProcessedOrder
	if err := json.Unmarshal(msg.Value, &processedOrder); err != nil {
		log.Printf("Error parsing processed order: %v", err)
		metrics.KafkaMessagesFailedTotal.WithLabelValues(msg.Topic).Inc()
		metrics.OrdersProcessedFailedTotal.Inc()
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

	metrics.OrdersProcessedTotal.Inc()
	metrics.OrdersProcessedSuccessTotal.Inc()
	metrics.OrderPreparationTime.Observe(time.Since(startTime).Seconds())
}

// WorkerPool represents a pool of workers
type WorkerPool struct {
	jobQueue   chan *sarama.ConsumerMessage
	workers    []*Worker
	workerPool int
	wg         sync.WaitGroup
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
	metrics.WorkerPoolQueueSize.Set(float64(len(wp.jobQueue)))
	wp.jobQueue <- msg
}

// QueueSize returns the current size of the job queue
func (wp *WorkerPool) QueueSize() int {
	return len(wp.jobQueue)
}

// ActiveWorkers returns the number of active workers
func (wp *WorkerPool) ActiveWorkers() int {
	activeCount := 0
	for _, worker := range wp.workers {
		if worker.IsActive() {
			activeCount++
		}
	}
	return activeCount
}
