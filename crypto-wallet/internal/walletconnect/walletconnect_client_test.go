package walletconnect

import (
	"context"
	"testing"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/config"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a test logger
func createTestLogger() *logger.Logger {
	logConfig := config.LoggingConfig{
		Level:  "info",
		Format: "json",
	}
	return logger.NewLogger(logConfig)
}

func TestNewWalletConnectClient(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultWalletConnectConfig()
	config.ProjectID = "test-project-id"

	client := NewWalletConnectClient(logger, config)

	assert.NotNil(t, client)
	assert.Equal(t, config.ProjectID, client.config.ProjectID)
	assert.Equal(t, config.SupportedChains, client.config.SupportedChains)
	assert.False(t, client.isRunning)
	assert.NotNil(t, client.sessions)
	assert.NotNil(t, client.connections)
	assert.NotNil(t, client.eventHandlers)
}

func TestWalletConnectClient_StartStop(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultWalletConnectConfig()
	config.ProjectID = "test-project-id"

	client := NewWalletConnectClient(logger, config)
	ctx := context.Background()

	err := client.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, client.IsRunning())

	err = client.Stop()
	assert.NoError(t, err)
	assert.False(t, client.IsRunning())
}

func TestWalletConnectClient_StartDisabled(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultWalletConnectConfig()
	config.Enabled = false

	client := NewWalletConnectClient(logger, config)
	ctx := context.Background()

	err := client.Start(ctx)
	assert.NoError(t, err)
	assert.False(t, client.IsRunning()) // Should remain false when disabled
}

func TestWalletConnectClient_CreateConnection(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultWalletConnectConfig()
	config.ProjectID = "test-project-id"

	client := NewWalletConnectClient(logger, config)

	sessionConfig := SessionConfig{
		Chains:  []string{"eip155:1", "eip155:137"},
		Methods: []string{"eth_sendTransaction", "personal_sign"},
		Events:  []string{"accountsChanged", "chainChanged"},
	}

	connection, err := client.CreateConnection(sessionConfig)
	assert.NoError(t, err)
	assert.NotNil(t, connection)
	assert.NotEmpty(t, connection.ID)
	assert.NotEmpty(t, connection.URI)
	assert.Equal(t, ConnectionStatusPending, connection.Status)
	assert.Equal(t, sessionConfig.Chains, connection.SessionConfig.Chains)
	assert.Equal(t, sessionConfig.Methods, connection.SessionConfig.Methods)

	// Test QR code generation
	if config.QRCodeConfig.Enabled {
		assert.NotEmpty(t, connection.QRCode)
	}

	// Test deep link generation
	if config.DeepLinkConfig.Enabled {
		assert.NotEmpty(t, connection.DeepLink)
	}
}

func TestWalletConnectClient_ApproveSession(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultWalletConnectConfig()
	config.ProjectID = "test-project-id"

	client := NewWalletConnectClient(logger, config)

	// Create a connection first
	sessionConfig := SessionConfig{
		Chains:  []string{"eip155:1"},
		Methods: []string{"eth_sendTransaction", "personal_sign"},
		Events:  []string{"accountsChanged"},
	}

	connection, err := client.CreateConnection(sessionConfig)
	require.NoError(t, err)

	// Approve the session
	accounts := []string{
		"0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E1",
		"0x8ba1f109551bD432803012645Hac136c5C1b5C8E1",
	}

	session, err := client.ApproveSession(connection.ProposalID, accounts)
	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.NotEmpty(t, session.ID)
	assert.Equal(t, SessionStatusActive, session.Status)
	assert.Equal(t, accounts, session.Accounts)
	assert.Equal(t, sessionConfig.Chains, session.Chains)
	assert.Equal(t, sessionConfig.Methods, session.Methods)
}

func TestWalletConnectClient_RejectSession(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultWalletConnectConfig()
	config.ProjectID = "test-project-id"

	client := NewWalletConnectClient(logger, config)

	// Create a connection first
	sessionConfig := SessionConfig{
		Chains:  []string{"eip155:1"},
		Methods: []string{"eth_sendTransaction"},
		Events:  []string{"accountsChanged"},
	}

	connection, err := client.CreateConnection(sessionConfig)
	require.NoError(t, err)

	// Reject the session
	err = client.RejectSession(connection.ProposalID, "User rejected")
	assert.NoError(t, err)

	// Check connection status
	updatedConnection, err := client.GetConnection(connection.ID)
	assert.NoError(t, err)
	assert.Equal(t, ConnectionStatusRejected, updatedConnection.Status)
}

