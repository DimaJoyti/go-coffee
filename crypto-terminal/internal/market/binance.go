package market

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/config"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/models"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

// BinanceProvider implements the Provider interface for Binance API
type BinanceProvider struct {
	config     config.BinanceConfig
	httpClient *http.Client
	isHealthy  bool
}

// BinanceTickerResponse represents the response from Binance ticker API
type BinanceTickerResponse struct {
	Symbol             string `json:"symbol"`
	PriceChange        string `json:"priceChange"`
	PriceChangePercent string `json:"priceChangePercent"`
	WeightedAvgPrice   string `json:"weightedAvgPrice"`
	PrevClosePrice     string `json:"prevClosePrice"`
	LastPrice          string `json:"lastPrice"`
	LastQty            string `json:"lastQty"`
	BidPrice           string `json:"bidPrice"`
	BidQty             string `json:"bidQty"`
	AskPrice           string `json:"askPrice"`
	AskQty             string `json:"askQty"`
	OpenPrice          string `json:"openPrice"`
	HighPrice          string `json:"highPrice"`
	LowPrice           string `json:"lowPrice"`
	Volume             string `json:"volume"`
	QuoteVolume        string `json:"quoteVolume"`
	OpenTime           int64  `json:"openTime"`
	CloseTime          int64  `json:"closeTime"`
	FirstId            int64  `json:"firstId"`
	LastId             int64  `json:"lastId"`
	Count              int64  `json:"count"`
}

// BinanceKlineResponse represents the response from Binance klines API
type BinanceKlineResponse []interface{}

// NewBinanceProvider creates a new Binance provider
func NewBinanceProvider(config config.BinanceConfig, httpClient *http.Client) *BinanceProvider {
	return &BinanceProvider{
		config:     config,
		httpClient: httpClient,
		isHealthy:  true,
	}
}

// GetPrice returns the current price for a specific symbol
func (p *BinanceProvider) GetPrice(ctx context.Context, symbol string) (*models.Price, error) {
	// Convert symbol to Binance format (e.g., bitcoin -> BTCUSDT)
	binanceSymbol := p.convertSymbolToBinance(symbol)
	
	url := fmt.Sprintf("%s/ticker/24hr?symbol=%s", p.config.RestURL, binanceSymbol)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		p.isHealthy = false
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		p.isHealthy = false
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result BinanceTickerResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	lastPrice, err := decimal.NewFromString(result.LastPrice)
	if err != nil {
		return nil, fmt.Errorf("failed to parse last price: %w", err)
	}

	quoteVolume, err := decimal.NewFromString(result.QuoteVolume)
	if err != nil {
		return nil, fmt.Errorf("failed to parse quote volume: %w", err)
	}

	priceChangePercent, err := decimal.NewFromString(result.PriceChangePercent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse price change percent: %w", err)
	}

	price := &models.Price{
		Symbol:    p.convertSymbolFromBinance(result.Symbol),
		Price:     lastPrice,
		Volume24h: quoteVolume,
		Change24h: priceChangePercent,
		Timestamp: time.Now(),
		Source:    "binance",
	}

	p.isHealthy = true
	return price, nil
}

