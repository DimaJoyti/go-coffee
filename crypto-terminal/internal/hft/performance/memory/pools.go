package memory

import (
	"runtime"
	"sync"
	"sync/atomic"
	"unsafe"
)

// PoolStats tracks memory pool statistics
type PoolStats struct {
	Allocations   uint64
	Deallocations uint64
	Hits          uint64
	Misses        uint64
	CurrentSize   int32
	MaxSize       int32
	TotalCreated  uint64
}

// ObjectPool is a generic lock-free object pool
type ObjectPool[T any] struct {
	pool     chan *T
	factory  func() *T
	reset    func(*T)
	stats    *PoolStats
	maxSize  int32
	name     string
}

// NewObjectPool creates a new object pool
func NewObjectPool[T any](name string, maxSize int, factory func() *T, reset func(*T)) *ObjectPool[T] {
	return &ObjectPool[T]{
		pool:    make(chan *T, maxSize),
		factory: factory,
		reset:   reset,
		stats: &PoolStats{
			MaxSize: int32(maxSize),
		},
		maxSize: int32(maxSize),
		name:    name,
	}
}

// Get retrieves an object from the pool
func (p *ObjectPool[T]) Get() *T {
	select {
	case obj := <-p.pool:
		atomic.AddUint64(&p.stats.Hits, 1)
		atomic.AddInt32(&p.stats.CurrentSize, -1)
		return obj
	default:
		atomic.AddUint64(&p.stats.Misses, 1)
		atomic.AddUint64(&p.stats.TotalCreated, 1)
		return p.factory()
	}
}

// Put returns an object to the pool
func (p *ObjectPool[T]) Put(obj *T) {
	if obj == nil {
		return
	}
	
	// Reset the object to clean state
	if p.reset != nil {
		p.reset(obj)
	}
	
	select {
	case p.pool <- obj:
		atomic.AddInt32(&p.stats.CurrentSize, 1)
		atomic.AddUint64(&p.stats.Deallocations, 1)
	default:
		// Pool is full, let GC handle it
		atomic.AddUint64(&p.stats.Allocations, 1)
	}
}

// GetStats returns pool statistics
func (p *ObjectPool[T]) GetStats() PoolStats {
	return PoolStats{
		Allocations:   atomic.LoadUint64(&p.stats.Allocations),
		Deallocations: atomic.LoadUint64(&p.stats.Deallocations),
		Hits:          atomic.LoadUint64(&p.stats.Hits),
		Misses:        atomic.LoadUint64(&p.stats.Misses),
		CurrentSize:   atomic.LoadInt32(&p.stats.CurrentSize),
		MaxSize:       p.stats.MaxSize,
		TotalCreated:  atomic.LoadUint64(&p.stats.TotalCreated),
	}
}

// Generic order type for pooling (avoiding external dependencies)
type GenericOrder struct {
	ID       string
	Symbol   string
	Side     string
	Type     string
	Quantity float64
	Price    float64
	Status   string
}

// OrderPool manages order object pooling
type OrderPool struct {
	pool *ObjectPool[GenericOrder]
}

// NewOrderPool creates a new order pool
func NewOrderPool(maxSize int) *OrderPool {
	return &OrderPool{
		pool: NewObjectPool[GenericOrder](
			"OrderPool",
			maxSize,
			func() *GenericOrder {
				return &GenericOrder{}
			},
			func(order *GenericOrder) {
				resetGenericOrder(order)
			},
		),
	}
}

// Get retrieves an order from the pool
func (op *OrderPool) Get() *GenericOrder {
	return op.pool.Get()
}

// Put returns an order to the pool
func (op *OrderPool) Put(order *GenericOrder) {
	op.pool.Put(order)
}

// GetStats returns pool statistics
func (op *OrderPool) GetStats() PoolStats {
	return op.pool.GetStats()
}

// MarketDataPool manages market data object pooling
type MarketDataPool struct {
	tickPool      *ObjectPool[MarketDataTick]
	orderBookPool *ObjectPool[OrderBookSnapshot]
}

