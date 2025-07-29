package walletconnect

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// WalletConnectClient provides WalletConnect v2 integration
type WalletConnectClient struct {
	logger *logger.Logger
	config WalletConnectConfig

	// Session management
	sessions     map[string]*Session
	sessionMutex sync.RWMutex

	// Connection management
	connections     map[string]*Connection
	connectionMutex sync.RWMutex

	// Event handlers
	eventHandlers map[EventType][]EventHandler
	handlerMutex  sync.RWMutex

	// State management
	isRunning bool
	stopChan  chan struct{}
	mutex     sync.RWMutex
}

// WalletConnectConfig holds configuration for WalletConnect
type WalletConnectConfig struct {
	Enabled           bool           `json:"enabled" yaml:"enabled"`
	ProjectID         string         `json:"project_id" yaml:"project_id"`
	RelayURL          string         `json:"relay_url" yaml:"relay_url"`
	Metadata          AppMetadata    `json:"metadata" yaml:"metadata"`
	SupportedChains   []string       `json:"supported_chains" yaml:"supported_chains"`
	SupportedMethods  []string       `json:"supported_methods" yaml:"supported_methods"`
	SupportedEvents   []string       `json:"supported_events" yaml:"supported_events"`
	SessionTimeout    time.Duration  `json:"session_timeout" yaml:"session_timeout"`
	ConnectionTimeout time.Duration  `json:"connection_timeout" yaml:"connection_timeout"`
	QRCodeConfig      QRCodeConfig   `json:"qr_code_config" yaml:"qr_code_config"`
	DeepLinkConfig    DeepLinkConfig `json:"deep_link_config" yaml:"deep_link_config"`
	SecurityConfig    SecurityConfig `json:"security_config" yaml:"security_config"`
	LoggingConfig     LoggingConfig  `json:"logging_config" yaml:"logging_config"`
}

// AppMetadata describes the dApp metadata
type AppMetadata struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	URL         string   `json:"url"`
	Icons       []string `json:"icons"`
	VerifyURL   string   `json:"verifyUrl,omitempty"`
}

// QRCodeConfig holds QR code generation settings
type QRCodeConfig struct {
	Enabled    bool   `json:"enabled"`
	Size       int    `json:"size"`
	ErrorLevel string `json:"error_level"`
	Border     int    `json:"border"`
	Format     string `json:"format"` // PNG, SVG, etc.
}

// DeepLinkConfig holds deep link settings
type DeepLinkConfig struct {
	Enabled        bool   `json:"enabled"`
	CustomScheme   string `json:"custom_scheme"`
	UniversalLink  string `json:"universal_link"`
	AndroidPackage string `json:"android_package"`
	IOSBundleID    string `json:"ios_bundle_id"`
}

// SecurityConfig holds security settings
type SecurityConfig struct {
	RequireApproval    bool          `json:"require_approval"`
	SessionEncryption  bool          `json:"session_encryption"`
	MessageSigning     bool          `json:"message_signing"`
	TransactionSigning bool          `json:"transaction_signing"`
	MaxSessionDuration time.Duration `json:"max_session_duration"`
	AllowedOrigins     []string      `json:"allowed_origins"`
	BlockedAddresses   []string      `json:"blocked_addresses"`
}

// LoggingConfig holds logging settings
type LoggingConfig struct {
	LogLevel        string `json:"log_level"`
	LogConnections  bool   `json:"log_connections"`
	LogTransactions bool   `json:"log_transactions"`
	LogEvents       bool   `json:"log_events"`
}

