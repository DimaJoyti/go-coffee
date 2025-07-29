package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/internal/walletconnect"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/config"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
)

func main() {
	fmt.Println("📱 WalletConnect Integration Example")
	fmt.Println("===================================")

	// Initialize logger
	logConfig := config.LoggingConfig{
		Level:  "info",
		Format: "json",
	}
	logger := logger.NewLogger(logConfig)

	// Create WalletConnect configuration
	config := walletconnect.GetDefaultWalletConnectConfig()
	config.ProjectID = "your-walletconnect-project-id" // Replace with actual project ID
	config.Metadata.Name = "Crypto Wallet Demo"
	config.Metadata.Description = "Advanced cryptocurrency automation platform with WalletConnect"
	config.Metadata.URL = "https://crypto-wallet-demo.example.com"
	config.QRCodeConfig.Size = 300
	config.DeepLinkConfig.CustomScheme = "cryptowallet"

	fmt.Printf("Configuration:\n")
	fmt.Printf("  Project ID: %s\n", config.ProjectID)
	fmt.Printf("  Relay URL: %s\n", config.RelayURL)
	fmt.Printf("  App Name: %s\n", config.Metadata.Name)
	fmt.Printf("  Supported Chains: %v\n", config.SupportedChains)
	fmt.Printf("  Supported Methods: %v\n", config.SupportedMethods[:3]) // Show first 3
	fmt.Printf("  QR Code Enabled: %v (Size: %d)\n", config.QRCodeConfig.Enabled, config.QRCodeConfig.Size)
	fmt.Printf("  Deep Links Enabled: %v (Scheme: %s)\n", config.DeepLinkConfig.Enabled, config.DeepLinkConfig.CustomScheme)
	fmt.Println()

	// Create WalletConnect client
	client := walletconnect.NewWalletConnectClient(logger, config)

	// Add event handlers
	setupEventHandlers(client)

	// Start the client
	ctx := context.Background()
	if err := client.Start(ctx); err != nil {
		log.Fatalf("Failed to start WalletConnect client: %v", err)
	}

	fmt.Println("✅ WalletConnect client started successfully!")
	fmt.Println()

	// Demonstrate creating a connection
	fmt.Println("🔗 Creating WalletConnect connection...")
	sessionConfig := walletconnect.SessionConfig{
		Chains: []string{
			"eip155:1",     // Ethereum Mainnet
			"eip155:137",   // Polygon
			"eip155:42161", // Arbitrum
		},
		Methods: []string{
			"eth_sendTransaction",
			"eth_signTransaction",
			"personal_sign",
			"eth_signTypedData_v4",
		},
		Events: []string{
			"accountsChanged",
			"chainChanged",
			"disconnect",
		},
	}

	connection, err := client.CreateConnection(sessionConfig)
	if err != nil {
		log.Fatalf("Failed to create connection: %v", err)
	}

	fmt.Printf("📱 Connection created successfully!\n")
	fmt.Printf("   Connection ID: %s\n", connection.ID)
	fmt.Printf("   WalletConnect URI: %s\n", connection.URI)
	fmt.Printf("   Status: %s\n", getConnectionStatusString(connection.Status))
	fmt.Printf("   Expires: %s\n", connection.ExpiresAt.Format("15:04:05"))
	
	if connection.QRCode != "" {
		fmt.Printf("   QR Code: %s...\n", connection.QRCode[:50])
	}
	
	if connection.DeepLink != "" {
		fmt.Printf("   Deep Link: %s\n", connection.DeepLink)
	}
	fmt.Println()

	// Simulate user scanning QR code and approving session
	fmt.Println("📲 Simulating wallet connection...")
	time.Sleep(1 * time.Second) // Simulate user interaction time

	// Approve the session with mock wallet accounts
	accounts := []string{
		"0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E1",
		"0x8ba1f109551bD432803012645Hac136c5C1b5C8E1",
	}

	session, err := client.ApproveSession(connection.ProposalID, accounts)
	if err != nil {
		log.Fatalf("Failed to approve session: %v", err)
	}

	fmt.Printf("✅ Session approved successfully!\n")
	fmt.Printf("   Session ID: %s\n", session.ID)
	fmt.Printf("   Topic: %s\n", session.Topic)
	fmt.Printf("   Connected Accounts: %v\n", session.Accounts)
	fmt.Printf("   Supported Chains: %v\n", session.Chains)
	fmt.Printf("   Available Methods: %v\n", session.Methods)
	fmt.Printf("   Session Expires: %s\n", session.Expiry.Format("2006-01-02 15:04:05"))
	fmt.Println()

	// Demonstrate sending a transaction request
	fmt.Println("💸 Sending transaction request...")
	txRequest := &walletconnect.TransactionRequest{
		ID:     "tx-demo-001",
		Method: "eth_sendTransaction",
		Chain:  "eip155:1", // Ethereum Mainnet
		From:   common.HexToAddress(accounts[0]),
		To:     &[]common.Address{common.HexToAddress("0x8ba1f109551bD432803012645Hac136c5C1b5C8E1")}[0],
		Value:  &[]decimal.Decimal{decimal.NewFromFloat(0.01)}[0], // 0.01 ETH
		Params: map[string]interface{}{
			"to":       "0x8ba1f109551bD432803012645Hac136c5C1b5C8E1",
			"value":    "0x2386f26fc10000", // 0.01 ETH in hex
			"gasLimit": "0x5208",            // 21000 gas
			"gasPrice": "0x4a817c800",       // 20 gwei
		},
		Metadata: map[string]interface{}{
			"type":        "transfer",
			"description": "Demo ETH transfer",
		},
	}

	txResponse, err := client.SendTransactionRequest(session.ID, txRequest)
	if err != nil {
		log.Printf("Failed to send transaction request: %v", err)
	} else {
		fmt.Printf("📤 Transaction request sent successfully!\n")
		fmt.Printf("   Request ID: %s\n", txResponse.ID)
		fmt.Printf("   Transaction Hash: %s\n", txResponse.TransactionHash)
		fmt.Printf("   Success: %v\n", txResponse.Success)
	}
	fmt.Println()

	// Demonstrate sending a sign request
	fmt.Println("✍️ Sending message signing request...")
	signRequest := &walletconnect.SignRequest{
		ID:      "sign-demo-001",
		Method:  "personal_sign",
		Address: common.HexToAddress(accounts[0]),
		Message: "Welcome to Crypto Wallet! Please sign this message to verify your identity.",
		Type:    walletconnect.SignTypePersonalSign,
		Metadata: map[string]interface{}{
			"type":        "authentication",
			"description": "Identity verification",
		},
	}

	signResponse, err := client.SendSignRequest(session.ID, signRequest)
	if err != nil {
		log.Printf("Failed to send sign request: %v", err)
	} else {
		fmt.Printf("✅ Sign request sent successfully!\n")
		fmt.Printf("   Request ID: %s\n", signResponse.ID)
		fmt.Printf("   Signature: %s\n", signResponse.Signature)
		fmt.Printf("   Success: %v\n", signResponse.Success)
	}
	fmt.Println()

	// Demonstrate session update
	fmt.Println("🔄 Updating session...")
	newAccounts := append(accounts, "0x9ba1f109551bD432803012645Hac136c5C1b5C8E1")
	newChains := append(session.Chains, "eip155:10") // Add Optimism

	err = client.UpdateSession(session.ID, newAccounts, newChains)
	if err != nil {
		log.Printf("Failed to update session: %v", err)
	} else {
		fmt.Printf("✅ Session updated successfully!\n")
		fmt.Printf("   New Accounts: %v\n", newAccounts)
		fmt.Printf("   New Chains: %v\n", newChains)
	}
	fmt.Println()

	// Show current metrics
	fmt.Println("📊 WalletConnect Metrics:")
	metrics := client.GetMetrics()
	fmt.Printf("   Total Sessions: %v\n", metrics["total_sessions"])
	fmt.Printf("   Total Connections: %v\n", metrics["total_connections"])
	fmt.Printf("   Active Sessions: %d\n", client.GetSessionCount())
	fmt.Printf("   Pending Connections: %d\n", client.GetConnectionCount())
	fmt.Printf("   Client Running: %v\n", metrics["is_running"])
	
	if sessionCounts, ok := metrics["session_counts"].(map[string]int); ok {
		fmt.Printf("   Session Status Breakdown:\n")
		for status, count := range sessionCounts {
			fmt.Printf("     %s: %d\n", status, count)
		}
	}
	fmt.Println()

	// Demonstrate QR code generation
	fmt.Println("📱 QR Code Generation Demo:")
	qrGenerator := walletconnect.NewQRGenerator(logger, config.QRCodeConfig)
	
	qrData, err := qrGenerator.GenerateQRCode(connection.URI)
	if err != nil {
		log.Printf("Failed to generate QR code: %v", err)
	} else {
		fmt.Printf("   QR Code Format: %s\n", qrData.Format)
		fmt.Printf("   QR Code Size: %dx%d\n", qrData.Size, qrData.Size)
		fmt.Printf("   Error Correction: %s\n", qrData.ErrorLevel)
		fmt.Printf("   Data URL: %s...\n", qrData.DataURL[:50])
	}
	fmt.Println()

	// Show active sessions
	fmt.Println("📋 Active Sessions:")
	activeSessions := client.GetActiveSessions()
	for i, sess := range activeSessions {
		fmt.Printf("   %d. Session ID: %s\n", i+1, sess.ID)
		fmt.Printf("      Status: %s | Accounts: %d | Chains: %d\n",
			getSessionStatusString(sess.Status), len(sess.Accounts), len(sess.Chains))
		fmt.Printf("      Created: %s | Expires: %s\n",
			sess.CreatedAt.Format("15:04:05"), sess.Expiry.Format("15:04:05"))
	}
	fmt.Println()

	// Simulate some time passing
	fmt.Println("⏰ Simulating active session usage...")
	time.Sleep(2 * time.Second)

	// Demonstrate disconnection
	fmt.Println("🔌 Disconnecting session...")
	err = client.DisconnectSession(session.ID, "Demo completed")
	if err != nil {
		log.Printf("Failed to disconnect session: %v", err)
	} else {
		fmt.Printf("✅ Session disconnected successfully!\n")
	}
	fmt.Println()

	// Final metrics
	fmt.Println("📊 Final Metrics:")
	finalMetrics := client.GetMetrics()
	fmt.Printf("   Active Sessions: %d\n", client.GetSessionCount())
	fmt.Printf("   Total Sessions Created: %v\n", finalMetrics["total_sessions"])
	fmt.Printf("   Total Connections Created: %v\n", finalMetrics["total_connections"])
	fmt.Println()

	fmt.Println("🎉 WalletConnect integration example completed!")
	fmt.Println()
	fmt.Println("Key Features Demonstrated:")
	fmt.Println("  ✅ WalletConnect v2 session management")
	fmt.Println("  ✅ QR code and deep link generation")
	fmt.Println("  ✅ Transaction and signing requests")
	fmt.Println("  ✅ Multi-chain support")
	fmt.Println("  ✅ Event handling system")
	fmt.Println("  ✅ Session updates and disconnection")
	fmt.Println("  ✅ Comprehensive metrics and monitoring")

	// Stop the client
	if err := client.Stop(); err != nil {
		log.Printf("Error stopping WalletConnect client: %v", err)
	} else {
		fmt.Println("\n🛑 WalletConnect client stopped")
	}
}