func TestWalletConnectClient_SendTransactionRequest(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultWalletConnectConfig()
	config.ProjectID = "test-project-id"

	client := NewWalletConnectClient(logger, config)

	// Create and approve a session
	sessionConfig := SessionConfig{
		Chains:  []string{"eip155:1"},
		Methods: []string{"eth_sendTransaction"},
		Events:  []string{"accountsChanged"},
	}

	connection, err := client.CreateConnection(sessionConfig)
	require.NoError(t, err)

	accounts := []string{"0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E1"}
	session, err := client.ApproveSession(connection.ProposalID, accounts)
	require.NoError(t, err)

	// Send transaction request
	txRequest := &TransactionRequest{
		ID:     "tx-001",
		Method: "eth_sendTransaction",
		Chain:  "eip155:1",
		From:   common.HexToAddress(accounts[0]),
		To:     &[]common.Address{common.HexToAddress("0x8ba1f109551bD432803012645Hac136c5C1b5C8E1")}[0],
		Value:  &[]decimal.Decimal{decimal.NewFromFloat(0.1)}[0], // 0.1 ETH
		Params: map[string]interface{}{
			"to":    "0x8ba1f109551bD432803012645Hac136c5C1b5C8E1",
			"value": "0x16345785d8a0000", // 0.1 ETH in hex
		},
		Metadata: map[string]interface{}{"type": "transfer"},
	}

	response, err := client.SendTransactionRequest(session.ID, txRequest)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, txRequest.ID, response.ID)
	assert.True(t, response.Success)
	assert.NotEmpty(t, response.TransactionHash)
}

func TestWalletConnectClient_SendSignRequest(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultWalletConnectConfig()
	config.ProjectID = "test-project-id"

	client := NewWalletConnectClient(logger, config)

	// Create and approve a session
	sessionConfig := SessionConfig{
		Chains:  []string{"eip155:1"},
		Methods: []string{"personal_sign"},
		Events:  []string{"accountsChanged"},
	}

	connection, err := client.CreateConnection(sessionConfig)
	require.NoError(t, err)

	accounts := []string{"0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E1"}
	session, err := client.ApproveSession(connection.ProposalID, accounts)
	require.NoError(t, err)

	// Send sign request
	signRequest := &SignRequest{
		ID:      "sign-001",
		Method:  "personal_sign",
		Address: common.HexToAddress(accounts[0]),
		Message: "Hello, WalletConnect!",
		Type:    SignTypePersonalSign,
		Metadata: map[string]interface{}{"type": "message_sign"},
	}

	response, err := client.SendSignRequest(session.ID, signRequest)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, signRequest.ID, response.ID)
	assert.True(t, response.Success)
	assert.NotEmpty(t, response.Signature)
}

func TestWalletConnectClient_UpdateSession(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultWalletConnectConfig()
	config.ProjectID = "test-project-id"

	client := NewWalletConnectClient(logger, config)

	// Create and approve a session
	sessionConfig := SessionConfig{
		Chains:  []string{"eip155:1"},
		Methods: []string{"eth_sendTransaction"},
		Events:  []string{"accountsChanged"},
	}

	connection, err := client.CreateConnection(sessionConfig)
	require.NoError(t, err)

	accounts := []string{"0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E1"}
	session, err := client.ApproveSession(connection.ProposalID, accounts)
	require.NoError(t, err)

	// Update session with new accounts and chains
	newAccounts := []string{
		"0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E1",
		"0x8ba1f109551bD432803012645Hac136c5C1b5C8E1",
	}
	newChains := []string{"eip155:1", "eip155:137"}

	err = client.UpdateSession(session.ID, newAccounts, newChains)
	assert.NoError(t, err)

	// Verify updates
	updatedSession, err := client.GetSession(session.ID)
	assert.NoError(t, err)
	assert.Equal(t, newAccounts, updatedSession.Accounts)
	assert.Equal(t, newChains, updatedSession.Chains)
}