// Session represents a WalletConnect session
type Session struct {
	ID          string                 `json:"id"`
	Topic       string                 `json:"topic"`
	PeerID      string                 `json:"peer_id"`
	Accounts    []string               `json:"accounts"`
	Chains      []string               `json:"chains"`
	Methods     []string               `json:"methods"`
	Events      []string               `json:"events"`
	Metadata    AppMetadata            `json:"metadata"`
	Expiry      time.Time              `json:"expiry"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Status      SessionStatus          `json:"status"`
	Permissions map[string]interface{} `json:"permissions"`
}

// Connection represents a WalletConnect connection
type Connection struct {
	ID            string           `json:"id"`
	URI           string           `json:"uri"`
	QRCode        string           `json:"qr_code,omitempty"`
	DeepLink      string           `json:"deep_link,omitempty"`
	Status        ConnectionStatus `json:"status"`
	CreatedAt     time.Time        `json:"created_at"`
	ExpiresAt     time.Time        `json:"expires_at"`
	Metadata      AppMetadata      `json:"metadata"`
	ProposalID    string           `json:"proposal_id"`
	SessionConfig SessionConfig    `json:"session_config"`
}

// SessionConfig defines session configuration
type SessionConfig struct {
	Chains          []string `json:"chains"`
	Methods         []string `json:"methods"`
	Events          []string `json:"events"`
	OptionalChains  []string `json:"optionalChains,omitempty"`
	OptionalMethods []string `json:"optionalMethods,omitempty"`
	OptionalEvents  []string `json:"optionalEvents,omitempty"`
}

// TransactionRequest represents a transaction request
type TransactionRequest struct {
	ID       string                 `json:"id"`
	Method   string                 `json:"method"`
	Params   map[string]interface{} `json:"params"`
	Chain    string                 `json:"chain"`
	From     common.Address         `json:"from"`
	To       *common.Address        `json:"to,omitempty"`
	Value    *decimal.Decimal       `json:"value,omitempty"`
	Data     string                 `json:"data,omitempty"`
	Gas      *decimal.Decimal       `json:"gas,omitempty"`
	GasPrice *decimal.Decimal       `json:"gasPrice,omitempty"`
	Nonce    *uint64                `json:"nonce,omitempty"`
	Metadata map[string]interface{} `json:"metadata"`
}

// TransactionResponse represents a transaction response
type TransactionResponse struct {
	ID              string                 `json:"id"`
	TransactionHash string                 `json:"transaction_hash,omitempty"`
	Success         bool                   `json:"success"`
	Error           string                 `json:"error,omitempty"`
	Signature       string                 `json:"signature,omitempty"`
	Receipt         map[string]interface{} `json:"receipt,omitempty"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// SignRequest represents a message signing request
type SignRequest struct {
	ID       string                 `json:"id"`
	Method   string                 `json:"method"`
	Address  common.Address         `json:"address"`
	Message  string                 `json:"message"`
	Type     SignType               `json:"type"`
	Metadata map[string]interface{} `json:"metadata"`
}

// SignResponse represents a message signing response
type SignResponse struct {
	ID        string                 `json:"id"`
	Signature string                 `json:"signature,omitempty"`
	Success   bool                   `json:"success"`
	Error     string                 `json:"error,omitempty"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// Event represents a WalletConnect event
type Event struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	SessionID string                 `json:"session_id,omitempty"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// EventHandler defines an event handler function
type EventHandler func(event *Event) error

// Enums
type SessionStatus int

const (
	SessionStatusPending SessionStatus = iota
	SessionStatusActive
	SessionStatusExpired
	SessionStatusDisconnected
	SessionStatusRejected
)

type ConnectionStatus int

const (
	ConnectionStatusPending ConnectionStatus = iota
	ConnectionStatusConnecting
	ConnectionStatusConnected
	ConnectionStatusDisconnected
	ConnectionStatusExpired
	ConnectionStatusRejected
)

type EventType int

const (
	EventTypeSessionProposal EventType = iota
	EventTypeSessionUpdate
	EventTypeSessionDelete
	EventTypeSessionRequest
	EventTypeSessionResponse
	EventTypeConnectionUpdate
	EventTypeAccountsChanged
	EventTypeChainChanged
	EventTypeDisconnect
	EventTypeError
)

type SignType int

const (
	SignTypePersonalSign SignType = iota
	SignTypeEthSign
	SignTypeTypedData
	SignTypeTypedDataV3
	SignTypeTypedDataV4
)

// NewWalletConnectClient creates a new WalletConnect client
func NewWalletConnectClient(logger *logger.Logger, config WalletConnectConfig) *WalletConnectClient {
	return &WalletConnectClient{
		logger:        logger.Named("walletconnect-client"),
		config:        config,
		sessions:      make(map[string]*Session),
		connections:   make(map[string]*Connection),
		eventHandlers: make(map[EventType][]EventHandler),
		stopChan:      make(chan struct{}),
	}
}

// Start starts the WalletConnect client
func (wc *WalletConnectClient) Start(ctx context.Context) error {
	wc.mutex.Lock()
	defer wc.mutex.Unlock()

	if wc.isRunning {
		return fmt.Errorf("WalletConnect client is already running")
	}

	if !wc.config.Enabled {
		wc.logger.Info("WalletConnect client is disabled")
		return nil
	}

	wc.logger.Info("Starting WalletConnect client",
		zap.String("project_id", wc.config.ProjectID),
		zap.String("relay_url", wc.config.RelayURL),
		zap.Strings("supported_chains", wc.config.SupportedChains),
		zap.Strings("supported_methods", wc.config.SupportedMethods))

	// Initialize WalletConnect core
	if err := wc.initializeCore(); err != nil {
		return fmt.Errorf("failed to initialize WalletConnect core: %w", err)
	}

	// Start event processing
	go wc.processEvents(ctx)

	// Start session cleanup
	go wc.cleanupExpiredSessions(ctx)

	wc.isRunning = true
	wc.logger.Info("WalletConnect client started successfully")
	return nil
}

// Stop stops the WalletConnect client
func (wc *WalletConnectClient) Stop() error {
	wc.mutex.Lock()
	defer wc.mutex.Unlock()

	if !wc.isRunning {
		return nil
	}

	wc.logger.Info("Stopping WalletConnect client")

	// Disconnect all active sessions
	wc.sessionMutex.Lock()
	for _, session := range wc.sessions {
		if session.Status == SessionStatusActive {
			wc.disconnectSessionInternal(session.ID)
		}
	}
	wc.sessionMutex.Unlock()

	wc.isRunning = false
	close(wc.stopChan)

	wc.logger.Info("WalletConnect client stopped")
	return nil
}

// Core functionality methods

// CreateConnection creates a new WalletConnect connection
func (wc *WalletConnectClient) CreateConnection(sessionConfig SessionConfig) (*Connection, error) {
	wc.logger.Info("Creating WalletConnect connection",
		zap.Strings("chains", sessionConfig.Chains),
		zap.Strings("methods", sessionConfig.Methods))

	// Generate connection ID and URI
	connectionID := wc.generateConnectionID()
	uri := wc.generateConnectionURI(connectionID, sessionConfig)

	// Generate QR code if enabled
	var qrCode string
	if wc.config.QRCodeConfig.Enabled {
		qrCode = wc.generateQRCode(uri)
	}

	// Generate deep link if enabled
	var deepLink string
	if wc.config.DeepLinkConfig.Enabled {
		deepLink = wc.generateDeepLink(uri)
	}

	connection := &Connection{
		ID:            connectionID,
		URI:           uri,
		QRCode:        qrCode,
		DeepLink:      deepLink,
		Status:        ConnectionStatusPending,
		CreatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(wc.config.ConnectionTimeout),
		Metadata:      wc.config.Metadata,
		ProposalID:    wc.generateProposalID(),
		SessionConfig: sessionConfig,
	}

	// Store connection
	wc.connectionMutex.Lock()
	wc.connections[connectionID] = connection
	wc.connectionMutex.Unlock()

	wc.logger.Info("WalletConnect connection created",
		zap.String("connection_id", connectionID),
		zap.String("uri", uri))

	return connection, nil
}

// ApproveSession approves a session proposal
func (wc *WalletConnectClient) ApproveSession(proposalID string, accounts []string) (*Session, error) {
	wc.logger.Info("Approving session proposal",
		zap.String("proposal_id", proposalID),
		zap.Strings("accounts", accounts))

	// Find connection by proposal ID
	var connection *Connection
	wc.connectionMutex.RLock()
	for _, conn := range wc.connections {
		if conn.ProposalID == proposalID {
			connection = conn
			break
		}
	}
	wc.connectionMutex.RUnlock()

	if connection == nil {
		return nil, fmt.Errorf("connection not found for proposal ID: %s", proposalID)
	}

	// Create session
	sessionID := wc.generateSessionID()
	session := &Session{
		ID:        sessionID,
		Topic:     wc.generateTopic(),
		PeerID:    wc.generatePeerID(),
		Accounts:  accounts,
		Chains:    connection.SessionConfig.Chains,
		Methods:   connection.SessionConfig.Methods,
		Events:    connection.SessionConfig.Events,
		Metadata:  connection.Metadata,
		Expiry:    time.Now().Add(wc.config.SessionTimeout),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Status:    SessionStatusActive,
		Permissions: map[string]interface{}{
			"chains":  connection.SessionConfig.Chains,
			"methods": connection.SessionConfig.Methods,
			"events":  connection.SessionConfig.Events,
		},
	}

	// Store session
	wc.sessionMutex.Lock()
	wc.sessions[sessionID] = session
	wc.sessionMutex.Unlock()

	// Update connection status
	wc.connectionMutex.Lock()
	connection.Status = ConnectionStatusConnected
	wc.connectionMutex.Unlock()

	// Emit session approval event
	wc.emitEvent(&Event{
		ID:        wc.generateEventID(),
		Type:      EventTypeSessionProposal,
		SessionID: sessionID,
		Data: map[string]interface{}{
			"session":    session,
			"connection": connection,
			"approved":   true,
		},
		Timestamp: time.Now(),
	})

	wc.logger.Info("Session approved successfully",
		zap.String("session_id", sessionID),
		zap.String("topic", session.Topic))

	return session, nil
}

// RejectSession rejects a session proposal
func (wc *WalletConnectClient) RejectSession(proposalID string, reason string) error {
	wc.logger.Info("Rejecting session proposal",
		zap.String("proposal_id", proposalID),
		zap.String("reason", reason))

	// Find and update connection
	wc.connectionMutex.Lock()
	for _, connection := range wc.connections {
		if connection.ProposalID == proposalID {
			connection.Status = ConnectionStatusRejected
			break
		}
	}
	wc.connectionMutex.Unlock()

	// Emit rejection event
	wc.emitEvent(&Event{
		ID:   wc.generateEventID(),
		Type: EventTypeSessionProposal,
		Data: map[string]interface{}{
			"proposal_id": proposalID,
			"approved":    false,
			"reason":      reason,
		},
		Timestamp: time.Now(),
	})

	wc.logger.Info("Session proposal rejected",
		zap.String("proposal_id", proposalID))

	return nil
}

// SendTransactionRequest sends a transaction request to the wallet
func (wc *WalletConnectClient) SendTransactionRequest(sessionID string, request *TransactionRequest) (*TransactionResponse, error) {
	wc.logger.Info("Sending transaction request",
		zap.String("session_id", sessionID),
		zap.String("method", request.Method),
		zap.String("from", request.From.Hex()))

	// Get session
	session, err := wc.GetSession(sessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if session.Status != SessionStatusActive {
		return nil, fmt.Errorf("session is not active")
	}

	// Validate request
	if err := wc.validateTransactionRequest(session, request); err != nil {
		return nil, fmt.Errorf("invalid transaction request: %w", err)
	}

	// Send request (mock implementation)
	response := &TransactionResponse{
		ID:              request.ID,
		TransactionHash: wc.generateTransactionHash(),
		Success:         true,
		Metadata:        request.Metadata,
	}

	// Emit transaction event
	wc.emitEvent(&Event{
		ID:        wc.generateEventID(),
		Type:      EventTypeSessionRequest,
		SessionID: sessionID,
		Data: map[string]interface{}{
			"request":  request,
			"response": response,
			"type":     "transaction",
		},
		Timestamp: time.Now(),
	})

	wc.logger.Info("Transaction request sent successfully",
		zap.String("session_id", sessionID),
		zap.String("tx_hash", response.TransactionHash))

	return response, nil
}

// SendSignRequest sends a message signing request to the wallet
func (wc *WalletConnectClient) SendSignRequest(sessionID string, request *SignRequest) (*SignResponse, error) {
	wc.logger.Info("Sending sign request",
		zap.String("session_id", sessionID),
		zap.String("method", request.Method),
		zap.String("address", request.Address.Hex()))

	// Get session
	session, err := wc.GetSession(sessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if session.Status != SessionStatusActive {
		return nil, fmt.Errorf("session is not active")
	}

	// Validate request
	if err := wc.validateSignRequest(session, request); err != nil {
		return nil, fmt.Errorf("invalid sign request: %w", err)
	}

	// Send request (mock implementation)
	response := &SignResponse{
		ID:        request.ID,
		Signature: wc.generateSignature(),
		Success:   true,
		Metadata:  request.Metadata,
	}

	// Emit sign event
	wc.emitEvent(&Event{
		ID:        wc.generateEventID(),
		Type:      EventTypeSessionRequest,
		SessionID: sessionID,
		Data: map[string]interface{}{
			"request":  request,
			"response": response,
			"type":     "sign",
		},
		Timestamp: time.Now(),
	})

	wc.logger.Info("Sign request sent successfully",
		zap.String("session_id", sessionID),
		zap.String("signature", response.Signature))

	return response, nil
}

// DisconnectSession disconnects a session
func (wc *WalletConnectClient) DisconnectSession(sessionID string, reason string) error {
	wc.logger.Info("Disconnecting session",
		zap.String("session_id", sessionID),
		zap.String("reason", reason))

	return wc.disconnectSessionInternal(sessionID)
}

// UpdateSession updates session accounts or chains
func (wc *WalletConnectClient) UpdateSession(sessionID string, accounts []string, chains []string) error {
	wc.logger.Info("Updating session",
		zap.String("session_id", sessionID),
		zap.Strings("accounts", accounts),
		zap.Strings("chains", chains))

	wc.sessionMutex.Lock()
	defer wc.sessionMutex.Unlock()

	session, exists := wc.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	// Update session
	if accounts != nil {
		session.Accounts = accounts
	}
	if chains != nil {
		session.Chains = chains
	}
	session.UpdatedAt = time.Now()

	// Emit update event
	wc.emitEvent(&Event{
		ID:        wc.generateEventID(),
		Type:      EventTypeSessionUpdate,
		SessionID: sessionID,
		Data: map[string]interface{}{
			"session":  session,
			"accounts": accounts,
			"chains":   chains,
		},
		Timestamp: time.Now(),
	})

	wc.logger.Info("Session updated successfully",
		zap.String("session_id", sessionID))

	return nil
}

// GetSession returns a session by ID
func (wc *WalletConnectClient) GetSession(sessionID string) (*Session, error) {
	wc.sessionMutex.RLock()
	defer wc.sessionMutex.RUnlock()

	session, exists := wc.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	return session, nil
}

// GetActiveSessions returns all active sessions
func (wc *WalletConnectClient) GetActiveSessions() []*Session {
	wc.sessionMutex.RLock()
	defer wc.sessionMutex.RUnlock()

	var activeSessions []*Session
	for _, session := range wc.sessions {
		if session.Status == SessionStatusActive {
			activeSessions = append(activeSessions, session)
		}
	}

	return activeSessions
}

// GetConnection returns a connection by ID
func (wc *WalletConnectClient) GetConnection(connectionID string) (*Connection, error) {
	wc.connectionMutex.RLock()
	defer wc.connectionMutex.RUnlock()

	connection, exists := wc.connections[connectionID]
	if !exists {
		return nil, fmt.Errorf("connection not found: %s", connectionID)
	}

	return connection, nil
}

// GetPendingConnections returns all pending connections
func (wc *WalletConnectClient) GetPendingConnections() []*Connection {
	wc.connectionMutex.RLock()
	defer wc.connectionMutex.RUnlock()

	var pendingConnections []*Connection
	for _, connection := range wc.connections {
		if connection.Status == ConnectionStatusPending {
			pendingConnections = append(pendingConnections, connection)
		}
	}

	return pendingConnections
}

// Event management methods

// AddEventHandler adds an event handler for a specific event type
func (wc *WalletConnectClient) AddEventHandler(eventType EventType, handler EventHandler) {
	wc.handlerMutex.Lock()
	defer wc.handlerMutex.Unlock()

	wc.eventHandlers[eventType] = append(wc.eventHandlers[eventType], handler)
}

// RemoveEventHandler removes an event handler (simplified implementation)
func (wc *WalletConnectClient) RemoveEventHandler(eventType EventType) {
	wc.handlerMutex.Lock()
	defer wc.handlerMutex.Unlock()

	delete(wc.eventHandlers, eventType)
}

// Helper methods

// initializeCore initializes the WalletConnect core
func (wc *WalletConnectClient) initializeCore() error {
	wc.logger.Info("Initializing WalletConnect core")

	// Mock initialization - in production would initialize actual WalletConnect SDK
	if wc.config.ProjectID == "" {
		return fmt.Errorf("project ID is required")
	}

	if wc.config.RelayURL == "" {
		wc.config.RelayURL = "wss://relay.walletconnect.com"
	}

	wc.logger.Info("WalletConnect core initialized successfully")
	return nil
}

// processEvents processes WalletConnect events
func (wc *WalletConnectClient) processEvents(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-wc.stopChan:
			return
		case <-ticker.C:
			// Mock event processing - in production would handle real events
			wc.logger.Debug("Processing WalletConnect events")
		}
	}
}

// cleanupExpiredSessions cleans up expired sessions
func (wc *WalletConnectClient) cleanupExpiredSessions(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-wc.stopChan:
			return
		case <-ticker.C:
			wc.performSessionCleanup()
		}
	}
}

