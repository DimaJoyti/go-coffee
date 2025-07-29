package translation

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/config"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a test logger
func createTestLoggerForTranslation() *logger.Logger {
	logConfig := config.LoggingConfig{
		Level:  "info",
		Format: "json",
	}
	return logger.NewLogger(logConfig)
}

// Helper function to create test transaction
func createTestTransactionForTranslation() (*types.Transaction, *types.Receipt) {
	to := common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E1")
	value := big.NewInt(1000000000000000000) // 1 ETH
	gasPrice := big.NewInt(20000000000)      // 20 gwei
	data := []byte{}

	tx := types.NewTransaction(1, to, value, 21000, gasPrice, data)

	receipt := &types.Receipt{
		Status:      types.ReceiptStatusSuccessful,
		GasUsed:     21000,
		BlockNumber: big.NewInt(12345),
		BlockHash:   common.HexToHash("0xabcd"),
		TxHash:      tx.Hash(),
	}

	return tx, receipt
}

func TestNewBlockchainDataTranslator(t *testing.T) {
	logger := createTestLoggerForTranslation()
	config := GetDefaultDataTranslatorConfig()

	translator := NewBlockchainDataTranslator(logger, config)

	assert.NotNil(t, translator)
	assert.Equal(t, config.Enabled, translator.config.Enabled)
	assert.Equal(t, config.Language, translator.config.Language)
	assert.False(t, translator.IsRunning())
	assert.NotNil(t, translator.addressResolver)
	assert.NotNil(t, translator.contractAnalyzer)
	assert.NotNil(t, translator.transactionParser)
	assert.NotNil(t, translator.eventDecoder)
	assert.NotNil(t, translator.valueFormatter)
	assert.NotNil(t, translator.templateEngine)
	assert.NotNil(t, translator.contextAnalyzer)
	assert.NotNil(t, translator.sentenceBuilder)
	assert.NotNil(t, translator.enrichmentEngine)
	assert.NotNil(t, translator.metadataProvider)
}

func TestBlockchainDataTranslator_StartStop(t *testing.T) {
	logger := createTestLoggerForTranslation()
	config := GetDefaultDataTranslatorConfig()

	translator := NewBlockchainDataTranslator(logger, config)
	ctx := context.Background()

	err := translator.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, translator.IsRunning())

	err = translator.Stop()
	assert.NoError(t, err)
	assert.False(t, translator.IsRunning())
}

func TestBlockchainDataTranslator_StartDisabled(t *testing.T) {
	logger := createTestLoggerForTranslation()
	config := GetDefaultDataTranslatorConfig()
	config.Enabled = false

	translator := NewBlockchainDataTranslator(logger, config)
	ctx := context.Background()

	err := translator.Start(ctx)
	assert.NoError(t, err)
	assert.False(t, translator.IsRunning()) // Should remain false when disabled
}

func TestBlockchainDataTranslator_TranslateTransaction(t *testing.T) {
	logger := createTestLoggerForTranslation()
	config := GetDefaultDataTranslatorConfig()

	translator := NewBlockchainDataTranslator(logger, config)
	ctx := context.Background()

	// Start the translator
	err := translator.Start(ctx)
	require.NoError(t, err)
	defer translator.Stop()

	// Create test transaction
	tx, receipt := createTestTransactionForTranslation()

	// Test different detail levels
	detailLevels := []string{"basic", "detailed", "technical"}

	for _, detailLevel := range detailLevels {
		options := &TranslationOptions{
			DetailLevel:     detailLevel,
			IncludeMetadata: true,
			Format:          "text",
			Language:        "en",
		}

		result, err := translator.TranslateTransaction(ctx, tx, receipt, options)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.Summary)
		assert.NotEmpty(t, result.DetailedDescription)
		assert.NotNil(t, result.KeyPoints)
		assert.NotNil(t, result.Metadata)
		assert.Equal(t, "transaction", result.Metadata.TranslationType)

		// Verify detail level affects description length
		if detailLevel == "technical" {
			assert.Contains(t, result.DetailedDescription, "Technical Details")
		}
	}
}

func TestBlockchainDataTranslator_TranslateAddress(t *testing.T) {
	logger := createTestLoggerForTranslation()
	config := GetDefaultDataTranslatorConfig()

	translator := NewBlockchainDataTranslator(logger, config)
	ctx := context.Background()

	// Start the translator
	err := translator.Start(ctx)
	require.NoError(t, err)
	defer translator.Stop()

	// Test different address types
	addresses := []common.Address{
		common.HexToAddress("0x1111111111111111111111111111111111111111"), // Exchange
		common.HexToAddress("0x2222222222222222222222222222222222222222"), // User
		common.HexToAddress("0x0000000000000000000000000000000000000000"), // Contract
	}

	for _, address := range addresses {
		options := &TranslationOptions{
			DetailLevel: "detailed",
			Format:      "text",
			Language:    "en",
		}

		result, err := translator.TranslateAddress(ctx, address, options)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.Summary)
		assert.NotEmpty(t, result.DetailedDescription)
		assert.NotNil(t, result.KeyPoints)
		assert.Equal(t, "address", result.Metadata.TranslationType)
	}
}

