package lockfree

import (
	"sync/atomic"
	"unsafe"

	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/hft/domain/entities"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/hft/domain/valueobjects"
	"github.com/shopspring/decimal"
)

// OrderBookLevel represents a price level in the order book
type OrderBookLevel struct {
	Price    decimal.Decimal
	Quantity decimal.Decimal
	Count    int32
	Orders   *OrderNode // Linked list of orders at this level
}

// OrderNode represents a single order in the order book
type OrderNode struct {
	Order *entities.Order
	Next  *OrderNode
}

// LockFreeOrderBook implements a lock-free order book for ultra-low latency
type LockFreeOrderBook struct {
	symbol   entities.Symbol
	exchange entities.Exchange
	
	// Atomic pointers to bid and ask trees
	bidTree unsafe.Pointer // *AVLTree
	askTree unsafe.Pointer // *AVLTree
	
	// Sequence number for ordering updates
	sequence uint64
	
	// Statistics
	updateCount uint64
	lastUpdate  int64 // Unix nanoseconds
}

// AVLNode represents a node in the AVL tree for price levels
type AVLNode struct {
	Level  *OrderBookLevel
	Left   unsafe.Pointer // *AVLNode
	Right  unsafe.Pointer // *AVLNode
	Height int32
}

// AVLTree represents a lock-free AVL tree for price levels
type AVLTree struct {
	Root unsafe.Pointer // *AVLNode
	Size int32
}

// NewLockFreeOrderBook creates a new lock-free order book
func NewLockFreeOrderBook(symbol entities.Symbol, exchange entities.Exchange) *LockFreeOrderBook {
	bidTree := &AVLTree{}
	askTree := &AVLTree{}
	
	return &LockFreeOrderBook{
		symbol:   symbol,
		exchange: exchange,
		bidTree:  unsafe.Pointer(bidTree),
		askTree:  unsafe.Pointer(askTree),
		sequence: 0,
	}
}

// AddOrder adds an order to the order book using lock-free operations
func (ob *LockFreeOrderBook) AddOrder(order *entities.Order) bool {
	atomic.AddUint64(&ob.sequence, 1)
	atomic.AddUint64(&ob.updateCount, 1)
	atomic.StoreInt64(&ob.lastUpdate, getCurrentNanoTime())
	
	price := order.GetPrice()
	side := order.GetSide()
	
	var tree *AVLTree
	if side == valueobjects.OrderSideBuy {
		tree = (*AVLTree)(atomic.LoadPointer(&ob.bidTree))
	} else {
		tree = (*AVLTree)(atomic.LoadPointer(&ob.askTree))
	}
	
	return ob.insertOrder(tree, price, order)
}

// RemoveOrder removes an order from the order book
func (ob *LockFreeOrderBook) RemoveOrder(orderID entities.OrderID, side valueobjects.OrderSide, price valueobjects.Price) bool {
	atomic.AddUint64(&ob.sequence, 1)
	atomic.AddUint64(&ob.updateCount, 1)
	atomic.StoreInt64(&ob.lastUpdate, getCurrentNanoTime())
	
	var tree *AVLTree
	if side == valueobjects.OrderSideBuy {
		tree = (*AVLTree)(atomic.LoadPointer(&ob.bidTree))
	} else {
		tree = (*AVLTree)(atomic.LoadPointer(&ob.askTree))
	}
	
	return ob.removeOrder(tree, price, orderID)
}

// GetBestBid returns the best bid price and quantity
func (ob *LockFreeOrderBook) GetBestBid() (valueobjects.Price, valueobjects.Quantity, bool) {
	tree := (*AVLTree)(atomic.LoadPointer(&ob.bidTree))
	if tree == nil {
		return valueobjects.Price{}, valueobjects.Quantity{}, false
	}
	
	// Find the maximum price in bid tree (best bid)
	node := ob.findMax(tree)
	if node == nil {
		return valueobjects.Price{}, valueobjects.Quantity{}, false
	}
	
	level := node.Level
	return valueobjects.Price{Decimal: level.Price}, 
		   valueobjects.Quantity{Decimal: level.Quantity}, 
		   true
}