// performSessionCleanup performs the actual session cleanup
func (wc *WalletConnectClient) performSessionCleanup() {
	now := time.Now()

	wc.sessionMutex.Lock()
	for sessionID, session := range wc.sessions {
		if now.After(session.Expiry) {
			session.Status = SessionStatusExpired
			wc.logger.Info("Session expired", zap.String("session_id", sessionID))

			// Emit expiry event
			wc.emitEvent(&Event{
				ID:        wc.generateEventID(),
				Type:      EventTypeSessionDelete,
				SessionID: sessionID,
				Data: map[string]interface{}{
					"session": session,
					"reason":  "expired",
				},
				Timestamp: now,
			})
		}
	}
	wc.sessionMutex.Unlock()

	// Cleanup expired connections
	wc.connectionMutex.Lock()
	for connectionID, connection := range wc.connections {
		if now.After(connection.ExpiresAt) {
			connection.Status = ConnectionStatusExpired
			wc.logger.Info("Connection expired", zap.String("connection_id", connectionID))
		}
	}
	wc.connectionMutex.Unlock()
}

// disconnectSessionInternal disconnects a session internally
func (wc *WalletConnectClient) disconnectSessionInternal(sessionID string) error {
	wc.sessionMutex.Lock()
	defer wc.sessionMutex.Unlock()

	session, exists := wc.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	session.Status = SessionStatusDisconnected
	session.UpdatedAt = time.Now()

	// Emit disconnect event
	wc.emitEvent(&Event{
		ID:        wc.generateEventID(),
		Type:      EventTypeDisconnect,
		SessionID: sessionID,
		Data: map[string]interface{}{
			"session": session,
		},
		Timestamp: time.Now(),
	})

	wc.logger.Info("Session disconnected", zap.String("session_id", sessionID))
	return nil
}