// setupEventHandlers sets up event handlers for the WalletConnect client
func setupEventHandlers(client *walletconnect.WalletConnectClient) {
	// Session proposal handler
	client.AddEventHandler(walletconnect.EventTypeSessionProposal, func(event *walletconnect.Event) error {
		fmt.Printf("🔔 Event: Session Proposal (ID: %s)\n", event.ID)
		return nil
	})

	// Session update handler
	client.AddEventHandler(walletconnect.EventTypeSessionUpdate, func(event *walletconnect.Event) error {
		fmt.Printf("🔔 Event: Session Update (ID: %s)\n", event.ID)
		return nil
	})

	// Disconnect handler
	client.AddEventHandler(walletconnect.EventTypeDisconnect, func(event *walletconnect.Event) error {
		fmt.Printf("🔔 Event: Session Disconnect (ID: %s)\n", event.ID)
		return nil
	})

	// Transaction request handler
	client.AddEventHandler(walletconnect.EventTypeSessionRequest, func(event *walletconnect.Event) error {
		fmt.Printf("🔔 Event: Session Request (ID: %s)\n", event.ID)
		return nil
	})
}

// Helper functions for status strings
func getConnectionStatusString(status walletconnect.ConnectionStatus) string {
	switch status {
	case walletconnect.ConnectionStatusPending:
		return "Pending"
	case walletconnect.ConnectionStatusConnecting:
		return "Connecting"
	case walletconnect.ConnectionStatusConnected:
		return "Connected"
	case walletconnect.ConnectionStatusDisconnected:
		return "Disconnected"
	case walletconnect.ConnectionStatusExpired:
		return "Expired"
	case walletconnect.ConnectionStatusRejected:
		return "Rejected"
	default:
		return "Unknown"
	}
}

func getSessionStatusString(status walletconnect.SessionStatus) string {
	switch status {
	case walletconnect.SessionStatusPending:
		return "Pending"
	case walletconnect.SessionStatusActive:
		return "Active"
	case walletconnect.SessionStatusExpired:
		return "Expired"
	case walletconnect.SessionStatusDisconnected:
		return "Disconnected"
	case walletconnect.SessionStatusRejected:
		return "Rejected"
	default:
		return "Unknown"
	}
}