// GetBestAsk returns the best ask price and quantity
func (ob *LockFreeOrderBook) GetBestAsk() (valueobjects.Price, valueobjects.Quantity, bool) {
	tree := (*AVLTree)(atomic.LoadPointer(&ob.askTree))
	if tree == nil {
		return valueobjects.Price{}, valueobjects.Quantity{}, false
	}
	
	// Find the minimum price in ask tree (best ask)
	node := ob.findMin(tree)
	if node == nil {
		return valueobjects.Price{}, valueobjects.Quantity{}, false
	}
	
	level := node.Level
	return valueobjects.Price{Decimal: level.Price}, 
		   valueobjects.Quantity{Decimal: level.Quantity}, 
		   true
}

// GetSpread returns the bid-ask spread
func (ob *LockFreeOrderBook) GetSpread() (valueobjects.Price, bool) {
	bestBid, _, hasBid := ob.GetBestBid()
	bestAsk, _, hasAsk := ob.GetBestAsk()
	
	if !hasBid || !hasAsk {
		return valueobjects.Price{}, false
	}
	
	spread := bestAsk.Sub(bestBid.Decimal)
	return valueobjects.Price{Decimal: spread}, true
}

// GetMidPrice returns the mid price
func (ob *LockFreeOrderBook) GetMidPrice() (valueobjects.Price, bool) {
	bestBid, _, hasBid := ob.GetBestBid()
	bestAsk, _, hasAsk := ob.GetBestAsk()
	
	if !hasBid || !hasAsk {
		return valueobjects.Price{}, false
	}
	
	mid := bestBid.Add(bestAsk.Decimal).Div(decimal.NewFromInt(2))
	return valueobjects.Price{Decimal: mid}, true
}

// GetDepth returns the order book depth for a given number of levels
func (ob *LockFreeOrderBook) GetDepth(levels int) ([]OrderBookLevel, []OrderBookLevel) {
	bidTree := (*AVLTree)(atomic.LoadPointer(&ob.bidTree))
	askTree := (*AVLTree)(atomic.LoadPointer(&ob.askTree))
	
	bids := ob.getTopLevels(bidTree, levels, true)  // Descending for bids
	asks := ob.getTopLevels(askTree, levels, false) // Ascending for asks
	
	return bids, asks
}

// GetSequence returns the current sequence number
func (ob *LockFreeOrderBook) GetSequence() uint64 {
	return atomic.LoadUint64(&ob.sequence)
}

// GetUpdateCount returns the total number of updates
func (ob *LockFreeOrderBook) GetUpdateCount() uint64 {
	return atomic.LoadUint64(&ob.updateCount)
}

// GetLastUpdateTime returns the last update time in nanoseconds
func (ob *LockFreeOrderBook) GetLastUpdateTime() int64 {
	return atomic.LoadInt64(&ob.lastUpdate)
}

// insertOrder inserts an order into the AVL tree
func (ob *LockFreeOrderBook) insertOrder(tree *AVLTree, price valueobjects.Price, order *entities.Order) bool {
	for {
		root := (*AVLNode)(atomic.LoadPointer(&tree.Root))
		newRoot, success := ob.insertNode(root, price, order)
		
		if atomic.CompareAndSwapPointer(&tree.Root, unsafe.Pointer(root), unsafe.Pointer(newRoot)) {
			if success {
				atomic.AddInt32(&tree.Size, 1)
			}
			return success
		}
		// Retry if CAS failed
	}
}

// removeOrder removes an order from the AVL tree
func (ob *LockFreeOrderBook) removeOrder(tree *AVLTree, price valueobjects.Price, orderID entities.OrderID) bool {
	for {
		root := (*AVLNode)(atomic.LoadPointer(&tree.Root))
		newRoot, success := ob.removeNode(root, price, orderID)
		
		if atomic.CompareAndSwapPointer(&tree.Root, unsafe.Pointer(root), unsafe.Pointer(newRoot)) {
			if success {
				atomic.AddInt32(&tree.Size, -1)
			}
			return success
		}
		// Retry if CAS failed
	}
}