// emitEvent emits an event to all registered handlers
func (wc *WalletConnectClient) emitEvent(event *Event) {
	wc.handlerMutex.RLock()
	handlers := wc.eventHandlers[event.Type]
	wc.handlerMutex.RUnlock()

	for _, handler := range handlers {
		go func(h EventHandler) {
			if err := h(event); err != nil {
				wc.logger.Error("Event handler error",
					zap.String("event_type", wc.getEventTypeString(event.Type)),
					zap.Error(err))
			}
		}(handler)
	}

	if wc.config.LoggingConfig.LogEvents {
		wc.logger.Debug("Event emitted",
			zap.String("event_id", event.ID),
			zap.String("event_type", wc.getEventTypeString(event.Type)),
			zap.String("session_id", event.SessionID))
	}
}

// Validation methods

// validateTransactionRequest validates a transaction request
func (wc *WalletConnectClient) validateTransactionRequest(session *Session, request *TransactionRequest) error {
	// Check if method is supported
	methodSupported := false
	for _, method := range session.Methods {
		if method == request.Method {
			methodSupported = true
			break
		}
	}
	if !methodSupported {
		return fmt.Errorf("method %s not supported in session", request.Method)
	}

	// Check if chain is supported
	chainSupported := false
	for _, chain := range session.Chains {
		if chain == request.Chain {
			chainSupported = true
			break
		}
	}
	if !chainSupported {
		return fmt.Errorf("chain %s not supported in session", request.Chain)
	}

	// Check if account is in session
	accountSupported := false
	fromAddress := request.From.Hex()
	for _, account := range session.Accounts {
		if account == fromAddress {
			accountSupported = true
			break
		}
	}
	if !accountSupported {
		return fmt.Errorf("account %s not found in session", fromAddress)
	}

	return nil
}

