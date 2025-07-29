package main

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/internal/nlp/translation"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/config"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/DimaJoyti/go-coffee/pkg/ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// Helper function to create sample transactions
func createSampleTransactions() []*types.Transaction {
	var transactions []*types.Transaction

	// Simple ETH transfer
	tx1 := types.NewTransaction(
		1,
		common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E1"),
		ethereum.EtherToWei(big.NewFloat(1.0)), // 1 ETH
		21000,
		big.NewInt(20000000000), // 20 gwei
		nil,
	)
	transactions = append(transactions, tx1)

	// Contract interaction with data
	contractData := []byte{0xa9, 0x05, 0x9c, 0xbb} // transfer method signature
	contractData = append(contractData, make([]byte, 64)...)
	tx2 := types.NewTransaction(
		2,
		common.HexToAddress("0x1234567890123456789012345678901234erc20"),
		big.NewInt(0),
		100000,
		big.NewInt(25000000000), // 25 gwei
		contractData,
	)
	transactions = append(transactions, tx2)

	// High-value transaction
	tx3 := types.NewTransaction(
		3,
		common.HexToAddress("0x1111111111111111111111111111111111111111"),
		ethereum.EtherToWei(big.NewFloat(10.0)), // 10 ETH
		21000,
		big.NewInt(50000000000), // 50 gwei
		nil,
	)
	transactions = append(transactions, tx3)

	return transactions
}

// Helper function to create sample receipts
func createSampleReceipts(transactions []*types.Transaction) []*types.Receipt {
	var receipts []*types.Receipt

	for i, tx := range transactions {
		status := types.ReceiptStatusSuccessful
		gasUsed := uint64(21000)

		// Make one transaction failed for demonstration
		if i == 1 {
			status = types.ReceiptStatusFailed
			gasUsed = 50000
		}

		receipt := &types.Receipt{
			Status:      status,
			GasUsed:     gasUsed,
			BlockNumber: big.NewInt(int64(12345 + i)),
			BlockHash:   common.HexToHash(fmt.Sprintf("0x%064x", 12345+i)),
			TxHash:      tx.Hash(),
		}
		receipts = append(receipts, receipt)
	}

	return receipts
}

// Helper function to create sample addresses
func createSampleAddresses() []common.Address {
	return []common.Address{
		common.HexToAddress("0x1111111111111111111111111111111111111111"), // Exchange
		common.HexToAddress("0x2222222222222222222222222222222222222222"), // User
		common.HexToAddress("0x0000000000000000000000000000000000000000"), // Contract
		common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E1"), // Regular address
	}
}