func TestBlockchainDataTranslator_TranslateContract(t *testing.T) {
	logger := createTestLoggerForTranslation()
	config := GetDefaultDataTranslatorConfig()

	translator := NewBlockchainDataTranslator(logger, config)
	ctx := context.Background()

	// Start the translator
	err := translator.Start(ctx)
	require.NoError(t, err)
	defer translator.Stop()

	// Test contract address
	contractAddress := common.HexToAddress("0x1234567890123456789012345678901234567890")

	options := &TranslationOptions{
		DetailLevel: "detailed",
		Format:      "text",
		Language:    "en",
	}

	result, err := translator.TranslateContract(ctx, contractAddress, options)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Summary)
	assert.NotEmpty(t, result.DetailedDescription)
	assert.NotNil(t, result.KeyPoints)
	assert.Equal(t, "contract", result.Metadata.TranslationType)
}

func TestBlockchainDataTranslator_Cache(t *testing.T) {
	logger := createTestLoggerForTranslation()
	config := GetDefaultDataTranslatorConfig()
	config.CacheConfig.TTL = 1 * time.Second // Short TTL for testing

	translator := NewBlockchainDataTranslator(logger, config)
	ctx := context.Background()

	// Start the translator
	err := translator.Start(ctx)
	require.NoError(t, err)
	defer translator.Stop()

	// Create test transaction
	tx, receipt := createTestTransactionForTranslation()
	options := &TranslationOptions{
		DetailLevel: "basic",
		Format:      "text",
		Language:    "en",
	}

	// First translation (should cache)
	result1, err := translator.TranslateTransaction(ctx, tx, receipt, options)
	assert.NoError(t, err)
	assert.NotNil(t, result1)

	// Second translation (should use cache)
	result2, err := translator.TranslateTransaction(ctx, tx, receipt, options)
	assert.NoError(t, err)
	assert.NotNil(t, result2)

	// Results should be identical (from cache)
	assert.Equal(t, result1.Summary, result2.Summary)

	// Wait for cache to expire
	time.Sleep(2 * time.Second)

	// Third translation (should not use cache)
	result3, err := translator.TranslateTransaction(ctx, tx, receipt, options)
	assert.NoError(t, err)
	assert.NotNil(t, result3)
}

func TestBlockchainDataTranslator_GetMetrics(t *testing.T) {
	logger := createTestLoggerForTranslation()
	config := GetDefaultDataTranslatorConfig()

	translator := NewBlockchainDataTranslator(logger, config)

	metrics := translator.GetMetrics()
	assert.NotNil(t, metrics)

	// Validate metrics structure
	assert.Contains(t, metrics, "is_running")
	assert.Contains(t, metrics, "cache_size")
	assert.Contains(t, metrics, "language")
	assert.Contains(t, metrics, "detail_level")
	assert.Contains(t, metrics, "cache_enabled")
	assert.Contains(t, metrics, "enrichment_enabled")
	assert.Contains(t, metrics, "address_resolver_enabled")
	assert.Contains(t, metrics, "contract_analyzer_enabled")

	assert.Equal(t, false, metrics["is_running"])
	assert.Equal(t, "en", metrics["language"])
	assert.Equal(t, "detailed", metrics["detail_level"])
	assert.Equal(t, true, metrics["cache_enabled"])
}

func TestMockAddressResolver(t *testing.T) {
	resolver := &MockAddressResolver{}
	ctx := context.Background()

	// Test different address patterns
	addresses := []struct {
		address  common.Address
		expected string
	}{
		{common.HexToAddress("0x1111111111111111111111111111111111111111"), "Mock Exchange"},
		{common.HexToAddress("0x2222222222222222222222222222222222222222"), "Mock User"},
		{common.HexToAddress("0x0000000000000000000000000000000000000000"), "Mock Contract"},
	}

	for _, test := range addresses {
		info, err := resolver.ResolveAddress(ctx, test.address)
		assert.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, test.expected, info.Label)
		assert.Equal(t, test.address, info.Address)
	}

	// Test ENS resolution
	ensAddress := common.HexToAddress("0x1111111111111111111111111111111111111111")
	ensName, err := resolver.ResolveENS(ctx, ensAddress)
	assert.NoError(t, err)
	assert.Equal(t, "exchange.eth", ensName)

	// Test address label
	label := resolver.GetAddressLabel(ensAddress)
	assert.Equal(t, "Mock Exchange", label)
}

