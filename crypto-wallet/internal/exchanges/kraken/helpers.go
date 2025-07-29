package kraken

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"net/url"
	"time"

	"github.com/shopspring/decimal"
)

// Rate limiter implementation
func NewRateLimiter(config RateConfig) *RateLimiter {
	return &RateLimiter{
		tokens:     config.BurstSize,
		maxTokens:  config.BurstSize,
		refillRate: time.Second / time.Duration(config.RequestsPerSecond),
		lastRefill: time.Now(),
	}
}

func (rl *RateLimiter) Wait() error {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastRefill)
	tokensToAdd := int(elapsed / rl.refillRate)

	if tokensToAdd > 0 {
		rl.tokens += tokensToAdd
		if rl.tokens > rl.maxTokens {
			rl.tokens = rl.maxTokens
		}
		rl.lastRefill = now
	}

	if rl.tokens <= 0 {
		return fmt.Errorf("rate limit exceeded")
	}

	rl.tokens--
	return nil
}

// Signature creation for Kraken API
func (kc *KrakenClient) createSignature(endpoint string, params url.Values, nonce string) (string, error) {
	// Decode API secret
	secretBytes, err := base64.StdEncoding.DecodeString(kc.config.APISecret)
	if err != nil {
		return "", fmt.Errorf("failed to decode API secret: %w", err)
	}

	// Create message
	message := params.Encode()
	sha := sha256.Sum256([]byte(nonce + message))
	mac := hmac.New(sha512.New, secretBytes)
	mac.Write([]byte(endpoint))
	mac.Write(sha[:])

	// Return base64 encoded signature
	return base64.StdEncoding.EncodeToString(mac.Sum(nil)), nil
}

// Response parsing methods
func (kc *KrakenClient) parseTickerResponse(symbol string, response map[string]interface{}) *Ticker {
	result, ok := response["result"].(map[string]interface{})
	if !ok {
		return nil
	}

	tickerData, ok := result[symbol].(map[string]interface{})
	if !ok {
		return nil
	}

	ticker := &Ticker{
		Symbol:    symbol,
		Timestamp: time.Now(),
	}

	// Parse ticker fields (simplified)
	if ask, ok := tickerData["a"].([]interface{}); ok && len(ask) > 0 {
		if askPrice, err := decimal.NewFromString(ask[0].(string)); err == nil {
			ticker.Ask = askPrice
		}
	}

	if bid, ok := tickerData["b"].([]interface{}); ok && len(bid) > 0 {
		if bidPrice, err := decimal.NewFromString(bid[0].(string)); err == nil {
			ticker.Bid = bidPrice
		}
	}

	if last, ok := tickerData["c"].([]interface{}); ok && len(last) > 0 {
		if lastPrice, err := decimal.NewFromString(last[0].(string)); err == nil {
			ticker.Last = lastPrice
		}
	}

	if vol, ok := tickerData["v"].([]interface{}); ok && len(vol) > 0 {
		if volume, err := decimal.NewFromString(vol[0].(string)); err == nil {
			ticker.Volume = volume
		}
	}

	if high, ok := tickerData["h"].([]interface{}); ok && len(high) > 0 {
		if highPrice, err := decimal.NewFromString(high[0].(string)); err == nil {
			ticker.High = highPrice
		}
	}

	if low, ok := tickerData["l"].([]interface{}); ok && len(low) > 0 {
		if lowPrice, err := decimal.NewFromString(low[0].(string)); err == nil {
			ticker.Low = lowPrice
		}
	}

	if open, ok := tickerData["o"].(string); ok {
		if openPrice, err := decimal.NewFromString(open); err == nil {
			ticker.Open = openPrice
		}
	}

	return ticker
}

func (kc *KrakenClient) parseOrderBookResponse(symbol string, response map[string]interface{}) *OrderBook {
	result, ok := response["result"].(map[string]interface{})
	if !ok {
		return nil
	}

	bookData, ok := result[symbol].(map[string]interface{})
	if !ok {
		return nil
	}

	orderBook := &OrderBook{
		Symbol:    symbol,
		Timestamp: time.Now(),
		Asks:      []BookEntry{},
		Bids:      []BookEntry{},
	}

	// Parse asks
	if asks, ok := bookData["asks"].([]interface{}); ok {
		for _, ask := range asks {
			if askData, ok := ask.([]interface{}); ok && len(askData) >= 2 {
				price, _ := decimal.NewFromString(askData[0].(string))
				volume, _ := decimal.NewFromString(askData[1].(string))
				orderBook.Asks = append(orderBook.Asks, BookEntry{
					Price:     price,
					Volume:    volume,
					Timestamp: time.Now(),
				})
			}
		}
	}

	// Parse bids
	if bids, ok := bookData["bids"].([]interface{}); ok {
		for _, bid := range bids {
			if bidData, ok := bid.([]interface{}); ok && len(bidData) >= 2 {
				price, _ := decimal.NewFromString(bidData[0].(string))
				volume, _ := decimal.NewFromString(bidData[1].(string))
				orderBook.Bids = append(orderBook.Bids, BookEntry{
					Price:     price,
					Volume:    volume,
					Timestamp: time.Now(),
				})
			}
		}
	}

	return orderBook
}

func (kc *KrakenClient) parseTradesResponse(symbol string, response map[string]interface{}) []Trade {
	result, ok := response["result"].(map[string]interface{})
	if !ok {
		return nil
	}

	tradesData, ok := result[symbol].([]interface{})
	if !ok {
		return nil
	}

	var trades []Trade
	for i, tradeData := range tradesData {
		if trade, ok := tradeData.([]interface{}); ok && len(trade) >= 6 {
			price, _ := decimal.NewFromString(trade[0].(string))
			volume, _ := decimal.NewFromString(trade[1].(string))
			timestamp := time.Unix(int64(trade[2].(float64)), 0)
			side := trade[3].(string)

			trades = append(trades, Trade{
				ID:        fmt.Sprintf("%s_%d", symbol, i),
				Symbol:    symbol,
				Price:     price,
				Volume:    volume,
				Side:      side,
				Timestamp: timestamp,
			})
		}
	}

	return trades
}

