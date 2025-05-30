package market

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/config"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/models"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

// CoinGeckoProvider implements the Provider interface for CoinGecko API
type CoinGeckoProvider struct {
	config     config.CoinGeckoConfig
	httpClient *http.Client
	isHealthy  bool
}

// CoinGeckoResponse represents the response from CoinGecko API
type CoinGeckoResponse struct {
	ID                string  `json:"id"`
	Symbol            string  `json:"symbol"`
	Name              string  `json:"name"`
	CurrentPrice      float64 `json:"current_price"`
	MarketCap         float64 `json:"market_cap"`
	MarketCapRank     int     `json:"market_cap_rank"`
	TotalVolume       float64 `json:"total_volume"`
	High24h           float64 `json:"high_24h"`
	Low24h            float64 `json:"low_24h"`
	PriceChange24h    float64 `json:"price_change_24h"`
	PriceChangePercent24h float64 `json:"price_change_percentage_24h"`
	PriceChangePercent7d  float64 `json:"price_change_percentage_7d"`
	PriceChangePercent30d float64 `json:"price_change_percentage_30d"`
	CirculatingSupply float64 `json:"circulating_supply"`
	TotalSupply       float64 `json:"total_supply"`
	MaxSupply         float64 `json:"max_supply"`
	ATH               float64 `json:"ath"`
	ATHDate           string  `json:"ath_date"`
	ATL               float64 `json:"atl"`
	ATLDate           string  `json:"atl_date"`
	LastUpdated       string  `json:"last_updated"`
}

// CoinGeckoOHLCVResponse represents OHLCV data from CoinGecko
type CoinGeckoOHLCVResponse [][]float64

// NewCoinGeckoProvider creates a new CoinGecko provider
func NewCoinGeckoProvider(config config.CoinGeckoConfig, httpClient *http.Client) *CoinGeckoProvider {
	return &CoinGeckoProvider{
		config:     config,
		httpClient: httpClient,
		isHealthy:  true,
	}
}