// validateSignRequest validates a sign request
func (wc *WalletConnectClient) validateSignRequest(session *Session, request *SignRequest) error {
	// Check if method is supported
	methodSupported := false
	for _, method := range session.Methods {
		if method == request.Method {
			methodSupported = true
			break
		}
	}
	if !methodSupported {
		return fmt.Errorf("method %s not supported in session", request.Method)
	}

	// Check if account is in session
	accountSupported := false
	address := request.Address.Hex()
	for _, account := range session.Accounts {
		if account == address {
			accountSupported = true
			break
		}
	}
	if !accountSupported {
		return fmt.Errorf("account %s not found in session", address)
	}

	return nil
}

// Generator methods

// generateConnectionID generates a unique connection ID
func (wc *WalletConnectClient) generateConnectionID() string {
	return fmt.Sprintf("conn_%d", time.Now().UnixNano())
}

// generateSessionID generates a unique session ID
func (wc *WalletConnectClient) generateSessionID() string {
	return fmt.Sprintf("sess_%d", time.Now().UnixNano())
}

// generateProposalID generates a unique proposal ID
func (wc *WalletConnectClient) generateProposalID() string {
	return fmt.Sprintf("prop_%d", time.Now().UnixNano())
}

// generateEventID generates a unique event ID
func (wc *WalletConnectClient) generateEventID() string {
	return fmt.Sprintf("evt_%d", time.Now().UnixNano())
}