func TestMockContractAnalyzer(t *testing.T) {
	analyzer := &MockContractAnalyzer{}
	ctx := context.Background()

	// Test ERC20 contract
	erc20Address := common.HexToAddress("0x1234567890123456789012345678901234erc20")
	contractInfo, err := analyzer.AnalyzeContract(ctx, erc20Address)
	assert.NoError(t, err)
	assert.NotNil(t, contractInfo)
	assert.Equal(t, "Mock Token", contractInfo.Name)
	assert.Equal(t, "MOCK", contractInfo.Symbol)
	assert.Equal(t, "ERC20", contractInfo.Standard)
	assert.False(t, contractInfo.IsProxy)
	assert.NotEmpty(t, contractInfo.Functions)
	assert.NotEmpty(t, contractInfo.Events)

	// Test proxy contract
	proxyAddress := common.HexToAddress("0x123456789012345678901234567890123proxy")
	proxyInfo, err := analyzer.AnalyzeContract(ctx, proxyAddress)
	assert.NoError(t, err)
	assert.NotNil(t, proxyInfo)
	assert.True(t, proxyInfo.IsProxy)
	assert.NotNil(t, proxyInfo.Implementation)

	// Test standard detection
	standard, err := analyzer.DetectStandard(ctx, erc20Address)
	assert.NoError(t, err)
	assert.Equal(t, "ERC20", standard)

	// Test proxy resolution
	impl, err := analyzer.ResolveProxy(ctx, proxyAddress)
	assert.NoError(t, err)
	assert.NotNil(t, impl)
}

func TestMockTransactionParser(t *testing.T) {
	parser := &MockTransactionParser{}
	ctx := context.Background()

	// Create test transaction
	tx, receipt := createTestTransactionForTranslation()

	// Test transaction parsing
	parsedTx, err := parser.ParseTransaction(ctx, tx, receipt)
	assert.NoError(t, err)
	assert.NotNil(t, parsedTx)
	assert.Equal(t, tx.Hash(), parsedTx.Hash)
	assert.Equal(t, "transfer", parsedTx.Type)
	assert.NotNil(t, parsedTx.From)
	assert.NotNil(t, parsedTx.To)
	assert.NotNil(t, parsedTx.GasAnalysis)
	assert.Equal(t, "success", parsedTx.Status)

	// Test gas analysis
	gasAnalysis := parser.AnalyzeGasUsage(tx, receipt)
	assert.NotNil(t, gasAnalysis)
	assert.Equal(t, tx.Gas(), gasAnalysis.GasLimit)
	assert.Equal(t, receipt.GasUsed, gasAnalysis.GasUsed)
}

func TestMockValueFormatter(t *testing.T) {
	formatter := &MockValueFormatter{}

	// Test value formatting
	ethValue := formatter.FormatValue(decimal.NewFromInt(1000000000000000000), "ETH")
	assert.Contains(t, ethValue, "1.0000 ETH")

	// Test gas formatting
	gasFormatted := formatter.FormatGas(21000, decimal.NewFromInt(20000000000))
	assert.Contains(t, gasFormatted, "ETH")
	assert.Contains(t, gasFormatted, "gwei")

	// Test time formatting
	now := time.Now()
	timeFormatted := formatter.FormatTime(now)
	assert.Contains(t, timeFormatted, "UTC")

	// Test address formatting
	address := common.HexToAddress("0x1234567890123456789012345678901234567890")
	addressInfo := &AddressInfo{
		Address: address,
		Label:   "Test Address",
	}
	addressFormatted := formatter.FormatAddress(address, addressInfo)
	assert.Contains(t, addressFormatted, "Test Address")
}

func TestGetDefaultDataTranslatorConfig(t *testing.T) {
	config := GetDefaultDataTranslatorConfig()

	assert.True(t, config.Enabled)
	assert.Equal(t, "en", config.Language)
	assert.Equal(t, "detailed", config.DetailLevel)
	assert.True(t, config.IncludeMetadata)

	// Check address resolver config
	assert.True(t, config.AddressResolverConfig.Enabled)
	assert.True(t, config.AddressResolverConfig.ResolveENS)

	// Check contract analyzer config
	assert.True(t, config.ContractAnalyzerConfig.Enabled)
	assert.True(t, config.ContractAnalyzerConfig.AnalyzeABI)

	// Check cache config
	assert.True(t, config.CacheConfig.Enabled)
	assert.Equal(t, 10000, config.CacheConfig.MaxSize)
	assert.Equal(t, 1*time.Hour, config.CacheConfig.TTL)
}