// insertNode inserts a node into the AVL tree (returns new root)
func (ob *LockFreeOrderBook) insertNode(node *AVLNode, price valueobjects.Price, order *entities.Order) (*AVLNode, bool) {
	if node == nil {
		// Create new level
		level := &OrderBookLevel{
			Price:    price.Decimal,
			Quantity: order.GetQuantity().Decimal,
			Count:    1,
			Orders: &OrderNode{
				Order: order,
				Next:  nil,
			},
		}
		
		return &AVLNode{
			Level:  level,
			Left:   nil,
			Right:  nil,
			Height: 1,
		}, true
	}
	
	cmp := price.Decimal.Cmp(node.Level.Price)
	
	if cmp == 0 {
		// Same price level, add order to the list
		return ob.addOrderToLevel(node, order), true
	} else if cmp < 0 {
		// Insert into left subtree
		left := (*AVLNode)(atomic.LoadPointer(&node.Left))
		newLeft, success := ob.insertNode(left, price, order)
		
		newNode := &AVLNode{
			Level:  node.Level,
			Left:   unsafe.Pointer(newLeft),
			Right:  node.Right,
			Height: node.Height,
		}
		
		return ob.rebalance(newNode), success
	} else {
		// Insert into right subtree
		right := (*AVLNode)(atomic.LoadPointer(&node.Right))
		newRight, success := ob.insertNode(right, price, order)
		
		newNode := &AVLNode{
			Level:  node.Level,
			Left:   node.Left,
			Right:  unsafe.Pointer(newRight),
			Height: node.Height,
		}
		
		return ob.rebalance(newNode), success
	}
}

// removeNode removes a node from the AVL tree (returns new root)
func (ob *LockFreeOrderBook) removeNode(node *AVLNode, price valueobjects.Price, orderID entities.OrderID) (*AVLNode, bool) {
	if node == nil {
		return nil, false
	}
	
	cmp := price.Decimal.Cmp(node.Level.Price)
	
	if cmp < 0 {
		// Remove from left subtree
		left := (*AVLNode)(atomic.LoadPointer(&node.Left))
		newLeft, success := ob.removeNode(left, price, orderID)
		
		if !success {
			return node, false
		}
		
		newNode := &AVLNode{
			Level:  node.Level,
			Left:   unsafe.Pointer(newLeft),
			Right:  node.Right,
			Height: node.Height,
		}
		
		return ob.rebalance(newNode), true
	} else if cmp > 0 {
		// Remove from right subtree
		right := (*AVLNode)(atomic.LoadPointer(&node.Right))
		newRight, success := ob.removeNode(right, price, orderID)
		
		if !success {
			return node, false
		}
		
		newNode := &AVLNode{
			Level:  node.Level,
			Left:   node.Left,
			Right:  unsafe.Pointer(newRight),
			Height: node.Height,
		}
		
		return ob.rebalance(newNode), true
	} else {
		// Found the price level, remove order from the list
		newNode, removed := ob.removeOrderFromLevel(node, orderID)
		if !removed {
			return node, false
		}
		
		// If level is empty, remove the node
		if newNode.Level.Count == 0 {
			return ob.deleteNode(newNode), true
		}
		
		return newNode, true
	}
}

// addOrderToLevel adds an order to an existing price level
func (ob *LockFreeOrderBook) addOrderToLevel(node *AVLNode, order *entities.Order) *AVLNode {
	// Create new order node
	orderNode := &OrderNode{
		Order: order,
		Next:  node.Level.Orders,
	}
	
	// Create new level with updated values
	newLevel := &OrderBookLevel{
		Price:    node.Level.Price,
		Quantity: node.Level.Quantity.Add(order.GetQuantity().Decimal),
		Count:    node.Level.Count + 1,
		Orders:   orderNode,
	}
	
	return &AVLNode{
		Level:  newLevel,
		Left:   node.Left,
		Right:  node.Right,
		Height: node.Height,
	}
}