func main() {
	fmt.Println("🗣️  Blockchain Data Translation System Example")
	fmt.Println("==============================================")

	// Initialize logger
	logConfig := config.LoggingConfig{
		Level:      "info",
		Format:     "console",
		Output:     "stdout",
		FilePath:   "",
		MaxSize:    100,
		MaxAge:     30,
		MaxBackups: 3,
		Compress:   true,
	}
	logger := logger.NewLogger(logConfig)

	// Create data translator configuration
	config := translation.GetDefaultDataTranslatorConfig()

	fmt.Printf("Configuration:\n")
	fmt.Printf("  Enabled: %v\n", config.Enabled)
	fmt.Printf("  Language: %s\n", config.Language)
	fmt.Printf("  Detail Level: %s\n", config.DetailLevel)
	fmt.Printf("  Include Metadata: %v\n", config.IncludeMetadata)
	fmt.Printf("  Address Resolution: %v\n", config.AddressResolverConfig.Enabled)
	fmt.Printf("  Contract Analysis: %v\n", config.ContractAnalyzerConfig.Enabled)
	fmt.Printf("  Transaction Parsing: %v\n", config.TransactionParserConfig.Enabled)
	fmt.Printf("  Event Decoding: %v\n", config.EventDecoderConfig.Enabled)
	fmt.Printf("  Value Formatting: %v\n", config.ValueFormatterConfig.Enabled)
	fmt.Printf("  Template Engine: %v\n", config.TemplateEngineConfig.Enabled)
	fmt.Printf("  Enrichment: %v\n", config.EnrichmentConfig.Enabled)
	fmt.Printf("  Caching: %v\n", config.CacheConfig.Enabled)
	fmt.Println()

	// Create blockchain data translator
	translator := translation.NewBlockchainDataTranslator(logger, config)

	// Start the translator
	ctx := context.Background()
	if err := translator.Start(ctx); err != nil {
		fmt.Printf("Failed to start blockchain data translator: %v\n", err)
		return
	}

	fmt.Println("✅ Blockchain data translator started successfully!")
	fmt.Println()

	// Show translator metrics
	fmt.Println("📊 Translator Metrics:")
	fmt.Println("=====================")
	metrics := translator.GetMetrics()
	fmt.Printf("  Is Running: %v\n", metrics["is_running"])
	fmt.Printf("  Cache Size: %v\n", metrics["cache_size"])
	fmt.Printf("  Language: %s\n", metrics["language"])
	fmt.Printf("  Detail Level: %s\n", metrics["detail_level"])
	fmt.Printf("  Cache Enabled: %v\n", metrics["cache_enabled"])
	fmt.Printf("  Enrichment Enabled: %v\n", metrics["enrichment_enabled"])
	fmt.Printf("  Address Resolver Enabled: %v\n", metrics["address_resolver_enabled"])
	fmt.Printf("  Contract Analyzer Enabled: %v\n", metrics["contract_analyzer_enabled"])
	fmt.Println()

	// Demonstrate transaction translation
	fmt.Println("💸 Transaction Translation Examples:")
	fmt.Println("===================================")

	transactions := createSampleTransactions()
	receipts := createSampleReceipts(transactions)

	for i, tx := range transactions {
		receipt := receipts[i]

		fmt.Printf("%d. Transaction %s:\n", i+1, tx.Hash().Hex()[:10]+"...")

		// Test different detail levels
		detailLevels := []string{"basic", "detailed", "technical"}

		for _, detailLevel := range detailLevels {
			options := &translation.TranslationOptions{
				DetailLevel:     detailLevel,
				IncludeMetadata: true,
				Format:          "text",
				Language:        "en",
			}

			result, err := translator.TranslateTransaction(ctx, tx, receipt, options)
			if err != nil {
				fmt.Printf("   ❌ %s translation failed: %v\n", detailLevel, err)
				continue
			}

			fmt.Printf("   📝 %s Translation:\n", strings.Title(detailLevel))
			fmt.Printf("      Summary: %s\n", result.Summary)

			if detailLevel == "detailed" || detailLevel == "technical" {
				// Show first few lines of detailed description
				lines := strings.Split(result.DetailedDescription, "\n")
				for j, line := range lines {
					if j >= 3 { // Show only first 3 lines
						fmt.Printf("      ...\n")
						break
					}
					if line != "" {
						fmt.Printf("      %s\n", line)
					}
				}
			}

			// Show key points
			if len(result.KeyPoints) > 0 {
				fmt.Printf("      Key Points:\n")
				for _, point := range result.KeyPoints {
					fmt.Printf("        • %s\n", point)
				}
			}

			// Show warnings
			if len(result.Warnings) > 0 {
				fmt.Printf("      Warnings:\n")
				for _, warning := range result.Warnings {
					fmt.Printf("        ⚠️  %s\n", warning)
				}
			}

			// Show recommendations
			if len(result.Recommendations) > 0 {
				fmt.Printf("      Recommendations:\n")
				for _, rec := range result.Recommendations {
					fmt.Printf("        💡 %s\n", rec)
				}
			}

			// Show metadata for technical level
			if detailLevel == "technical" && result.Metadata != nil {
				fmt.Printf("      Metadata:\n")
				fmt.Printf("        Processing Time: %v\n", result.Metadata.ProcessingTime)
				fmt.Printf("        Data Sources: %v\n", result.Metadata.DataSources)
			}

			fmt.Println()
		}
		fmt.Println()
	}

	// Demonstrate address translation
	fmt.Println("👤 Address Translation Examples:")
	fmt.Println("================================")

	addresses := createSampleAddresses()

	for i, address := range addresses {
		fmt.Printf("%d. Address %s:\n", i+1, address.Hex()[:10]+"...")

		options := &translation.TranslationOptions{
			DetailLevel: "detailed",
			Format:      "text",
			Language:    "en",
		}

		result, err := translator.TranslateAddress(ctx, address, options)
		if err != nil {
			fmt.Printf("   ❌ Translation failed: %v\n", err)
			continue
		}

		fmt.Printf("   Summary: %s\n", result.Summary)

		// Show detailed description
		lines := strings.Split(result.DetailedDescription, "\n")
		for _, line := range lines {
			if line != "" {
				fmt.Printf("   %s\n", line)
			}
		}

		// Show key points
		if len(result.KeyPoints) > 0 {
			fmt.Printf("   Key Points:\n")
			for _, point := range result.KeyPoints {
				fmt.Printf("     • %s\n", point)
			}
		}

		fmt.Println()
	}

	// Demonstrate contract translation
	fmt.Println("📋 Contract Translation Examples:")
	fmt.Println("=================================")

	contractAddresses := []common.Address{
		common.HexToAddress("0x1234567890123456789012345678901234erc20"),
		common.HexToAddress("0x123456789012345678901234567890123proxy"),
		common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	for i, contractAddr := range contractAddresses {
		fmt.Printf("%d. Contract %s:\n", i+1, contractAddr.Hex()[:10]+"...")

		options := &translation.TranslationOptions{
			DetailLevel: "detailed",
			Format:      "text",
			Language:    "en",
		}

		result, err := translator.TranslateContract(ctx, contractAddr, options)
		if err != nil {
			fmt.Printf("   ❌ Translation failed: %v\n", err)
			continue
		}

		fmt.Printf("   Summary: %s\n", result.Summary)

		// Show detailed description
		lines := strings.Split(result.DetailedDescription, "\n")
		for _, line := range lines {
			if line != "" {
				fmt.Printf("   %s\n", line)
			}
		}

		// Show key points
		if len(result.KeyPoints) > 0 {
			fmt.Printf("   Key Points:\n")
			for _, point := range result.KeyPoints {
				fmt.Printf("     • %s\n", point)
			}
		}

		fmt.Println()
	}

	// Demonstrate multi-language support
	fmt.Println("🌍 Multi-Language Translation Examples:")
	fmt.Println("=======================================")

	languages := []string{"en", "es", "fr", "de", "ja"}
	testTx := transactions[0]
	testReceipt := receipts[0]

	for _, lang := range languages {
		langConfig := translation.GetMultiLanguageConfig(lang)
		langTranslator := translation.NewBlockchainDataTranslator(logger, langConfig)

		if err := langTranslator.Start(ctx); err != nil {
			fmt.Printf("Failed to start %s translator: %v\n", lang, err)
			continue
		}

		options := &translation.TranslationOptions{
			DetailLevel: "basic",
			Format:      "text",
			Language:    lang,
		}

		result, err := langTranslator.TranslateTransaction(ctx, testTx, testReceipt, options)
		if err != nil {
			fmt.Printf("   ❌ %s translation failed: %v\n", lang, err)
			langTranslator.Stop()
			continue
		}

		langName := map[string]string{
			"en": "English",
			"es": "Spanish",
			"fr": "French",
			"de": "German",
			"ja": "Japanese",
		}[lang]

		fmt.Printf("   %s (%s): %s\n", langName, lang, result.Summary)

		langTranslator.Stop()
	}
	fmt.Println()

	// Show configuration profiles
	fmt.Println("🔧 Configuration Profiles:")
	fmt.Println("==========================")

	// Basic config
	basicConfig := translation.GetBasicDataTranslatorConfig()
	fmt.Printf("📱 Basic Configuration:\n")
	fmt.Printf("  Detail Level: %s\n", basicConfig.DetailLevel)
	fmt.Printf("  Include Metadata: %v\n", basicConfig.IncludeMetadata)
	fmt.Printf("  Contract ABI Analysis: %v\n", basicConfig.ContractAnalyzerConfig.AnalyzeABI)
	fmt.Printf("  Event Decoding: %v\n", basicConfig.EventDecoderConfig.DecodeKnownEvents)
	fmt.Printf("  Cache Size: %d\n", basicConfig.CacheConfig.MaxSize)
	fmt.Println()

	// Technical config
	technicalConfig := translation.GetTechnicalDataTranslatorConfig()
	fmt.Printf("🔬 Technical Configuration:\n")
	fmt.Printf("  Detail Level: %s\n", technicalConfig.DetailLevel)
	fmt.Printf("  Include Metadata: %v\n", technicalConfig.IncludeMetadata)
	fmt.Printf("  Include Raw Data: %v\n", technicalConfig.EventDecoderConfig.IncludeRawData)
	fmt.Printf("  Historical Data: %v\n", technicalConfig.EnrichmentConfig.IncludeHistoricalData)
	fmt.Printf("  Max Related Txs: %d\n", technicalConfig.EnrichmentConfig.MaxRelatedTxs)
	fmt.Printf("  Cache Size: %d\n", technicalConfig.CacheConfig.MaxSize)
	fmt.Println()

	// Show supported features
	fmt.Println("🛠️  Supported Features:")
	fmt.Println("======================")

	fmt.Println("Languages:")
	languages = translation.GetSupportedLanguages()
	for i, lang := range languages {
		if i > 0 && i%6 == 0 {
			fmt.Println()
		}
		fmt.Printf("  %-4s", lang)
	}
	fmt.Println("\n")

	fmt.Println("Detail Levels:")
	detailLevels := translation.GetSupportedDetailLevels()
	for _, level := range detailLevels {
		fmt.Printf("  %-10s", level)
	}
	fmt.Println("\n")

	fmt.Println("Currencies:")
	currencies := translation.GetSupportedCurrencies()
	for i, currency := range currencies {
		if i > 0 && i%10 == 0 {
			fmt.Println()
		}
		fmt.Printf("  %-4s", currency)
	}
	fmt.Println("\n")

	fmt.Println("Output Formats:")
	formats := translation.GetSupportedFormats()
	for _, format := range formats {
		fmt.Printf("  %-10s", format)
	}
	fmt.Println("\n")

	// Show descriptions
	fmt.Println("📖 Feature Descriptions:")
	fmt.Println("========================")

	fmt.Println("Translation Types:")
	typeDesc := translation.GetTranslationTypeDescription()
	for transType, desc := range typeDesc {
		fmt.Printf("  %s: %s\n", strings.Title(transType), desc)
	}
	fmt.Println()

	fmt.Println("Detail Levels:")
	detailDesc := translation.GetDetailLevelDescription()
	for level, desc := range detailDesc {
		fmt.Printf("  %s: %s\n", strings.Title(level), desc)
	}
	fmt.Println()

	// Best practices
	fmt.Println("💡 Best Practices:")
	fmt.Println("==================")
	fmt.Println("1. Choose appropriate detail level based on user expertise")
	fmt.Println("2. Enable caching for better performance with repeated queries")
	fmt.Println("3. Use enrichment features for comprehensive context")
	fmt.Println("4. Customize templates for specific use cases")
	fmt.Println("5. Implement proper error handling for translation failures")
	fmt.Println("6. Consider user preferences for language and currency")
	fmt.Println("7. Use address resolution for better user experience")
	fmt.Println("8. Enable contract analysis for smart contract interactions")
	fmt.Println()

	fmt.Println("🎉 Blockchain Data Translation System example completed!")
	fmt.Println()
	fmt.Println("Key Features Demonstrated:")
	fmt.Println("  ✅ Natural language translation of blockchain data")
	fmt.Println("  ✅ Multi-level detail support (basic, detailed, technical)")
	fmt.Println("  ✅ Multi-language support with localized templates")
	fmt.Println("  ✅ Address resolution with ENS and labels")
	fmt.Println("  ✅ Smart contract analysis and explanation")
	fmt.Println("  ✅ Transaction parsing with method call decoding")
	fmt.Println("  ✅ Event decoding and explanation")
	fmt.Println("  ✅ Value formatting with currency support")
	fmt.Println("  ✅ Context-aware translations with warnings and recommendations")
	fmt.Println("  ✅ Caching for improved performance")
	fmt.Println()
	fmt.Println("Note: This example demonstrates the system with mock data and translations.")
	fmt.Println("Integrate with real blockchain data and NLP models for production use.")

	// Stop the translator
	if err := translator.Stop(); err != nil {
		fmt.Printf("Error stopping blockchain data translator: %v\n", err)
	} else {
		fmt.Println("\n🛑 Blockchain data translator stopped")
	}
}
