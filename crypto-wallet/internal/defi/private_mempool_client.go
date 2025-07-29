package defi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/ethereum/go-ethereum/core/types"
	"go.uber.org/zap"
)

// PrivateMempoolClient handles interactions with private mempool services
type PrivateMempoolClient struct {
	endpoint   string
	httpClient *http.Client
	logger     *logger.Logger
	
	// Configuration
	maxRetries     int
	retryDelay     time.Duration
	requestTimeout time.Duration
	apiKey         string
}

// PrivateMempoolRequest represents a request to private mempool
type PrivateMempoolRequest struct {
	Method string      `json:"method"`
	Params interface{} `json:"params"`
	ID     int         `json:"id"`
}

// PrivateMempoolResponse represents a response from private mempool
type PrivateMempoolResponse struct {
	Result interface{} `json:"result"`
	Error  *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
	ID int `json:"id"`
}

// TransactionSubmissionParams represents parameters for transaction submission
type TransactionSubmissionParams struct {
	SignedTransaction string                 `json:"signedTransaction"`
	Priority          string                 `json:"priority,omitempty"`
	PrivacyLevel      string                 `json:"privacyLevel,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// SubmissionResult represents the result of transaction submission
type SubmissionResult struct {
	TransactionHash string    `json:"transactionHash"`
	Status          string    `json:"status"`
	SubmittedAt     time.Time `json:"submittedAt"`
	EstimatedTime   int       `json:"estimatedTimeSeconds"`
	Priority        string    `json:"priority"`
}

// MempoolStatus represents the status of the private mempool
type MempoolStatus struct {
	IsHealthy        bool      `json:"isHealthy"`
	QueueSize        int       `json:"queueSize"`
	AverageWaitTime  int       `json:"averageWaitTimeSeconds"`
	SuccessRate      float64   `json:"successRate"`
	LastUpdate       time.Time `json:"lastUpdate"`
}

// NewPrivateMempoolClient creates a new private mempool client
func NewPrivateMempoolClient(endpoint string, logger *logger.Logger) *PrivateMempoolClient {
	return &PrivateMempoolClient{
		endpoint: endpoint,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger:         logger.Named("private-mempool-client"),
		maxRetries:     3,
		retryDelay:     1 * time.Second,
		requestTimeout: 30 * time.Second,
	}
}

// SetAPIKey sets the API key for authentication
func (pmc *PrivateMempoolClient) SetAPIKey(apiKey string) {
	pmc.apiKey = apiKey
}

// SubmitTransaction submits a transaction to the private mempool
func (pmc *PrivateMempoolClient) SubmitTransaction(ctx context.Context, tx *types.Transaction) error {
	pmc.logger.Info("Submitting transaction to private mempool",
		zap.String("hash", tx.Hash().Hex()),
		zap.Uint64("gas_price", tx.GasPrice().Uint64()),
		zap.Uint64("gas_limit", tx.Gas()))

	// Encode transaction
	txData, err := tx.MarshalBinary()
	if err != nil {
		return fmt.Errorf("failed to encode transaction: %w", err)
	}

	// Prepare submission parameters
	params := TransactionSubmissionParams{
		SignedTransaction: fmt.Sprintf("0x%x", txData),
		Priority:          pmc.calculatePriority(tx),
		PrivacyLevel:      "high",
		Metadata: map[string]interface{}{
			"source":    "go-coffee-platform",
			"timestamp": time.Now().Unix(),
			"gas_price": tx.GasPrice().String(),
			"gas_limit": tx.Gas(),
		},
	}

	// Submit transaction
	response, err := pmc.makeRequest(ctx, "eth_sendPrivateTransaction", params)
	if err != nil {
		pmc.logger.Error("Failed to submit transaction to private mempool", zap.Error(err))
		return fmt.Errorf("failed to submit transaction: %w", err)
	}

	// Parse submission result
	var result SubmissionResult
	if response.Result != nil {
		resultBytes, err := json.Marshal(response.Result)
		if err != nil {
			return fmt.Errorf("failed to marshal submission result: %w", err)
		}

		if err := json.Unmarshal(resultBytes, &result); err != nil {
			return fmt.Errorf("failed to unmarshal submission result: %w", err)
		}
	}

	pmc.logger.Info("Transaction submitted to private mempool successfully",
		zap.String("hash", tx.Hash().Hex()),
		zap.String("status", result.Status),
		zap.Int("estimated_time", result.EstimatedTime))

	return nil
}

// calculatePriority calculates transaction priority for private mempool
func (pmc *PrivateMempoolClient) calculatePriority(tx *types.Transaction) string {
	// High gas price = high priority
	gasPrice := tx.GasPrice().Uint64()
	
	if gasPrice > 50000000000 { // > 50 gwei
		return "urgent"
	} else if gasPrice > 20000000000 { // > 20 gwei
		return "high"
	} else if gasPrice > 10000000000 { // > 10 gwei
		return "medium"
	}
	
	return "low"
}

// GetTransactionStatus gets the status of a submitted transaction
func (pmc *PrivateMempoolClient) GetTransactionStatus(ctx context.Context, txHash string) (*TransactionStatus, error) {
	pmc.logger.Debug("Getting transaction status", zap.String("hash", txHash))

	params := map[string]interface{}{
		"transactionHash": txHash,
	}

	response, err := pmc.makeRequest(ctx, "eth_getPrivateTransactionStatus", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction status: %w", err)
	}

	// Parse status
	var status TransactionStatus
	if response.Result != nil {
		resultBytes, err := json.Marshal(response.Result)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal transaction status: %w", err)
		}

		if err := json.Unmarshal(resultBytes, &status); err != nil {
			return nil, fmt.Errorf("failed to unmarshal transaction status: %w", err)
		}
	}

	return &status, nil
}

// TransactionStatus represents the status of a transaction in private mempool
type TransactionStatus struct {
	Hash            string    `json:"hash"`
	Status          string    `json:"status"` // pending, submitted, confirmed, failed
	SubmittedAt     time.Time `json:"submittedAt"`
	ConfirmedAt     *time.Time `json:"confirmedAt,omitempty"`
	BlockNumber     *uint64   `json:"blockNumber,omitempty"`
	TransactionIndex *uint    `json:"transactionIndex,omitempty"`
	GasUsed         *uint64   `json:"gasUsed,omitempty"`
	EffectiveGasPrice *string `json:"effectiveGasPrice,omitempty"`
	Error           string    `json:"error,omitempty"`
}

// GetMempoolStatus gets the current status of the private mempool
func (pmc *PrivateMempoolClient) GetMempoolStatus(ctx context.Context) (*MempoolStatus, error) {
	pmc.logger.Debug("Getting private mempool status")

	response, err := pmc.makeRequest(ctx, "eth_getPrivateMempoolStatus", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get mempool status: %w", err)
	}

	// Parse status
	var status MempoolStatus
	if response.Result != nil {
		resultBytes, err := json.Marshal(response.Result)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal mempool status: %w", err)
		}

		if err := json.Unmarshal(resultBytes, &status); err != nil {
			return nil, fmt.Errorf("failed to unmarshal mempool status: %w", err)
		}
	}

	return &status, nil
}

// CancelTransaction attempts to cancel a pending transaction
func (pmc *PrivateMempoolClient) CancelTransaction(ctx context.Context, txHash string) error {
	pmc.logger.Info("Cancelling transaction", zap.String("hash", txHash))

	params := map[string]interface{}{
		"transactionHash": txHash,
	}

	response, err := pmc.makeRequest(ctx, "eth_cancelPrivateTransaction", params)
	if err != nil {
		return fmt.Errorf("failed to cancel transaction: %w", err)
	}

	// Check if cancellation was successful
	if response.Result != nil {
		if success, ok := response.Result.(bool); ok && !success {
			return fmt.Errorf("transaction cancellation failed")
		}
	}

	pmc.logger.Info("Transaction cancelled successfully", zap.String("hash", txHash))
	return nil
}

// EstimateSubmissionTime estimates how long it will take for a transaction to be included
func (pmc *PrivateMempoolClient) EstimateSubmissionTime(ctx context.Context, gasPrice uint64) (int, error) {
	pmc.logger.Debug("Estimating submission time", zap.Uint64("gas_price", gasPrice))

	params := map[string]interface{}{
		"gasPrice": fmt.Sprintf("0x%x", gasPrice),
	}

	response, err := pmc.makeRequest(ctx, "eth_estimatePrivateSubmissionTime", params)
	if err != nil {
		return 0, fmt.Errorf("failed to estimate submission time: %w", err)
	}

	// Parse estimation
	if response.Result != nil {
		if timeSeconds, ok := response.Result.(float64); ok {
			return int(timeSeconds), nil
		}
	}

	return 0, fmt.Errorf("invalid estimation response format")
}

// makeRequest makes an HTTP request to the private mempool service
func (pmc *PrivateMempoolClient) makeRequest(ctx context.Context, method string, params interface{}) (*PrivateMempoolResponse, error) {
	// Create request
	request := PrivateMempoolRequest{
		Method: method,
		Params: params,
		ID:     int(time.Now().UnixNano() % 1000000),
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request with timeout context
	reqCtx, cancel := context.WithTimeout(ctx, pmc.requestTimeout)
	defer cancel()

	var lastErr error
	for attempt := 0; attempt < pmc.maxRetries; attempt++ {
		if attempt > 0 {
			pmc.logger.Debug("Retrying private mempool request",
				zap.Int("attempt", attempt+1),
				zap.String("method", method))
			
			select {
			case <-time.After(pmc.retryDelay):
			case <-reqCtx.Done():
				return nil, reqCtx.Err()
			}
		}

		httpReq, err := http.NewRequestWithContext(reqCtx, "POST", pmc.endpoint, bytes.NewBuffer(requestBody))
		if err != nil {
			lastErr = fmt.Errorf("failed to create HTTP request: %w", err)
			continue
		}

		// Set headers
		httpReq.Header.Set("Content-Type", "application/json")
		if pmc.apiKey != "" {
			httpReq.Header.Set("Authorization", "Bearer "+pmc.apiKey)
		}

		// Make request
		resp, err := pmc.httpClient.Do(httpReq)
		if err != nil {
			lastErr = fmt.Errorf("HTTP request failed: %w", err)
			continue
		}

		// Read response
		responseBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			lastErr = fmt.Errorf("failed to read response body: %w", err)
			continue
		}

		// Check HTTP status
		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(responseBody))
			continue
		}

		// Parse response
		var privateMempoolResp PrivateMempoolResponse
		if err := json.Unmarshal(responseBody, &privateMempoolResp); err != nil {
			lastErr = fmt.Errorf("failed to unmarshal response: %w", err)
			continue
		}

		// Check for service error
		if privateMempoolResp.Error != nil {
			lastErr = fmt.Errorf("private mempool error %d: %s", 
				privateMempoolResp.Error.Code, privateMempoolResp.Error.Message)
			continue
		}

		pmc.logger.Debug("Private mempool request successful",
			zap.String("method", method),
			zap.Int("attempt", attempt+1))

		return &privateMempoolResp, nil
	}

	return nil, fmt.Errorf("all retry attempts failed, last error: %w", lastErr)
}

// IsHealthy checks if the private mempool service is healthy
func (pmc *PrivateMempoolClient) IsHealthy(ctx context.Context) bool {
	status, err := pmc.GetMempoolStatus(ctx)
	if err != nil {
		pmc.logger.Warn("Private mempool health check failed", zap.Error(err))
		return false
	}
	
	return status.IsHealthy
}

// GetRecommendedGasPrice gets recommended gas price from private mempool
func (pmc *PrivateMempoolClient) GetRecommendedGasPrice(ctx context.Context, priority string) (uint64, error) {
	params := map[string]interface{}{
		"priority": priority,
	}

	response, err := pmc.makeRequest(ctx, "eth_getRecommendedGasPrice", params)
	if err != nil {
		return 0, fmt.Errorf("failed to get recommended gas price: %w", err)
	}

	if response.Result != nil {
		if gasPriceHex, ok := response.Result.(string); ok {
			// Parse hex gas price
			var gasPrice uint64
			if _, err := fmt.Sscanf(gasPriceHex, "0x%x", &gasPrice); err != nil {
				return 0, fmt.Errorf("failed to parse gas price: %w", err)
			}
			return gasPrice, nil
		}
	}

	return 0, fmt.Errorf("invalid gas price response format")
}

// GetQueuePosition gets the position of a transaction in the private mempool queue
func (pmc *PrivateMempoolClient) GetQueuePosition(ctx context.Context, txHash string) (int, error) {
	params := map[string]interface{}{
		"transactionHash": txHash,
	}

	response, err := pmc.makeRequest(ctx, "eth_getQueuePosition", params)
	if err != nil {
		return 0, fmt.Errorf("failed to get queue position: %w", err)
	}

	if response.Result != nil {
		if position, ok := response.Result.(float64); ok {
			return int(position), nil
		}
	}

	return 0, fmt.Errorf("invalid queue position response format")
}