// removeOrderFromLevel removes an order from a price level
func (ob *LockFreeOrderBook) removeOrderFromLevel(node *AVLNode, orderID entities.OrderID) (*AVLNode, bool) {
	var newOrders *OrderNode
	var removedQuantity decimal.Decimal
	var found bool
	
	// Traverse the order list and remove the target order
	current := node.Level.Orders
	var prev *OrderNode
	
	for current != nil {
		if current.Order.GetID() == orderID {
			// Found the order to remove
			removedQuantity = current.Order.GetQuantity().Decimal
			found = true
			
			if prev == nil {
				// Removing the first order
				newOrders = current.Next
			} else {
				// Create new list without the removed order
				newOrders = ob.copyOrderListExcept(node.Level.Orders, orderID)
			}
			break
		}
		prev = current
		current = current.Next
	}
	
	if !found {
		return node, false
	}
	
	// Create new level with updated values
	newLevel := &OrderBookLevel{
		Price:    node.Level.Price,
		Quantity: node.Level.Quantity.Sub(removedQuantity),
		Count:    node.Level.Count - 1,
		Orders:   newOrders,
	}
	
	return &AVLNode{
		Level:  newLevel,
		Left:   node.Left,
		Right:  node.Right,
		Height: node.Height,
	}, true
}

// Helper functions for AVL tree operations
func (ob *LockFreeOrderBook) getHeight(node *AVLNode) int32 {
	if node == nil {
		return 0
	}
	return node.Height
}

func (ob *LockFreeOrderBook) updateHeight(node *AVLNode) {
	if node == nil {
		return
	}
	
	leftHeight := ob.getHeight((*AVLNode)(atomic.LoadPointer(&node.Left)))
	rightHeight := ob.getHeight((*AVLNode)(atomic.LoadPointer(&node.Right)))
	
	if leftHeight > rightHeight {
		node.Height = leftHeight + 1
	} else {
		node.Height = rightHeight + 1
	}
}

func (ob *LockFreeOrderBook) getBalance(node *AVLNode) int32 {
	if node == nil {
		return 0
	}
	
	leftHeight := ob.getHeight((*AVLNode)(atomic.LoadPointer(&node.Left)))
	rightHeight := ob.getHeight((*AVLNode)(atomic.LoadPointer(&node.Right)))
	
	return leftHeight - rightHeight
}

func (ob *LockFreeOrderBook) rebalance(node *AVLNode) *AVLNode {
	if node == nil {
		return nil
	}
	
	ob.updateHeight(node)
	balance := ob.getBalance(node)
	
	// Left heavy
	if balance > 1 {
		left := (*AVLNode)(atomic.LoadPointer(&node.Left))
		if ob.getBalance(left) < 0 {
			// Left-Right case
			node.Left = unsafe.Pointer(ob.rotateLeft(left))
		}
		// Left-Left case
		return ob.rotateRight(node)
	}
	
	// Right heavy
	if balance < -1 {
		right := (*AVLNode)(atomic.LoadPointer(&node.Right))
		if ob.getBalance(right) > 0 {
			// Right-Left case
			node.Right = unsafe.Pointer(ob.rotateRight(right))
		}
		// Right-Right case
		return ob.rotateLeft(node)
	}
	
	return node
}

func (ob *LockFreeOrderBook) rotateLeft(node *AVLNode) *AVLNode {
	right := (*AVLNode)(atomic.LoadPointer(&node.Right))
	if right == nil {
		return node
	}
	
	newNode := &AVLNode{
		Level:  node.Level,
		Left:   node.Left,
		Right:  right.Left,
		Height: node.Height,
	}
	
	ob.updateHeight(newNode)
	
	result := &AVLNode{
		Level:  right.Level,
		Left:   unsafe.Pointer(newNode),
		Right:  right.Right,
		Height: right.Height,
	}
	
	ob.updateHeight(result)
	return result
}

func (ob *LockFreeOrderBook) rotateRight(node *AVLNode) *AVLNode {
	left := (*AVLNode)(atomic.LoadPointer(&node.Left))
	if left == nil {
		return node
	}
	
	newNode := &AVLNode{
		Level:  node.Level,
		Left:   left.Right,
		Right:  node.Right,
		Height: node.Height,
	}
	
	ob.updateHeight(newNode)
	
	result := &AVLNode{
		Level:  left.Level,
		Left:   left.Left,
		Right:  unsafe.Pointer(newNode),
		Height: left.Height,
	}
	
	ob.updateHeight(result)
	return result
}

func (ob *LockFreeOrderBook) findMin(tree *AVLTree) *AVLNode {
	if tree == nil {
		return nil
	}
	
	node := (*AVLNode)(atomic.LoadPointer(&tree.Root))
	if node == nil {
		return nil
	}
	
	for {
		left := (*AVLNode)(atomic.LoadPointer(&node.Left))
		if left == nil {
			break
		}
		node = left
	}
	
	return node
}