func (kc *KrakenClient) parseOHLCVResponse(symbol, interval string, response map[string]interface{}) []OHLCV {
	result, ok := response["result"].(map[string]interface{})
	if !ok {
		return nil
	}

	ohlcvData, ok := result[symbol].([]interface{})
	if !ok {
		return nil
	}

	var ohlcv []OHLCV
	for _, data := range ohlcvData {
		if candle, ok := data.([]interface{}); ok && len(candle) >= 7 {
			timestamp := time.Unix(int64(candle[0].(float64)), 0)
			open, _ := decimal.NewFromString(candle[1].(string))
			high, _ := decimal.NewFromString(candle[2].(string))
			low, _ := decimal.NewFromString(candle[3].(string))
			close, _ := decimal.NewFromString(candle[4].(string))
			volume, _ := decimal.NewFromString(candle[6].(string))

			ohlcv = append(ohlcv, OHLCV{
				Symbol:    symbol,
				Open:      open,
				High:      high,
				Low:       low,
				Close:     close,
				Volume:    volume,
				Timestamp: timestamp,
				Interval:  interval,
			})
		}
	}

	return ohlcv
}

func (kc *KrakenClient) parseBalancesResponse(response map[string]interface{}) []Balance {
	result, ok := response["result"].(map[string]interface{})
	if !ok {
		return nil
	}

	var balances []Balance
	for currency, balanceData := range result {
		if balanceStr, ok := balanceData.(string); ok {
			if balance, err := decimal.NewFromString(balanceStr); err == nil {
				balances = append(balances, Balance{
					Currency:  currency,
					Available: balance,
					Reserved:  decimal.Zero,
					Total:     balance,
				})
			}
		}
	}

	return balances
}

func (kc *KrakenClient) parseOrderResponse(response map[string]interface{}) *Order {
	result, ok := response["result"].(map[string]interface{})
	if !ok {
		return nil
	}

	// Parse order response (simplified)
	order := &Order{
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	if txid, ok := result["txid"].([]interface{}); ok && len(txid) > 0 {
		order.ID = txid[0].(string)
	}

	return order
}

func (kc *KrakenClient) parseOrdersResponse(response map[string]interface{}) []Order {
	result, ok := response["result"].(map[string]interface{})
	if !ok {
		return nil
	}

	var orders []Order
	if openOrders, ok := result["open"].(map[string]interface{}); ok {
		for orderID, orderData := range openOrders {
			if orderInfo, ok := orderData.(map[string]interface{}); ok {
				order := kc.parseOrderInfo(orderID, orderInfo)
				orders = append(orders, *order)
			}
		}
	}

	return orders
}

func (kc *KrakenClient) parseOrderInfo(orderID string, orderInfo map[string]interface{}) *Order {
	order := &Order{
		ID:        orderID,
		Status:    "open",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	// Parse order fields (simplified)
	if descr, ok := orderInfo["descr"].(map[string]interface{}); ok {
		if pair, ok := descr["pair"].(string); ok {
			order.Symbol = pair
		}
		if side, ok := descr["type"].(string); ok {
			order.Side = side
		}
		if orderType, ok := descr["ordertype"].(string); ok {
			order.Type = orderType
		}
		if price, ok := descr["price"].(string); ok {
			if p, err := decimal.NewFromString(price); err == nil {
				order.Price = p
			}
		}
	}

	if vol, ok := orderInfo["vol"].(string); ok {
		if amount, err := decimal.NewFromString(vol); err == nil {
			order.Amount = amount
		}
	}

	if volExec, ok := orderInfo["vol_exec"].(string); ok {
		if filled, err := decimal.NewFromString(volExec); err == nil {
			order.FilledAmount = filled
			order.RemainingAmount = order.Amount.Sub(filled)
		}
	}

	return order
}

// Event type string conversion
func (kc *KrakenClient) getEventTypeString(eventType EventType) string {
	switch eventType {
	case EventTypeTicker:
		return "ticker"
	case EventTypeOrderBook:
		return "orderbook"
	case EventTypeTrade:
		return "trade"
	case EventTypeOHLCV:
		return "ohlcv"
	case EventTypeOrderUpdate:
		return "order_update"
	case EventTypeBalanceUpdate:
		return "balance_update"
	case EventTypePositionUpdate:
		return "position_update"
	case EventTypeError:
		return "error"
	case EventTypeConnection:
		return "connection"
	case EventTypeDisconnection:
		return "disconnection"
	default:
		return "unknown"
	}
}

// Public interface methods
func (kc *KrakenClient) IsRunning() bool {
	kc.mutex.RLock()
	defer kc.mutex.RUnlock()
	return kc.isRunning
}

func (kc *KrakenClient) GetConfig() KrakenConfig {
	return kc.config
}

func (kc *KrakenClient) AddEventHandler(eventType EventType, handler EventHandler) {
	kc.handlerMutex.Lock()
	defer kc.handlerMutex.Unlock()
	kc.eventHandlers[eventType] = append(kc.eventHandlers[eventType], handler)
}

func (kc *KrakenClient) GetSubscriptions() map[string]*Subscription {
	kc.subMutex.RLock()
	defer kc.subMutex.RUnlock()

	subscriptions := make(map[string]*Subscription)
	for id, sub := range kc.subscriptions {
		subscriptions[id] = sub
	}
	return subscriptions
}
