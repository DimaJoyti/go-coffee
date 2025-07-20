package websocket

import (
	"context"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/hft/performance/memory"
	"github.com/gorilla/websocket"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// MessageType represents the type of WebSocket message
type MessageType int

const (
	MessageTypeText MessageType = iota
	MessageTypeBinary
	MessageTypePing
	MessageTypePong
	MessageTypeClose
)

// OptimizedMessage represents a zero-copy WebSocket message
type OptimizedMessage struct {
	Type      MessageType
	Data      []byte
	Timestamp int64
	Latency   time.Duration
}

// MessageHandler handles incoming WebSocket messages
type MessageHandler interface {
	HandleMessage(ctx context.Context, msg *OptimizedMessage) error
}

// ConnectionStats tracks WebSocket connection statistics
type ConnectionStats struct {
	MessagesReceived  uint64
	MessagesSent      uint64
	BytesReceived     uint64
	BytesSent         uint64
	Reconnections     uint64
	Errors            uint64
	AvgLatency        int64 // nanoseconds
	LastMessageTime   int64
	ConnectionTime    int64
	IsConnected       int32 // atomic bool
}

// OptimizedWebSocketClient provides ultra-low latency WebSocket connectivity
type OptimizedWebSocketClient struct {
	url             string
	conn            *websocket.Conn
	connMutex       sync.RWMutex
	
	// Message handling
	messageHandler  MessageHandler
	messagePool     *memory.ObjectPool[OptimizedMessage]
	bufferPool      *memory.ByteBufferPool
	
	// Performance optimization
	readBuffer      []byte
	writeBuffer     []byte
	readBufferSize  int
	writeBufferSize int
	
	// Connection management
	ctx             context.Context
	cancel          context.CancelFunc
	reconnectDelay  time.Duration
	maxReconnects   int
	
	// Statistics
	stats           *ConnectionStats
	
	// Observability
	tracer          trace.Tracer
	
	// Configuration
	config          *WebSocketConfig
	
	// Lock-free message queue
	messageQueue    unsafe.Pointer // *LockFreeQueue[*OptimizedMessage]
	
	// Worker goroutines
	readerDone      chan struct{}
	writerDone      chan struct{}
	processorDone   chan struct{}
}

// WebSocketConfig holds configuration for optimized WebSocket client
type WebSocketConfig struct {
	URL                 string
	ReadBufferSize      int
	WriteBufferSize     int
	HandshakeTimeout    time.Duration
	ReadTimeout         time.Duration
	WriteTimeout        time.Duration
	PingInterval        time.Duration
	PongTimeout         time.Duration
	ReconnectDelay      time.Duration
	MaxReconnects       int
	EnableCompression   bool
	EnableBinaryFrames  bool
	MessageQueueSize    int
	WorkerCount         int
	CPUAffinity         []int
	SocketBufferSize    int
	NoDelay             bool
	KeepAlive           bool
	KeepAlivePeriod     time.Duration
}

// DefaultWebSocketConfig returns default configuration
func DefaultWebSocketConfig() *WebSocketConfig {
	return &WebSocketConfig{
		ReadBufferSize:      65536,  // 64KB
		WriteBufferSize:     65536,  // 64KB
		HandshakeTimeout:    10 * time.Second,
		ReadTimeout:         0, // No timeout for market data
		WriteTimeout:        5 * time.Second,
		PingInterval:        30 * time.Second,
		PongTimeout:        10 * time.Second,
		ReconnectDelay:     1 * time.Second,
		MaxReconnects:      10,
		EnableCompression:  false, // Disable for latency
		EnableBinaryFrames: true,
		MessageQueueSize:   10000,
		WorkerCount:        2,
		SocketBufferSize:   1048576, // 1MB
		NoDelay:            true,
		KeepAlive:          true,
		KeepAlivePeriod:    30 * time.Second,
	}
}

// NewOptimizedWebSocketClient creates a new optimized WebSocket client
func NewOptimizedWebSocketClient(config *WebSocketConfig, handler MessageHandler) *OptimizedWebSocketClient {
	ctx, cancel := context.WithCancel(context.Background())
	
	client := &OptimizedWebSocketClient{
		url:             config.URL,
		messageHandler:  handler,
		ctx:             ctx,
		cancel:          cancel,
		reconnectDelay:  config.ReconnectDelay,
		maxReconnects:   config.MaxReconnects,
		config:          config,
		stats:           &ConnectionStats{},
		tracer:          otel.Tracer("hft.websocket.client"),
		readBuffer:      make([]byte, config.ReadBufferSize),
		writeBuffer:     make([]byte, config.WriteBufferSize),
		readBufferSize:  config.ReadBufferSize,
		writeBufferSize: config.WriteBufferSize,
		readerDone:      make(chan struct{}),
		writerDone:      make(chan struct{}),
		processorDone:   make(chan struct{}),
	}
	
	// Initialize message pool
	client.messagePool = memory.NewObjectPool[OptimizedMessage](
		"WebSocketMessagePool",
		1000,
		func() *OptimizedMessage {
			return &OptimizedMessage{}
		},
		func(msg *OptimizedMessage) {
			msg.Type = MessageTypeText
			msg.Data = nil
			msg.Timestamp = 0
			msg.Latency = 0
		},
	)
	
	// Initialize buffer pool
	client.bufferPool = memory.NewByteBufferPool(500, 200, 100)
	
	// Initialize lock-free message queue
	client.initMessageQueue()
	
	return client
}

// Connect establishes WebSocket connection with optimizations
func (c *OptimizedWebSocketClient) Connect() error {
	ctx, span := c.tracer.Start(c.ctx, "OptimizedWebSocketClient.Connect")
	defer span.End()
	
	span.SetAttributes(attribute.String("url", c.url))
	
	// Configure dialer with optimizations
	dialer := &websocket.Dialer{
		HandshakeTimeout: c.config.HandshakeTimeout,
		ReadBufferSize:   c.config.ReadBufferSize,
		WriteBufferSize:  c.config.WriteBufferSize,
		EnableCompression: c.config.EnableCompression,
		NetDialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			d := &net.Dialer{
				Timeout:   c.config.HandshakeTimeout,
				KeepAlive: c.config.KeepAlivePeriod,
			}
			
			conn, err := d.DialContext(ctx, network, addr)
			if err != nil {
				return nil, err
			}
			
			// Apply socket optimizations
			if tcpConn, ok := conn.(*net.TCPConn); ok {
				if err := c.optimizeTCPConnection(tcpConn); err != nil {
					span.RecordError(err)
					// Log warning but don't fail connection
				}
			}
			
			return conn, nil
		},
	}
	
	// Establish connection
	conn, _, err := dialer.DialContext(ctx, c.url, nil)
	if err != nil {
		span.RecordError(err)
		atomic.AddUint64(&c.stats.Errors, 1)
		return fmt.Errorf("failed to connect to %s: %w", c.url, err)
	}
	
	c.connMutex.Lock()
	c.conn = conn
	c.connMutex.Unlock()
	
	// Configure connection timeouts
	if c.config.ReadTimeout > 0 {
		conn.SetReadDeadline(time.Now().Add(c.config.ReadTimeout))
	}
	if c.config.WriteTimeout > 0 {
		conn.SetWriteDeadline(time.Now().Add(c.config.WriteTimeout))
	}
	
	// Set connection statistics
	atomic.StoreInt32(&c.stats.IsConnected, 1)
	atomic.StoreInt64(&c.stats.ConnectionTime, time.Now().UnixNano())
	
	// Start worker goroutines
	go c.readerWorker()
	go c.writerWorker()
	go c.messageProcessor()
	
	// Start ping/pong handler
	if c.config.PingInterval > 0 {
		go c.pingHandler()
	}
	
	span.SetAttributes(attribute.Bool("connected", true))
	return nil
}

