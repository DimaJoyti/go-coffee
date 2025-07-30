package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests
	
	handlers := NewEnhancedTradingHandlers(logger)
	
	v3 := router.Group("/api/v3")
	handlers.RegisterRoutes(v3)
	
	return router
}

func TestGetMarketDepth(t *testing.T) {
	router := setupTestRouter()

	tests := []struct {
		name           string
		symbol         string
		limit          string
		expectedStatus int
		expectedSymbol string
	}{
		{
			name:           "Valid request with default limit",
			symbol:         "BTCUSDT",
			limit:          "",
			expectedStatus: http.StatusOK,
			expectedSymbol: "BTCUSDT",
		},
		{
			name:           "Valid request with custom limit",
			symbol:         "ETHUSDT",
			limit:          "50",
			expectedStatus: http.StatusOK,
			expectedSymbol: "ETHUSDT",
		},
		{
			name:           "Invalid limit parameter",
			symbol:         "BTCUSDT",
			limit:          "invalid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/v3/market/depth/" + tt.symbol
			if tt.limit != "" {
				url += "?limit=" + tt.limit
			}

			req, err := http.NewRequest("GET", url, nil)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response models.MarketDepth
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				
				assert.Equal(t, tt.expectedSymbol, response.Symbol)
				assert.NotEmpty(t, response.Bids)
				assert.NotEmpty(t, response.Asks)
				assert.Greater(t, response.LastUpdateID, int64(0))
			}
		})
	}
}

func TestGetRecentTrades(t *testing.T) {
	router := setupTestRouter()

	tests := []struct {
		name           string
		symbol         string
		limit          string
		expectedStatus int
	}{
		{
			name:           "Valid request",
			symbol:         "BTCUSDT",
			limit:          "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Valid request with limit",
			symbol:         "ETHUSDT",
			limit:          "10",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid limit",
			symbol:         "BTCUSDT",
			limit:          "invalid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/v3/market/trades/" + tt.symbol
			if tt.limit != "" {
				url += "?limit=" + tt.limit
			}

			req, err := http.NewRequest("GET", url, nil)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				
				assert.Equal(t, tt.symbol, response["symbol"])
				assert.Contains(t, response, "trades")
			}
		})
	}
}

func TestGetTicker24hr(t *testing.T) {
	router := setupTestRouter()

	req, err := http.NewRequest("GET", "/api/v3/market/ticker/BTCUSDT", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Ticker24hr
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "BTCUSDT", response.Symbol)
	assert.Greater(t, response.LastPrice, float64(0))
	assert.Greater(t, response.Volume, float64(0))
	assert.Greater(t, response.CloseTime, int64(0))
}

func TestPlaceOrder(t *testing.T) {
	router := setupTestRouter()

	tests := []struct {
		name           string
		orderRequest   models.OrderRequest
		expectedStatus int
	}{
		{
			name: "Valid market order",
			orderRequest: models.OrderRequest{
				Symbol:   "BTCUSDT",
				Side:     "BUY",
				Type:     "MARKET",
				Quantity: 0.001,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Valid limit order",
			orderRequest: models.OrderRequest{
				Symbol:      "BTCUSDT",
				Side:        "SELL",
				Type:        "LIMIT",
				Quantity:    0.001,
				Price:       45000.00,
				TimeInForce: "GTC",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Invalid order - missing symbol",
			orderRequest: models.OrderRequest{
				Side:     "BUY",
				Type:     "MARKET",
				Quantity: 0.001,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid order - missing quantity",
			orderRequest: models.OrderRequest{
				Symbol: "BTCUSDT",
				Side:   "BUY",
				Type:   "MARKET",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.orderRequest)
			require.NoError(t, err)

			req, err := http.NewRequest("POST", "/api/v3/trading/order", bytes.NewBuffer(body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response models.TradingOrder
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				
				assert.Equal(t, tt.orderRequest.Symbol, response.Symbol)
				assert.Equal(t, tt.orderRequest.Side, response.Side)
				assert.Equal(t, tt.orderRequest.Type, response.Type)
				assert.NotEmpty(t, response.OrderID)
				assert.Equal(t, "NEW", response.Status)
			}
		})
	}
}

func TestGetOpenOrders(t *testing.T) {
	router := setupTestRouter()

	tests := []struct {
		name           string
		symbol         string
		expectedStatus int
	}{
		{
			name:           "Get all open orders",
			symbol:         "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get open orders for specific symbol",
			symbol:         "BTCUSDT",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/v3/trading/orders"
			if tt.symbol != "" {
				url += "?symbol=" + tt.symbol
			}

			req, err := http.NewRequest("GET", url, nil)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response []models.TradingOrder
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			
			// Should return array (even if empty)
			assert.IsType(t, []models.TradingOrder{}, response)
		})
	}
}

func TestCancelOrder(t *testing.T) {
	router := setupTestRouter()

	req, err := http.NewRequest("DELETE", "/api/v3/trading/order/12345678", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.CancelOrderResult
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "12345678", response.OrderID)
	assert.Equal(t, "CANCELED", response.Status)
	assert.Greater(t, response.TransactTime, int64(0))
}

func TestGetAccountBalance(t *testing.T) {
	router := setupTestRouter()

	req, err := http.NewRequest("GET", "/api/v3/account/balance", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "balances")
	balances := response["balances"].([]interface{})
	assert.Greater(t, len(balances), 0)
}

func TestGetMarketHeatmap(t *testing.T) {
	router := setupTestRouter()

	req, err := http.NewRequest("GET", "/api/v3/advanced/heatmap", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "data")
	data := response["data"].([]interface{})
	assert.Greater(t, len(data), 0)
}

// Benchmark tests for performance
func BenchmarkGetMarketDepth(b *testing.B) {
	router := setupTestRouter()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/api/v3/market/depth/BTCUSDT", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkPlaceOrder(b *testing.B) {
	router := setupTestRouter()
	
	orderRequest := models.OrderRequest{
		Symbol:   "BTCUSDT",
		Side:     "BUY",
		Type:     "MARKET",
		Quantity: 0.001,
	}
	
	body, _ := json.Marshal(orderRequest)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/api/v3/trading/order", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkGetTicker24hr(b *testing.B) {
	router := setupTestRouter()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/api/v3/market/ticker/BTCUSDT", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