// generateTopic generates a unique topic
func (wc *WalletConnectClient) generateTopic() string {
	return fmt.Sprintf("topic_%d", time.Now().UnixNano())
}

// generatePeerID generates a unique peer ID
func (wc *WalletConnectClient) generatePeerID() string {
	return fmt.Sprintf("peer_%d", time.Now().UnixNano())
}

// generateConnectionURI generates a WalletConnect URI
func (wc *WalletConnectClient) generateConnectionURI(connectionID string, config SessionConfig) string {
	// Mock URI generation - in production would use actual WalletConnect URI format
	// The config parameter would be used in actual WalletConnect URI generation
	_ = config // Suppress unused parameter warning in mock implementation
	return fmt.Sprintf("wc:%s@2?relay-protocol=irn&symKey=%s", connectionID, wc.generateSymKey())
}

// generateSymKey generates a symmetric key
func (wc *WalletConnectClient) generateSymKey() string {
	return fmt.Sprintf("symkey_%d", time.Now().UnixNano())
}

// generateQRCode generates a QR code for the URI
func (wc *WalletConnectClient) generateQRCode(uri string) string {
	// Mock QR code generation - in production would generate actual QR code
	return fmt.Sprintf("data:image/png;base64,QRCode_%s", uri)
}

// generateDeepLink generates a deep link for the URI
func (wc *WalletConnectClient) generateDeepLink(uri string) string {
	if wc.config.DeepLinkConfig.CustomScheme != "" {
		return fmt.Sprintf("%s://wc?uri=%s", wc.config.DeepLinkConfig.CustomScheme, uri)
	}
	return fmt.Sprintf("walletconnect://wc?uri=%s", uri)
}