// optimizeTCPConnection applies TCP-level optimizations
func (c *OptimizedWebSocketClient) optimizeTCPConnection(conn *net.TCPConn) error {
	// Disable Nagle's algorithm for low latency
	if c.config.NoDelay {
		if err := conn.SetNoDelay(true); err != nil {
			return fmt.Errorf("failed to set TCP_NODELAY: %w", err)
		}
	}
	
	// Enable keep-alive
	if c.config.KeepAlive {
		if err := conn.SetKeepAlive(true); err != nil {
			return fmt.Errorf("failed to enable keep-alive: %w", err)
		}
		
		if err := conn.SetKeepAlivePeriod(c.config.KeepAlivePeriod); err != nil {
			return fmt.Errorf("failed to set keep-alive period: %w", err)
		}
	}
	
	// Set socket buffer sizes (requires syscalls on Linux)
	// This would typically use syscall.SetsockoptInt with SO_RCVBUF/SO_SNDBUF
	// For now, we'll skip this implementation
	
	return nil
}

// readerWorker handles incoming WebSocket messages
func (c *OptimizedWebSocketClient) readerWorker() {
	defer close(c.readerDone)
	
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}
		
		c.connMutex.RLock()
		conn := c.conn
		c.connMutex.RUnlock()
		
		if conn == nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		
		// Read message with zero-copy optimization
		messageType, data, err := conn.ReadMessage()
		receiveTime := time.Now().UnixNano()
		
		if err != nil {
			atomic.AddUint64(&c.stats.Errors, 1)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				// Attempt reconnection
				go c.reconnect()
			}
			continue
		}
		
		// Update statistics
		atomic.AddUint64(&c.stats.MessagesReceived, 1)
		atomic.AddUint64(&c.stats.BytesReceived, uint64(len(data)))
		atomic.StoreInt64(&c.stats.LastMessageTime, receiveTime)
		
		// Create optimized message
		msg := c.messagePool.Get()
		msg.Type = MessageType(messageType)
		msg.Data = data
		msg.Timestamp = receiveTime
		
		// Enqueue message for processing
		c.enqueueMessage(msg)
	}
}