// GetPrice returns the current price for a specific symbol
func (p *CoinGeckoProvider) GetPrice(ctx context.Context, symbol string) (*models.Price, error) {
	url := fmt.Sprintf("%s/simple/price?ids=%s&vs_currencies=usd&include_24hr_vol=true&include_24hr_change=true",
		p.config.BaseURL, symbol)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if p.config.APIKey != "" {
		req.Header.Set("X-CG-Demo-API-Key", p.config.APIKey)
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

	var result map[string]map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	data, exists := result[symbol]
	if !exists {
		return nil, fmt.Errorf("symbol %s not found in response", symbol)
	}

	price := &models.Price{
		Symbol:    strings.ToUpper(symbol),
		Price:     decimal.NewFromFloat(data["usd"]),
		Volume24h: decimal.NewFromFloat(data["usd_24h_vol"]),
		Change24h: decimal.NewFromFloat(data["usd_24h_change"]),
		Timestamp: time.Now(),
		Source:    "coingecko",
	}

	p.isHealthy = true
	return price, nil
}

// GetPrices returns prices for multiple symbols
func (p *CoinGeckoProvider) GetPrices(ctx context.Context, symbols []string) ([]*models.Price, error) {
	if len(symbols) == 0 {
		return nil, fmt.Errorf("no symbols provided")
	}

	symbolsStr := strings.Join(symbols, ",")
	url := fmt.Sprintf("%s/coins/markets?vs_currency=usd&ids=%s&order=market_cap_desc&per_page=250&page=1&sparkline=false&price_change_percentage=24h,7d,30d",
		p.config.BaseURL, symbolsStr)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if p.config.APIKey != "" {
		req.Header.Set("X-CG-Demo-API-Key", p.config.APIKey)
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

	var result []CoinGeckoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var prices []*models.Price
	for _, coin := range result {
		price := &models.Price{
			Symbol:    strings.ToUpper(coin.Symbol),
			Price:     decimal.NewFromFloat(coin.CurrentPrice),
			Volume24h: decimal.NewFromFloat(coin.TotalVolume),
			Change24h: decimal.NewFromFloat(coin.PriceChangePercent24h),
			Timestamp: time.Now(),
			Source:    "coingecko",
		}
		prices = append(prices, price)
	}

	p.isHealthy = true
	return prices, nil
}

// GetPriceHistory returns historical price data
func (p *CoinGeckoProvider) GetPriceHistory(ctx context.Context, symbol, timeframe string, limit int) ([]*models.OHLCV, error) {
	// Convert timeframe to CoinGecko format
	days := p.timeframeToDays(timeframe, limit)
	
	url := fmt.Sprintf("%s/coins/%s/ohlc?vs_currency=usd&days=%d",
		p.config.BaseURL, symbol, days)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if p.config.APIKey != "" {
		req.Header.Set("X-CG-Demo-API-Key", p.config.APIKey)
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

	var result CoinGeckoOHLCVResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var ohlcv []*models.OHLCV
	for _, data := range result {
		if len(data) >= 5 {
			candle := &models.OHLCV{
				Symbol:    strings.ToUpper(symbol),
				Timeframe: timeframe,
				Open:      decimal.NewFromFloat(data[1]),
				High:      decimal.NewFromFloat(data[2]),
				Low:       decimal.NewFromFloat(data[3]),
				Close:     decimal.NewFromFloat(data[4]),
				Volume:    decimal.NewFromFloat(0), // CoinGecko OHLC doesn't include volume
				Timestamp: time.Unix(int64(data[0])/1000, 0),
			}
			ohlcv = append(ohlcv, candle)
		}
	}

	// Limit results
	if len(ohlcv) > limit {
		ohlcv = ohlcv[len(ohlcv)-limit:]
	}

	p.isHealthy = true
	return ohlcv, nil
}

// GetMarketData returns comprehensive market data for a symbol
func (p *CoinGeckoProvider) GetMarketData(ctx context.Context, symbol string) (*models.MarketData, error) {
	url := fmt.Sprintf("%s/coins/markets?vs_currency=usd&ids=%s&order=market_cap_desc&per_page=1&page=1&sparkline=false&price_change_percentage=24h,7d,30d",
		p.config.BaseURL, symbol)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if p.config.APIKey != "" {
		req.Header.Set("X-CG-Demo-API-Key", p.config.APIKey)
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

	var result []CoinGeckoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no data found for symbol %s", symbol)
	}

	coin := result[0]
	
	athDate, _ := time.Parse(time.RFC3339, coin.ATHDate)
	atlDate, _ := time.Parse(time.RFC3339, coin.ATLDate)
	lastUpdated, _ := time.Parse(time.RFC3339, coin.LastUpdated)

	marketData := &models.MarketData{
		Symbol:            strings.ToUpper(coin.Symbol),
		Name:              coin.Name,
		CurrentPrice:      decimal.NewFromFloat(coin.CurrentPrice),
		MarketCap:         decimal.NewFromFloat(coin.MarketCap),
		MarketCapRank:     coin.MarketCapRank,
		Volume24h:         decimal.NewFromFloat(coin.TotalVolume),
		Change24h:         decimal.NewFromFloat(coin.PriceChangePercent24h),
		Change7d:          decimal.NewFromFloat(coin.PriceChangePercent7d),
		Change30d:         decimal.NewFromFloat(coin.PriceChangePercent30d),
		High24h:           decimal.NewFromFloat(coin.High24h),
		Low24h:            decimal.NewFromFloat(coin.Low24h),
		CirculatingSupply: decimal.NewFromFloat(coin.CirculatingSupply),
		TotalSupply:       decimal.NewFromFloat(coin.TotalSupply),
		MaxSupply:         decimal.NewFromFloat(coin.MaxSupply),
		ATH:               decimal.NewFromFloat(coin.ATH),
		ATHDate:           athDate,
		ATL:               decimal.NewFromFloat(coin.ATL),
		ATLDate:           atlDate,
		LastUpdated:       lastUpdated,
	}

	p.isHealthy = true
	return marketData, nil
}

// IsHealthy returns the health status of the provider
func (p *CoinGeckoProvider) IsHealthy() bool {
	return p.isHealthy
}

// timeframeToDays converts timeframe to days for CoinGecko API
func (p *CoinGeckoProvider) timeframeToDays(timeframe string, limit int) int {
	switch timeframe {
	case "1m":
		return 1 // 1 day for minute data
	case "5m":
		return 1
	case "15m":
		return 1
	case "1h":
		return limit / 24
	case "4h":
		return limit / 6
	case "1d":
		return limit
	case "1w":
		return limit * 7
	default:
		return 30 // Default to 30 days
	}
}