// MarketDataTick represents a market data tick for pooling
type MarketDataTick struct {
	Symbol      string
	Exchange    string
	Price       float64
	Quantity    float64
	Side        string
	BidPrice    float64
	BidQuantity float64
	AskPrice    float64
	AskQuantity float64
	Timestamp   int64
	ReceiveTime int64
	ProcessTime int64
	Latency     int64
	SequenceNum uint64
}

// OrderBookSnapshot represents an order book snapshot for pooling
type OrderBookSnapshot struct {
	Symbol      string
	Exchange    string
	Bids        []PriceLevel
	Asks        []PriceLevel
	Timestamp   int64
	ReceiveTime int64
	SequenceNum uint64
}

// PriceLevel represents a price level in the order book
type PriceLevel struct {
	Price    float64
	Quantity float64
	Count    int32
}

// NewMarketDataPool creates a new market data pool
func NewMarketDataPool(maxTickSize, maxOrderBookSize int) *MarketDataPool {
	return &MarketDataPool{
		tickPool: NewObjectPool[MarketDataTick](
			"MarketDataTickPool",
			maxTickSize,
			func() *MarketDataTick {
				return &MarketDataTick{}
			},
			func(tick *MarketDataTick) {
				resetMarketDataTick(tick)
			},
		),
		orderBookPool: NewObjectPool[OrderBookSnapshot](
			"OrderBookPool",
			maxOrderBookSize,
			func() *OrderBookSnapshot {
				return &OrderBookSnapshot{
					Bids: make([]PriceLevel, 0, 20),
					Asks: make([]PriceLevel, 0, 20),
				}
			},
			func(ob *OrderBookSnapshot) {
				resetOrderBookSnapshot(ob)
			},
		),
	}
}

// GetTick retrieves a market data tick from the pool
func (mdp *MarketDataPool) GetTick() *MarketDataTick {
	return mdp.tickPool.Get()
}

// PutTick returns a market data tick to the pool
func (mdp *MarketDataPool) PutTick(tick *MarketDataTick) {
	mdp.tickPool.Put(tick)
}

// GetOrderBook retrieves an order book snapshot from the pool
func (mdp *MarketDataPool) GetOrderBook() *OrderBookSnapshot {
	return mdp.orderBookPool.Get()
}

// PutOrderBook returns an order book snapshot to the pool
func (mdp *MarketDataPool) PutOrderBook(ob *OrderBookSnapshot) {
	mdp.orderBookPool.Put(ob)
}

// GetTickStats returns tick pool statistics
func (mdp *MarketDataPool) GetTickStats() PoolStats {
	return mdp.tickPool.GetStats()
}

// GetOrderBookStats returns order book pool statistics
func (mdp *MarketDataPool) GetOrderBookStats() PoolStats {
	return mdp.orderBookPool.GetStats()
}

// ByteBufferPool manages byte buffer pooling for network operations
type ByteBufferPool struct {
	smallPool  *ObjectPool[[]byte] // 1KB buffers
	mediumPool *ObjectPool[[]byte] // 8KB buffers
	largePool  *ObjectPool[[]byte] // 64KB buffers
}

// NewByteBufferPool creates a new byte buffer pool
func NewByteBufferPool(smallSize, mediumSize, largeSize int) *ByteBufferPool {
	return &ByteBufferPool{
		smallPool: NewObjectPool[[]byte](
			"SmallByteBufferPool",
			smallSize,
			func() *[]byte {
				buf := make([]byte, 1024) // 1KB
				return &buf
			},
			func(buf *[]byte) {
				*buf = (*buf)[:0] // Reset length to 0
			},
		),
		mediumPool: NewObjectPool[[]byte](
			"MediumByteBufferPool",
			mediumSize,
			func() *[]byte {
				buf := make([]byte, 8192) // 8KB
				return &buf
			},
			func(buf *[]byte) {
				*buf = (*buf)[:0] // Reset length to 0
			},
		),
		largePool: NewObjectPool[[]byte](
			"LargeByteBufferPool",
			largeSize,
			func() *[]byte {
				buf := make([]byte, 65536) // 64KB
				return &buf
			},
			func(buf *[]byte) {
				*buf = (*buf)[:0] // Reset length to 0
			},
		),
	}
}