// writerWorker handles outgoing WebSocket messages
func (c *OptimizedWebSocketClient) writerWorker() {
	defer close(c.writerDone)
	
	ticker := time.NewTicker(c.config.PingInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			// Send ping
			c.sendPing()
		}
	}
}

// messageProcessor processes incoming messages
func (c *OptimizedWebSocketClient) messageProcessor() {
	defer close(c.processorDone)
	
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}
		
		msg := c.dequeueMessage()
		if msg == nil {
			time.Sleep(10 * time.Microsecond) // Very short sleep
			continue
		}
		
		// Calculate processing latency
		processTime := time.Now().UnixNano()
		msg.Latency = time.Duration(processTime - msg.Timestamp)
		
		// Update average latency
		c.updateAverageLatency(msg.Latency)
		
		// Handle message
		if c.messageHandler != nil {
			if err := c.messageHandler.HandleMessage(c.ctx, msg); err != nil {
				atomic.AddUint64(&c.stats.Errors, 1)
			}
		}
		
		// Return message to pool
		c.messagePool.Put(msg)
	}
}

// sendPing sends a ping message
func (c *OptimizedWebSocketClient) sendPing() {
	c.connMutex.RLock()
	conn := c.conn
	c.connMutex.RUnlock()
	
	if conn == nil {
		return
	}
	
	if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
		atomic.AddUint64(&c.stats.Errors, 1)
	}
}

// pingHandler handles ping/pong messages
func (c *OptimizedWebSocketClient) pingHandler() {
	c.connMutex.RLock()
	conn := c.conn
	c.connMutex.RUnlock()
	
	if conn == nil {
		return
	}
	
	conn.SetPongHandler(func(appData string) error {
		// Reset read deadline on pong
		if c.config.ReadTimeout > 0 {
			return conn.SetReadDeadline(time.Now().Add(c.config.ReadTimeout))
		}
		return nil
	})
}

// reconnect attempts to reconnect the WebSocket
func (c *OptimizedWebSocketClient) reconnect() {
	atomic.StoreInt32(&c.stats.IsConnected, 0)
	atomic.AddUint64(&c.stats.Reconnections, 1)
	
	for attempt := 0; attempt < c.maxReconnects; attempt++ {
		select {
		case <-c.ctx.Done():
			return
		default:
		}
		
		time.Sleep(c.reconnectDelay)
		
		if err := c.Connect(); err == nil {
			return // Successfully reconnected
		}
		
		// Exponential backoff
		c.reconnectDelay *= 2
		if c.reconnectDelay > 30*time.Second {
			c.reconnectDelay = 30 * time.Second
		}
	}
}