// generateTransactionHash generates a mock transaction hash
func (wc *WalletConnectClient) generateTransactionHash() string {
	return fmt.Sprintf("0x%x", time.Now().UnixNano())
}

// generateSignature generates a mock signature
func (wc *WalletConnectClient) generateSignature() string {
	return fmt.Sprintf("0x%x", time.Now().UnixNano())
}

// getEventTypeString converts event type to string
func (wc *WalletConnectClient) getEventTypeString(eventType EventType) string {
	switch eventType {
	case EventTypeSessionProposal:
		return "session_proposal"
	case EventTypeSessionUpdate:
		return "session_update"
	case EventTypeSessionDelete:
		return "session_delete"
	case EventTypeSessionRequest:
		return "session_request"
	case EventTypeSessionResponse:
		return "session_response"
	case EventTypeConnectionUpdate:
		return "connection_update"
	case EventTypeAccountsChanged:
		return "accounts_changed"
	case EventTypeChainChanged:
		return "chain_changed"
	case EventTypeDisconnect:
		return "disconnect"
	case EventTypeError:
		return "error"
	default:
		return "unknown"
	}
}

// Public interface methods

// IsRunning returns whether the WalletConnect client is running
func (wc *WalletConnectClient) IsRunning() bool {
	wc.mutex.RLock()
	defer wc.mutex.RUnlock()
	return wc.isRunning
}

// GetConfig returns the current configuration
func (wc *WalletConnectClient) GetConfig() WalletConnectConfig {
	return wc.config
}

// UpdateConfig updates the configuration
func (wc *WalletConnectClient) UpdateConfig(config WalletConnectConfig) error {
	wc.config = config
	wc.logger.Info("WalletConnect configuration updated")
	return nil
}

// GetSessionCount returns the number of active sessions
func (wc *WalletConnectClient) GetSessionCount() int {
	wc.sessionMutex.RLock()
	defer wc.sessionMutex.RUnlock()

	count := 0
	for _, session := range wc.sessions {
		if session.Status == SessionStatusActive {
			count++
		}
	}
	return count
}

