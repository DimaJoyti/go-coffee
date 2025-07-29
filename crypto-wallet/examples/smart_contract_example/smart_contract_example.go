package main

import (
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/internal/blockchain"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/config"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
)

func main() {
	fmt.Println("🔗 Smart Contract Interaction Engine Example")
	fmt.Println("=============================================")

	// Initialize logger
	logConfig := config.LoggingConfig{
		Level:  "info",
		Format: "json",
	}
	logger := logger.NewLogger(logConfig)

	// Create smart contract engine configuration
	config := blockchain.GetDefaultSmartContractConfig()
	config.SupportedChains = []string{"ethereum", "polygon", "arbitrum"}
	config.EnableBatchTransactions = true
	config.GasMultiplier = decimal.NewFromFloat(1.2) // 20% gas buffer

	fmt.Printf("Configuration:\n")
	fmt.Printf("  Default Chain: %s\n", config.DefaultChain)
	fmt.Printf("  Supported Chains: %v\n", config.SupportedChains)
	fmt.Printf("  Batch Transactions: %v\n", config.EnableBatchTransactions)
	fmt.Printf("  Gas Multiplier: %s\n", config.GasMultiplier.String())
	fmt.Printf("  Supported Protocols: %v\n", getProtocolNames(config.ProtocolConfigs))
	fmt.Println()

	// Create smart contract engine
	engine := blockchain.NewSmartContractEngine(logger, config)

	// Note: In this example, we won't actually start the engine since it requires
	// real RPC endpoints. Instead, we'll demonstrate the API and configuration.

	fmt.Println("🏗️ Smart Contract Engine created successfully!")
	fmt.Println()

	// Demonstrate contract registration
	fmt.Println("📝 Registering custom smart contracts...")

	// Register a custom ERC-20 token contract
	erc20Contract := &blockchain.ContractDefinition{
		Name:     "Custom ERC-20 Token",
		Address:  common.HexToAddress("0xA0b86a33E6441b8C4505B6B8C0E4F7c3C4b5C8E1"),
		Chain:    "ethereum",
		Protocol: "erc20",
		Category: "Token",
		Version:  "1.0",
		Functions: map[string]blockchain.FunctionDef{
			"transfer": {
				Name:            "transfer",
				Signature:       "transfer(address,uint256)",
				StateMutability: "nonpayable",
				GasEstimate:     65000,
				Description:     "Transfer tokens to another address",
				Inputs: []blockchain.ParameterDef{
					{Name: "to", Type: "address", Description: "Recipient address"},
					{Name: "amount", Type: "uint256", Description: "Amount to transfer"},
				},
				Examples: []blockchain.FunctionExample{
					{
						Description: "Transfer 100 tokens",
						Inputs: map[string]interface{}{
							"to":     "0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E1",
							"amount": "100000000000000000000", // 100 tokens with 18 decimals
						},
						GasUsed: 65000,
					},
				},
			},
			"balanceOf": {
				Name:            "balanceOf",
				Signature:       "balanceOf(address)",
				StateMutability: "view",
				GasEstimate:     0,
				Description:     "Get token balance of an address",
				Inputs: []blockchain.ParameterDef{
					{Name: "owner", Type: "address", Description: "Address to check balance"},
				},
			},
		},
		SecurityScore: decimal.NewFromFloat(0.85),
		TrustLevel:    "medium",
		IsActive:      true,
	}

	err := engine.RegisterContract(erc20Contract)
	if err != nil {
		log.Printf("Failed to register ERC-20 contract: %v", err)
	} else {
		fmt.Printf("✅ Registered ERC-20 contract: %s\n", erc20Contract.Name)
	}

	// Register a custom DeFi protocol contract
	defiContract := &blockchain.ContractDefinition{
		Name:     "Custom DeFi Vault",
		Address:  common.HexToAddress("0xB1234567890123456789012345678901234567890"),
		Chain:    "ethereum",
		Protocol: "custom_defi",
		Category: "Vault",
		Version:  "2.0",
		Functions: map[string]blockchain.FunctionDef{
			"deposit": {
				Name:            "deposit",
				Signature:       "deposit(uint256)",
				StateMutability: "nonpayable",
				GasEstimate:     180000,
				Description:     "Deposit tokens into the vault",
				Inputs: []blockchain.ParameterDef{
					{Name: "amount", Type: "uint256", Description: "Amount to deposit"},
				},
			},
			"withdraw": {
				Name:            "withdraw",
				Signature:       "withdraw(uint256)",
				StateMutability: "nonpayable",
				GasEstimate:     150000,
				Description:     "Withdraw tokens from the vault",
				Inputs: []blockchain.ParameterDef{
					{Name: "amount", Type: "uint256", Description: "Amount to withdraw"},
				},
			},
			"getAPY": {
				Name:            "getAPY",
				Signature:       "getAPY()",
				StateMutability: "view",
				GasEstimate:     0,
				Description:     "Get current APY of the vault",
			},
		},
		SecurityScore: decimal.NewFromFloat(0.75),
		TrustLevel:    "medium",
		IsActive:      true,
	}

	err = engine.RegisterContract(defiContract)
	if err != nil {
		log.Printf("Failed to register DeFi contract: %v", err)
	} else {
		fmt.Printf("✅ Registered DeFi contract: %s\n", defiContract.Name)
	}

	fmt.Println()

	// Demonstrate listing contracts
	fmt.Println("📋 Listing all registered contracts:")
	contracts := engine.ListContracts()
	for i, contract := range contracts {
		fmt.Printf("  %d. %s (%s)\n", i+1, contract.Name, contract.Protocol)
		fmt.Printf("     Address: %s | Chain: %s\n", contract.Address.Hex(), contract.Chain)
		fmt.Printf("     Category: %s | Trust Level: %s\n", contract.Category, contract.TrustLevel)
		fmt.Printf("     Functions: %d | Security Score: %s\n",
			len(contract.Functions), contract.SecurityScore.StringFixed(2))
	}
	fmt.Println()

	// Demonstrate transaction request creation
	fmt.Println("🔄 Creating sample transaction requests...")

	// ERC-20 transfer transaction
	transferRequest := &blockchain.TransactionRequest{
		ID:              "transfer-tx-001",
		Chain:           "ethereum",
		ContractAddress: erc20Contract.Address,
		FunctionName:    "transfer",
		Inputs: []interface{}{
			common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E1"),
			big.NewInt(0).Mul(big.NewInt(100), big.NewInt(1e18)), // 100 tokens
		},
		GasLimit:  65000,
		GasPrice:  big.NewInt(20000000000), // 20 gwei
		From:      common.HexToAddress("0xSenderAddress123456789012345678901234567890"),
		Deadline:  time.Now().Add(10 * time.Minute),
		Priority:  blockchain.PriorityNormal,
		Metadata:  map[string]interface{}{"type": "token_transfer"},
		CreatedAt: time.Now(),
	}

	// DeFi vault deposit transaction
	depositRequest := &blockchain.TransactionRequest{
		ID:              "deposit-tx-001",
		Chain:           "ethereum",
		ContractAddress: defiContract.Address,
		FunctionName:    "deposit",
		Inputs: []interface{}{
			big.NewInt(0).Mul(big.NewInt(1000), big.NewInt(1e18)), // 1000 tokens
		},
		GasLimit:  180000,
		GasPrice:  big.NewInt(25000000000), // 25 gwei
		From:      common.HexToAddress("0xSenderAddress123456789012345678901234567890"),
		Deadline:  time.Now().Add(15 * time.Minute),
		Priority:  blockchain.PriorityHigh,
		Metadata:  map[string]interface{}{"type": "vault_deposit", "vault": "custom_defi"},
		CreatedAt: time.Now(),
	}

	fmt.Printf("📤 Created transaction requests:\n")
	fmt.Printf("  1. ERC-20 Transfer: %s\n", transferRequest.ID)
	fmt.Printf("     Function: %s | Gas Limit: %d\n", transferRequest.FunctionName, transferRequest.GasLimit)
	fmt.Printf("  2. Vault Deposit: %s\n", depositRequest.ID)
	fmt.Printf("     Function: %s | Gas Limit: %d\n", depositRequest.FunctionName, depositRequest.GasLimit)
	fmt.Println()

	// Demonstrate batch transaction creation
	fmt.Println("📦 Creating batch transaction request...")
	batchRequest := &blockchain.BatchTransactionRequest{
		ID:           "batch-001",
		Transactions: []blockchain.TransactionRequest{*transferRequest, *depositRequest},
		ExecuteOrder: blockchain.OrderSequential,
		FailureMode:  blockchain.FailureModeStopOnError,
		Deadline:     time.Now().Add(20 * time.Minute),
		CreatedAt:    time.Now(),
	}

	fmt.Printf("✅ Created batch transaction: %s\n", batchRequest.ID)
	fmt.Printf("   Transactions: %d | Order: Sequential | Failure Mode: Stop on Error\n",
		len(batchRequest.Transactions))
	fmt.Println()

	// Demonstrate contract call creation
	fmt.Println("👁️ Creating contract call requests...")

	// Read-only call to get token balance
	balanceCall := &blockchain.ContractCall{
		Chain:           "ethereum",
		ContractAddress: erc20Contract.Address,
		FunctionName:    "balanceOf",
		Inputs: []interface{}{
			common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E1"),
		},
		Metadata: map[string]interface{}{"type": "balance_check"},
	}

	// Read-only call to get vault APY
	apyCall := &blockchain.ContractCall{
		Chain:           "ethereum",
		ContractAddress: defiContract.Address,
		FunctionName:    "getAPY",
		Inputs:          []interface{}{},
		Metadata:        map[string]interface{}{"type": "apy_check"},
	}

	fmt.Printf("📞 Created contract calls:\n")
	fmt.Printf("  1. Token Balance Check: %s.%s\n",
		balanceCall.ContractAddress.Hex()[:10]+"...", balanceCall.FunctionName)
	fmt.Printf("  2. Vault APY Check: %s.%s\n",
		apyCall.ContractAddress.Hex()[:10]+"...", apyCall.FunctionName)
	fmt.Println()

	// Demonstrate protocol adapter features
	fmt.Println("🔌 Protocol adapter capabilities:")
	supportedProtocols := engine.GetSupportedProtocols()
	fmt.Printf("   Supported Protocols: %v\n", supportedProtocols)
	fmt.Println()

	// Demonstrate gas oracle features
	fmt.Println("⛽ Gas oracle demonstration:")
	fmt.Printf("   Gas price priorities: Low, Normal, High, Urgent\n")
	fmt.Printf("   EIP-1559 support: Yes\n")
	fmt.Printf("   Multi-chain support: %v\n", config.SupportedChains)
	fmt.Printf("   Gas optimization: %v\n", config.GasMultiplier.String())
	fmt.Println()

	// Show engine status
	fmt.Println("📊 Engine Status:")
	fmt.Printf("   Running: %v\n", engine.IsRunning())
	fmt.Printf("   Registered Contracts: %d\n", len(contracts))
	fmt.Printf("   Supported Protocols: %d\n", len(supportedProtocols))
	fmt.Printf("   Supported Chains: %d\n", len(config.SupportedChains))
	fmt.Println()

	fmt.Println("🎉 Smart Contract Interaction Engine example completed!")
	fmt.Println()
	fmt.Println("Key Features Demonstrated:")
	fmt.Println("  ✅ Contract registration and management")
	fmt.Println("  ✅ Transaction request creation and validation")
	fmt.Println("  ✅ Batch transaction support")
	fmt.Println("  ✅ Read-only contract calls")
	fmt.Println("  ✅ Protocol adapter system")
	fmt.Println("  ✅ Gas oracle integration")
	fmt.Println("  ✅ Multi-chain support")
	fmt.Println("  ✅ Comprehensive configuration")
	fmt.Println()
	fmt.Println("Note: This example demonstrates the API without executing actual")
	fmt.Println("blockchain transactions. In production, configure real RPC endpoints")
	fmt.Println("and start the engine to execute transactions.")
}

// Helper function to extract protocol names from config
func getProtocolNames(configs map[string]blockchain.ProtocolConfig) []string {
	names := make([]string, 0, len(configs))
	for name := range configs {
		names = append(names, name)
	}
	return names
}