// GetBuffer retrieves a buffer of appropriate size
func (bbp *ByteBufferPool) GetBuffer(size int) *[]byte {
	switch {
	case size <= 1024:
		return bbp.smallPool.Get()
	case size <= 8192:
		return bbp.mediumPool.Get()
	default:
		return bbp.largePool.Get()
	}
}

// PutBuffer returns a buffer to the appropriate pool
func (bbp *ByteBufferPool) PutBuffer(buf *[]byte, size int) {
	if buf == nil {
		return
	}
	
	switch {
	case size <= 1024:
		bbp.smallPool.Put(buf)
	case size <= 8192:
		bbp.mediumPool.Put(buf)
	default:
		bbp.largePool.Put(buf)
	}
}

// MemoryManager coordinates all memory pools
type MemoryManager struct {
	orderPool      *OrderPool
	marketDataPool *MarketDataPool
	byteBufferPool *ByteBufferPool
	
	// Statistics
	totalAllocations   uint64
	totalDeallocations uint64
	gcCount            uint64
	lastGCTime         int64
}

// NewMemoryManager creates a new memory manager
func NewMemoryManager() *MemoryManager {
	return &MemoryManager{
		orderPool:      NewOrderPool(1000),
		marketDataPool: NewMarketDataPool(5000, 1000),
		byteBufferPool: NewByteBufferPool(500, 200, 100),
	}
}

// GetOrderPool returns the order pool
func (mm *MemoryManager) GetOrderPool() *OrderPool {
	return mm.orderPool
}

// GetMarketDataPool returns the market data pool
func (mm *MemoryManager) GetMarketDataPool() *MarketDataPool {
	return mm.marketDataPool
}

// GetByteBufferPool returns the byte buffer pool
func (mm *MemoryManager) GetByteBufferPool() *ByteBufferPool {
	return mm.byteBufferPool
}

// GetGlobalStats returns global memory statistics
func (mm *MemoryManager) GetGlobalStats() map[string]interface{} {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	return map[string]interface{}{
		"heap_alloc":        memStats.HeapAlloc,
		"heap_sys":          memStats.HeapSys,
		"heap_idle":         memStats.HeapIdle,
		"heap_inuse":        memStats.HeapInuse,
		"heap_released":     memStats.HeapReleased,
		"heap_objects":      memStats.HeapObjects,
		"stack_inuse":       memStats.StackInuse,
		"stack_sys":         memStats.StackSys,
		"gc_count":          memStats.NumGC,
		"gc_pause_total":    memStats.PauseTotalNs,
		"gc_pause_last":     memStats.PauseNs[(memStats.NumGC+255)%256],
		"total_allocations": atomic.LoadUint64(&mm.totalAllocations),
		"total_deallocations": atomic.LoadUint64(&mm.totalDeallocations),
		"order_pool":        mm.orderPool.GetStats(),
		"tick_pool":         mm.marketDataPool.GetTickStats(),
		"orderbook_pool":    mm.marketDataPool.GetOrderBookStats(),
	}
}

// ForceGC forces garbage collection and updates statistics
func (mm *MemoryManager) ForceGC() {
	runtime.GC()
	atomic.AddUint64(&mm.gcCount, 1)
	atomic.StoreInt64(&mm.lastGCTime, getCurrentNanoTime())
}