// GetConnectionCount returns the number of pending connections
func (wc *WalletConnectClient) GetConnectionCount() int {
	wc.connectionMutex.RLock()
	defer wc.connectionMutex.RUnlock()

	count := 0
	for _, connection := range wc.connections {
		if connection.Status == ConnectionStatusPending {
			count++
		}
	}
	return count
}

// GetMetrics returns WalletConnect metrics
func (wc *WalletConnectClient) GetMetrics() map[string]interface{} {
	wc.sessionMutex.RLock()
	wc.connectionMutex.RLock()
	defer wc.sessionMutex.RUnlock()
	defer wc.connectionMutex.RUnlock()

	sessionCounts := make(map[string]int)
	for _, session := range wc.sessions {
		status := wc.getSessionStatusString(session.Status)
		sessionCounts[status]++
	}

	connectionCounts := make(map[string]int)
	for _, connection := range wc.connections {
		status := wc.getConnectionStatusString(connection.Status)
		connectionCounts[status]++
	}

	return map[string]interface{}{
		"total_sessions":    len(wc.sessions),
		"total_connections": len(wc.connections),
		"session_counts":    sessionCounts,
		"connection_counts": connectionCounts,
		"is_running":        wc.isRunning,
		"last_updated":      time.Now(),
	}
}

// Helper methods for status strings
func (wc *WalletConnectClient) getSessionStatusString(status SessionStatus) string {
	switch status {
	case SessionStatusPending:
		return "pending"
	case SessionStatusActive:
		return "active"
	case SessionStatusExpired:
		return "expired"
	case SessionStatusDisconnected:
		return "disconnected"
	case SessionStatusRejected:
		return "rejected"
	default:
		return "unknown"
	}
}

func (wc *WalletConnectClient) getConnectionStatusString(status ConnectionStatus) string {
	switch status {
	case ConnectionStatusPending:
		return "pending"
	case ConnectionStatusConnecting:
		return "connecting"
	case ConnectionStatusConnected:
		return "connected"
	case ConnectionStatusDisconnected:
		return "disconnected"
	case ConnectionStatusExpired:
		return "expired"
	case ConnectionStatusRejected:
		return "rejected"
	default:
		return "unknown"
	}
}

// GetDefaultWalletConnectConfig returns default WalletConnect configuration
func GetDefaultWalletConnectConfig() WalletConnectConfig {
	return WalletConnectConfig{
		Enabled:   true,
		ProjectID: "", // Must be provided by user
		RelayURL:  "wss://relay.walletconnect.com",
		Metadata: AppMetadata{
			Name:        "Crypto Wallet",
			Description: "Advanced cryptocurrency automation platform",
			URL:         "https://crypto-wallet.example.com",
			Icons:       []string{"https://crypto-wallet.example.com/icon.png"},
		},
		SupportedChains: []string{
			"eip155:1",     // Ethereum Mainnet
			"eip155:137",   // Polygon
			"eip155:42161", // Arbitrum
			"eip155:10",    // Optimism
		},
		SupportedMethods: []string{
			"eth_sendTransaction",
			"eth_signTransaction",
			"eth_sign",
			"personal_sign",
			"eth_signTypedData",
			"eth_signTypedData_v3",
			"eth_signTypedData_v4",
			"wallet_switchEthereumChain",
			"wallet_addEthereumChain",
		},
		SupportedEvents: []string{
			"accountsChanged",
			"chainChanged",
			"disconnect",
		},
		SessionTimeout:    24 * time.Hour,
		ConnectionTimeout: 5 * time.Minute,
		QRCodeConfig: QRCodeConfig{
			Enabled:    true,
			Size:       256,
			ErrorLevel: "M",
			Border:     4,
			Format:     "PNG",
		},
		DeepLinkConfig: DeepLinkConfig{
			Enabled:      true,
			CustomScheme: "cryptowallet",
		},
		SecurityConfig: SecurityConfig{
			RequireApproval:    true,
			SessionEncryption:  true,
			MessageSigning:     true,
			TransactionSigning: true,
			MaxSessionDuration: 7 * 24 * time.Hour, // 7 days
			AllowedOrigins:     []string{},
			BlockedAddresses:   []string{},
		},
		LoggingConfig: LoggingConfig{
			LogLevel:        "info",
			LogConnections:  true,
			LogTransactions: true,
			LogEvents:       true,
		},
	}
}