// GetPrices returns prices for multiple symbols
func (p *BinanceProvider) GetPrices(ctx context.Context, symbols []string) ([]*models.Price, error) {
	if len(symbols) == 0 {
		return nil, fmt.Errorf("no symbols provided")
	}

	// Get all tickers
	url := fmt.Sprintf("%s/ticker/24hr", p.config.RestURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		p.isHealthy = false
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		p.isHealthy = false
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result []BinanceTickerResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Create a map of requested symbols
	symbolMap := make(map[string]bool)
	for _, symbol := range symbols {
		binanceSymbol := p.convertSymbolToBinance(symbol)
		symbolMap[binanceSymbol] = true
	}

	var prices []*models.Price
	for _, ticker := range result {
		if symbolMap[ticker.Symbol] {
			lastPrice, err := decimal.NewFromString(ticker.LastPrice)
			if err != nil {
				logrus.Warnf("Failed to parse last price for %s: %v", ticker.Symbol, err)
				continue
			}

			quoteVolume, err := decimal.NewFromString(ticker.QuoteVolume)
			if err != nil {
				logrus.Warnf("Failed to parse quote volume for %s: %v", ticker.Symbol, err)
				continue
			}

			priceChangePercent, err := decimal.NewFromString(ticker.PriceChangePercent)
			if err != nil {
				logrus.Warnf("Failed to parse price change percent for %s: %v", ticker.Symbol, err)
				continue
			}

			price := &models.Price{
				Symbol:    p.convertSymbolFromBinance(ticker.Symbol),
				Price:     lastPrice,
				Volume24h: quoteVolume,
				Change24h: priceChangePercent,
				Timestamp: time.Now(),
				Source:    "binance",
			}
			prices = append(prices, price)
		}
	}

	p.isHealthy = true
	return prices, nil
}

// GetPriceHistory returns historical price data
func (p *BinanceProvider) GetPriceHistory(ctx context.Context, symbol, timeframe string, limit int) ([]*models.OHLCV, error) {
	binanceSymbol := p.convertSymbolToBinance(symbol)
	binanceInterval := p.convertTimeframeToBinance(timeframe)

	url := fmt.Sprintf("%s/klines?symbol=%s&interval=%s&limit=%d",
		p.config.RestURL, binanceSymbol, binanceInterval, limit)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		p.isHealthy = false
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		p.isHealthy = false
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result []BinanceKlineResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var ohlcv []*models.OHLCV
	for _, kline := range result {
		if len(kline) >= 6 {
			openTime, _ := kline[0].(float64)
			openPrice, _ := kline[1].(string)
			highPrice, _ := kline[2].(string)
			lowPrice, _ := kline[3].(string)
			closePrice, _ := kline[4].(string)
			volume, _ := kline[5].(string)

			open, _ := decimal.NewFromString(openPrice)
			high, _ := decimal.NewFromString(highPrice)
			low, _ := decimal.NewFromString(lowPrice)
			close, _ := decimal.NewFromString(closePrice)
			vol, _ := decimal.NewFromString(volume)

			candle := &models.OHLCV{
				Symbol:    p.convertSymbolFromBinance(binanceSymbol),
				Timeframe: timeframe,
				Open:      open,
				High:      high,
				Low:       low,
				Close:     close,
				Volume:    vol,
				Timestamp: time.Unix(int64(openTime)/1000, 0),
			}
			ohlcv = append(ohlcv, candle)
		}
	}

	p.isHealthy = true
	return ohlcv, nil
}

// GetMarketData returns comprehensive market data for a symbol
func (p *BinanceProvider) GetMarketData(ctx context.Context, symbol string) (*models.MarketData, error) {
	// Binance doesn't provide comprehensive market data like market cap, etc.
	// This would need to be combined with other providers
	return nil, fmt.Errorf("market data not available from Binance provider")
}

// IsHealthy returns the health status of the provider
func (p *BinanceProvider) IsHealthy() bool {
	return p.isHealthy
}

// convertSymbolToBinance converts a symbol to Binance format
func (p *BinanceProvider) convertSymbolToBinance(symbol string) string {
	symbolMap := map[string]string{
		"bitcoin":     "BTCUSDT",
		"ethereum":    "ETHUSDT",
		"binancecoin": "BNBUSDT",
		"cardano":     "ADAUSDT",
		"solana":      "SOLUSDT",
		"polkadot":    "DOTUSDT",
		"dogecoin":    "DOGEUSDT",
		"avalanche-2": "AVAXUSDT",
		"polygon":     "MATICUSDT",
		"chainlink":   "LINKUSDT",
	}

	if binanceSymbol, exists := symbolMap[strings.ToLower(symbol)]; exists {
		return binanceSymbol
	}

	// Default: assume it's already in the correct format
	return strings.ToUpper(symbol) + "USDT"
}

// convertSymbolFromBinance converts a Binance symbol back to standard format
func (p *BinanceProvider) convertSymbolFromBinance(binanceSymbol string) string {
	symbolMap := map[string]string{
		"BTCUSDT":  "BTC",
		"ETHUSDT":  "ETH",
		"BNBUSDT":  "BNB",
		"ADAUSDT":  "ADA",
		"SOLUSDT":  "SOL",
		"DOTUSDT":  "DOT",
		"DOGEUSDT": "DOGE",
		"AVAXUSDT": "AVAX",
		"MATICUSDT": "MATIC",
		"LINKUSDT": "LINK",
	}

	if symbol, exists := symbolMap[binanceSymbol]; exists {
		return symbol
	}

	// Default: remove USDT suffix
	return strings.TrimSuffix(binanceSymbol, "USDT")
}

// convertTimeframeToBinance converts timeframe to Binance interval format
func (p *BinanceProvider) convertTimeframeToBinance(timeframe string) string {
	intervalMap := map[string]string{
		"1m":  "1m",
		"5m":  "5m",
		"15m": "15m",
		"1h":  "1h",
		"4h":  "4h",
		"1d":  "1d",
		"1w":  "1w",
	}

	if interval, exists := intervalMap[timeframe]; exists {
		return interval
	}

	return "1h" // Default to 1 hour
}