func (ob *LockFreeOrderBook) findMax(tree *AVLTree) *AVLNode {
	if tree == nil {
		return nil
	}
	
	node := (*AVLNode)(atomic.LoadPointer(&tree.Root))
	if node == nil {
		return nil
	}
	
	for {
		right := (*AVLNode)(atomic.LoadPointer(&node.Right))
		if right == nil {
			break
		}
		node = right
	}
	
	return node
}

func (ob *LockFreeOrderBook) deleteNode(node *AVLNode) *AVLNode {
	left := (*AVLNode)(atomic.LoadPointer(&node.Left))
	right := (*AVLNode)(atomic.LoadPointer(&node.Right))
	
	if left == nil {
		return right
	}
	if right == nil {
		return left
	}
	
	// Find inorder successor (minimum in right subtree)
	successor := ob.findMinNode(right)
	
	// Create new node with successor's data
	newNode := &AVLNode{
		Level:  successor.Level,
		Left:   node.Left,
		Right:  unsafe.Pointer(ob.removeMinNode(right)),
		Height: node.Height,
	}
	
	return ob.rebalance(newNode)
}

func (ob *LockFreeOrderBook) findMinNode(node *AVLNode) *AVLNode {
	for {
		left := (*AVLNode)(atomic.LoadPointer(&node.Left))
		if left == nil {
			break
		}
		node = left
	}
	return node
}

func (ob *LockFreeOrderBook) removeMinNode(node *AVLNode) *AVLNode {
	if node == nil {
		return nil
	}
	
	left := (*AVLNode)(atomic.LoadPointer(&node.Left))
	if left == nil {
		return (*AVLNode)(atomic.LoadPointer(&node.Right))
	}
	
	newNode := &AVLNode{
		Level:  node.Level,
		Left:   unsafe.Pointer(ob.removeMinNode(left)),
		Right:  node.Right,
		Height: node.Height,
	}
	
	return ob.rebalance(newNode)
}

func (ob *LockFreeOrderBook) copyOrderListExcept(head *OrderNode, excludeID entities.OrderID) *OrderNode {
	if head == nil {
		return nil
	}
	
	if head.Order.GetID() == excludeID {
		return ob.copyOrderListExcept(head.Next, excludeID)
	}
	
	return &OrderNode{
		Order: head.Order,
		Next:  ob.copyOrderListExcept(head.Next, excludeID),
	}
}

func (ob *LockFreeOrderBook) getTopLevels(tree *AVLTree, count int, descending bool) []OrderBookLevel {
	if tree == nil || count <= 0 {
		return nil
	}
	
	levels := make([]OrderBookLevel, 0, count)
	ob.inorderTraversal((*AVLNode)(atomic.LoadPointer(&tree.Root)), &levels, count, descending)
	
	return levels
}

func (ob *LockFreeOrderBook) inorderTraversal(node *AVLNode, levels *[]OrderBookLevel, maxCount int, descending bool) {
	if node == nil || len(*levels) >= maxCount {
		return
	}
	
	if descending {
		// Right -> Root -> Left for descending order
		ob.inorderTraversal((*AVLNode)(atomic.LoadPointer(&node.Right)), levels, maxCount, descending)
		
		if len(*levels) < maxCount {
			*levels = append(*levels, *node.Level)
		}
		
		ob.inorderTraversal((*AVLNode)(atomic.LoadPointer(&node.Left)), levels, maxCount, descending)
	} else {
		// Left -> Root -> Right for ascending order
		ob.inorderTraversal((*AVLNode)(atomic.LoadPointer(&node.Left)), levels, maxCount, descending)
		
		if len(*levels) < maxCount {
			*levels = append(*levels, *node.Level)
		}
		
		ob.inorderTraversal((*AVLNode)(atomic.LoadPointer(&node.Right)), levels, maxCount, descending)
	}
}

// getCurrentNanoTime returns current time in nanoseconds
func getCurrentNanoTime() int64 {
	// This would typically use a high-resolution timer
	// For now, we'll use a placeholder
	return 0 // time.Now().UnixNano()
}