// PreallocateMemory preallocates memory pools to avoid initial allocation overhead
func (mm *MemoryManager) PreallocateMemory() {
	// Preallocate orders
	orders := make([]*GenericOrder, 100)
	for i := 0; i < 100; i++ {
		orders[i] = mm.orderPool.Get()
	}
	for _, order := range orders {
		mm.orderPool.Put(order)
	}
	
	// Preallocate market data ticks
	ticks := make([]*MarketDataTick, 500)
	for i := 0; i < 500; i++ {
		ticks[i] = mm.marketDataPool.GetTick()
	}
	for _, tick := range ticks {
		mm.marketDataPool.PutTick(tick)
	}
	
	// Preallocate order books
	orderBooks := make([]*OrderBookSnapshot, 100)
	for i := 0; i < 100; i++ {
		orderBooks[i] = mm.marketDataPool.GetOrderBook()
	}
	for _, ob := range orderBooks {
		mm.marketDataPool.PutOrderBook(ob)
	}
	
	// Preallocate byte buffers
	smallBuffers := make([]*[]byte, 50)
	for i := 0; i < 50; i++ {
		smallBuffers[i] = mm.byteBufferPool.GetBuffer(1024)
	}
	for _, buf := range smallBuffers {
		mm.byteBufferPool.PutBuffer(buf, 1024)
	}
}

// Reset functions for object cleanup

func resetGenericOrder(order *GenericOrder) {
	// Reset order fields to zero values
	order.ID = ""
	order.Symbol = ""
	order.Side = ""
	order.Type = ""
	order.Quantity = 0
	order.Price = 0
	order.Status = ""
}

func resetMarketDataTick(tick *MarketDataTick) {
	*tick = MarketDataTick{}
}

func resetOrderBookSnapshot(ob *OrderBookSnapshot) {
	ob.Symbol = ""
	ob.Exchange = ""
	ob.Bids = ob.Bids[:0]
	ob.Asks = ob.Asks[:0]
	ob.Timestamp = 0
	ob.ReceiveTime = 0
	ob.SequenceNum = 0
}

// Global memory manager instance
var globalMemoryManager *MemoryManager
var memoryManagerOnce sync.Once

// GetGlobalMemoryManager returns the global memory manager instance
func GetGlobalMemoryManager() *MemoryManager {
	memoryManagerOnce.Do(func() {
		globalMemoryManager = NewMemoryManager()
		globalMemoryManager.PreallocateMemory()
	})
	return globalMemoryManager
}

// Helper functions

func getCurrentNanoTime() int64 {
	// This would typically use a high-resolution timer
	// For now, we'll use a placeholder
	return 0 // time.Now().UnixNano()
}

// NUMA-aware memory allocation (Linux-specific)
// This would require CGO and Linux-specific system calls
// For now, we'll provide the interface

// NUMANode represents a NUMA node
type NUMANode struct {
	ID       int
	CPUs     []int
	Memory   uint64
	Distance []int
}

// NUMATopology represents the NUMA topology
type NUMATopology struct {
	Nodes []NUMANode
}

// GetNUMATopology returns the NUMA topology (placeholder)
func GetNUMATopology() (*NUMATopology, error) {
	// This would query the system for NUMA topology
	// For now, return a simple single-node topology
	return &NUMATopology{
		Nodes: []NUMANode{
			{
				ID:       0,
				CPUs:     []int{0, 1, 2, 3, 4, 5, 6, 7},
				Memory:   16 * 1024 * 1024 * 1024, // 16GB
				Distance: []int{10},
			},
		},
	}, nil
}

// AllocateOnNUMANode allocates memory on a specific NUMA node (placeholder)
func AllocateOnNUMANode(size int, nodeID int) (unsafe.Pointer, error) {
	// This would use Linux mbind() system call
	// For now, use regular allocation
	ptr := unsafe.Pointer(&make([]byte, size)[0])
	return ptr, nil
}

// MemoryPrefetch prefetches memory to cache (placeholder)
func MemoryPrefetch(ptr unsafe.Pointer, size int) {
	// This would use CPU-specific prefetch instructions
	// For now, this is a no-op
}

// HugePage support (Linux-specific, placeholder)
func AllocateHugePage(size int) (unsafe.Pointer, error) {
	// This would use mmap with MAP_HUGETLB
	// For now, use regular allocation
	ptr := unsafe.Pointer(&make([]byte, size)[0])
	return ptr, nil
}