// SendMessage sends a message through the WebSocket
func (c *OptimizedWebSocketClient) SendMessage(messageType MessageType, data []byte) error {
	c.connMutex.RLock()
	conn := c.conn
	c.connMutex.RUnlock()
	
	if conn == nil {
		return fmt.Errorf("connection not established")
	}
	
	wsMessageType := websocket.TextMessage
	if messageType == MessageTypeBinary {
		wsMessageType = websocket.BinaryMessage
	}
	
	if err := conn.WriteMessage(wsMessageType, data); err != nil {
		atomic.AddUint64(&c.stats.Errors, 1)
		return err
	}
	
	atomic.AddUint64(&c.stats.MessagesSent, 1)
	atomic.AddUint64(&c.stats.BytesSent, uint64(len(data)))
	
	return nil
}

// Close closes the WebSocket connection
func (c *OptimizedWebSocketClient) Close() error {
	c.cancel()
	
	c.connMutex.Lock()
	conn := c.conn
	c.conn = nil
	c.connMutex.Unlock()
	
	if conn != nil {
		// Send close message
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		conn.Close()
	}
	
	// Wait for workers to finish
	<-c.readerDone
	<-c.writerDone
	<-c.processorDone
	
	atomic.StoreInt32(&c.stats.IsConnected, 0)
	return nil
}

// GetStats returns connection statistics
func (c *OptimizedWebSocketClient) GetStats() ConnectionStats {
	return ConnectionStats{
		MessagesReceived: atomic.LoadUint64(&c.stats.MessagesReceived),
		MessagesSent:     atomic.LoadUint64(&c.stats.MessagesSent),
		BytesReceived:    atomic.LoadUint64(&c.stats.BytesReceived),
		BytesSent:        atomic.LoadUint64(&c.stats.BytesSent),
		Reconnections:    atomic.LoadUint64(&c.stats.Reconnections),
		Errors:           atomic.LoadUint64(&c.stats.Errors),
		AvgLatency:       atomic.LoadInt64(&c.stats.AvgLatency),
		LastMessageTime:  atomic.LoadInt64(&c.stats.LastMessageTime),
		ConnectionTime:   atomic.LoadInt64(&c.stats.ConnectionTime),
		IsConnected:      atomic.LoadInt32(&c.stats.IsConnected),
	}
}

// IsConnected returns true if the WebSocket is connected
func (c *OptimizedWebSocketClient) IsConnected() bool {
	return atomic.LoadInt32(&c.stats.IsConnected) == 1
}

// updateAverageLatency updates the running average latency
func (c *OptimizedWebSocketClient) updateAverageLatency(latency time.Duration) {
	for {
		current := atomic.LoadInt64(&c.stats.AvgLatency)
		// Simple exponential moving average
		newAvg := (current*9 + int64(latency)) / 10
		if atomic.CompareAndSwapInt64(&c.stats.AvgLatency, current, newAvg) {
			break
		}
	}
}

// Lock-free message queue operations (simplified implementation)
func (c *OptimizedWebSocketClient) initMessageQueue() {
	// This would initialize a lock-free queue
	// For now, we'll use a simple channel-based implementation
	queue := make(chan *OptimizedMessage, c.config.MessageQueueSize)
	atomic.StorePointer(&c.messageQueue, unsafe.Pointer(&queue))
}

func (c *OptimizedWebSocketClient) enqueueMessage(msg *OptimizedMessage) {
	queuePtr := (*chan *OptimizedMessage)(atomic.LoadPointer(&c.messageQueue))
	if queuePtr != nil {
		select {
		case *queuePtr <- msg:
		default:
			// Queue full, drop message
			c.messagePool.Put(msg)
			atomic.AddUint64(&c.stats.Errors, 1)
		}
	}
}

func (c *OptimizedWebSocketClient) dequeueMessage() *OptimizedMessage {
	queuePtr := (*chan *OptimizedMessage)(atomic.LoadPointer(&c.messageQueue))
	if queuePtr != nil {
		select {
		case msg := <-*queuePtr:
			return msg
		default:
			return nil
		}
	}
	return nil
}