func TestValidateDataTranslatorConfig(t *testing.T) {
	// Test valid config
	validConfig := GetDefaultDataTranslatorConfig()
	err := ValidateDataTranslatorConfig(validConfig)
	assert.NoError(t, err)

	// Test disabled config
	disabledConfig := GetDefaultDataTranslatorConfig()
	disabledConfig.Enabled = false
	err = ValidateDataTranslatorConfig(disabledConfig)
	assert.NoError(t, err)

	// Test invalid language
	invalidLangConfig := GetDefaultDataTranslatorConfig()
	invalidLangConfig.Language = "invalid"
	err = ValidateDataTranslatorConfig(invalidLangConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported language")

	// Test invalid detail level
	invalidDetailConfig := GetDefaultDataTranslatorConfig()
	invalidDetailConfig.DetailLevel = "invalid"
	err = ValidateDataTranslatorConfig(invalidDetailConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported detail level")

	// Test invalid currency
	invalidCurrencyConfig := GetDefaultDataTranslatorConfig()
	invalidCurrencyConfig.ValueFormatterConfig.Currency = "INVALID"
	err = ValidateDataTranslatorConfig(invalidCurrencyConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported currency")
}

func TestConfigVariants(t *testing.T) {
	// Test basic config
	basicConfig := GetBasicDataTranslatorConfig()
	assert.Equal(t, "basic", basicConfig.DetailLevel)
	assert.False(t, basicConfig.IncludeMetadata)
	assert.False(t, basicConfig.ContractAnalyzerConfig.AnalyzeABI)

	// Test technical config
	technicalConfig := GetTechnicalDataTranslatorConfig()
	assert.Equal(t, "technical", technicalConfig.DetailLevel)
	assert.True(t, technicalConfig.IncludeMetadata)
	assert.True(t, technicalConfig.EventDecoderConfig.IncludeRawData)

	// Test multi-language config
	spanishConfig := GetMultiLanguageConfig("es")
	assert.Equal(t, "es", spanishConfig.Language)
	assert.Equal(t, "EUR", spanishConfig.ValueFormatterConfig.Currency)

	// Validate all configs
	assert.NoError(t, ValidateDataTranslatorConfig(basicConfig))
	assert.NoError(t, ValidateDataTranslatorConfig(technicalConfig))
	assert.NoError(t, ValidateDataTranslatorConfig(spanishConfig))
}

func TestUtilityFunctions(t *testing.T) {
	// Test supported languages
	languages := GetSupportedLanguages()
	assert.NotEmpty(t, languages)
	assert.Contains(t, languages, "en")
	assert.Contains(t, languages, "es")

	// Test supported detail levels
	detailLevels := GetSupportedDetailLevels()
	assert.NotEmpty(t, detailLevels)
	assert.Contains(t, detailLevels, "basic")
	assert.Contains(t, detailLevels, "detailed")
	assert.Contains(t, detailLevels, "technical")

	// Test supported currencies
	currencies := GetSupportedCurrencies()
	assert.NotEmpty(t, currencies)
	assert.Contains(t, currencies, "USD")
	assert.Contains(t, currencies, "EUR")

	// Test optimal config for use case
	config, err := GetOptimalConfigForUseCase("basic")
	assert.NoError(t, err)
	assert.Equal(t, "basic", config.DetailLevel)

	// Test invalid use case
	_, err = GetOptimalConfigForUseCase("invalid")
	assert.Error(t, err)

	// Test descriptions
	langDesc := GetLanguageDescription()
	assert.NotEmpty(t, langDesc)
	assert.Contains(t, langDesc, "en")

	detailDesc := GetDetailLevelDescription()
	assert.NotEmpty(t, detailDesc)
	assert.Contains(t, detailDesc, "basic")

	useCaseDesc := GetUseCaseDescription()
	assert.NotEmpty(t, useCaseDesc)
	assert.Contains(t, useCaseDesc, "basic")

	// Test template examples
	templates := GetTemplateExamples()
	assert.NotEmpty(t, templates)
	assert.Contains(t, templates, "en")
	assert.Contains(t, templates["en"], "simple_transfer")

	// Test common labels
	labels := GetCommonAddressLabels()
	assert.NotEmpty(t, labels)
	assert.Contains(t, labels, "0x0000000000000000000000000000000000000000")

	// Test event signatures
	eventSigs := GetEventSignatures()
	assert.NotEmpty(t, eventSigs)

	// Test method signatures
	methodSigs := GetMethodSignatures()
	assert.NotEmpty(t, methodSigs)
}