func TestWalletConnectClient_DisconnectSession(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultWalletConnectConfig()
	config.ProjectID = "test-project-id"

	client := NewWalletConnectClient(logger, config)

	// Create and approve a session
	sessionConfig := SessionConfig{
		Chains:  []string{"eip155:1"},
		Methods: []string{"eth_sendTransaction"},
		Events:  []string{"accountsChanged"},
	}

	connection, err := client.CreateConnection(sessionConfig)
	require.NoError(t, err)

	accounts := []string{"0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E1"}
	session, err := client.ApproveSession(connection.ProposalID, accounts)
	require.NoError(t, err)

	// Disconnect session
	err = client.DisconnectSession(session.ID, "User disconnected")
	assert.NoError(t, err)

	// Verify disconnection
	disconnectedSession, err := client.GetSession(session.ID)
	assert.NoError(t, err)
	assert.Equal(t, SessionStatusDisconnected, disconnectedSession.Status)
}

func TestWalletConnectClient_EventHandlers(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultWalletConnectConfig()
	config.ProjectID = "test-project-id"

	client := NewWalletConnectClient(logger, config)

	// Add event handler
	eventReceived := false
	handler := func(event *Event) error {
		eventReceived = true
		assert.Equal(t, EventTypeSessionProposal, event.Type)
		return nil
	}

	client.AddEventHandler(EventTypeSessionProposal, handler)

	// Create a connection to trigger an event
	sessionConfig := SessionConfig{
		Chains:  []string{"eip155:1"},
		Methods: []string{"eth_sendTransaction"},
		Events:  []string{"accountsChanged"},
	}

	connection, err := client.CreateConnection(sessionConfig)
	require.NoError(t, err)

	// Approve session to trigger event
	accounts := []string{"0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E1"}
	_, err = client.ApproveSession(connection.ProposalID, accounts)
	assert.NoError(t, err)

	// Give some time for event processing
	time.Sleep(100 * time.Millisecond)
	assert.True(t, eventReceived)
}

func TestWalletConnectClient_GetMetrics(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultWalletConnectConfig()
	config.ProjectID = "test-project-id"

	client := NewWalletConnectClient(logger, config)

	// Get initial metrics
	metrics := client.GetMetrics()
	assert.NotNil(t, metrics)
	assert.Equal(t, 0, metrics["total_sessions"])
	assert.Equal(t, 0, metrics["total_connections"])
	assert.Equal(t, false, metrics["is_running"])

	// Create a connection and session
	sessionConfig := SessionConfig{
		Chains:  []string{"eip155:1"},
		Methods: []string{"eth_sendTransaction"},
		Events:  []string{"accountsChanged"},
	}

	connection, err := client.CreateConnection(sessionConfig)
	require.NoError(t, err)

	accounts := []string{"0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E1"}
	_, err = client.ApproveSession(connection.ProposalID, accounts)
	require.NoError(t, err)

	// Get updated metrics
	updatedMetrics := client.GetMetrics()
	assert.Equal(t, 1, updatedMetrics["total_sessions"])
	assert.Equal(t, 1, updatedMetrics["total_connections"])
	assert.Equal(t, 1, client.GetSessionCount())
	assert.Equal(t, 0, client.GetConnectionCount()) // Connection becomes connected after approval
}

func TestGetDefaultWalletConnectConfig(t *testing.T) {
	config := GetDefaultWalletConnectConfig()

	assert.True(t, config.Enabled)
	assert.Equal(t, "wss://relay.walletconnect.com", config.RelayURL)
	assert.Equal(t, "Crypto Wallet", config.Metadata.Name)
	assert.Contains(t, config.SupportedChains, "eip155:1")
	assert.Contains(t, config.SupportedMethods, "eth_sendTransaction")
	assert.Contains(t, config.SupportedEvents, "accountsChanged")
	assert.Equal(t, 24*time.Hour, config.SessionTimeout)
	assert.Equal(t, 5*time.Minute, config.ConnectionTimeout)
	assert.True(t, config.QRCodeConfig.Enabled)
	assert.True(t, config.DeepLinkConfig.Enabled)
	assert.True(t, config.SecurityConfig.RequireApproval)
	assert.Equal(t, "info", config.LoggingConfig.LogLevel)
}
