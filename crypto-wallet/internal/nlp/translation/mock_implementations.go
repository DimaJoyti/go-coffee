package translation

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/shopspring/decimal"
)

// Mock implementations for testing and demonstration

// MockAddressResolver provides mock address resolution
type MockAddressResolver struct{}

func (m *MockAddressResolver) ResolveAddress(ctx context.Context, address common.Address) (*AddressInfo, error) {
	// Mock address resolution based on known patterns
	addressStr := address.Hex()

	var addressType, label, ensName string
	var isVerified bool
	var tags []string
	var description string

	// Simulate different address types based on address patterns
	switch {
	case strings.HasSuffix(addressStr, "0000"):
		addressType = "contract"
		label = "Mock Contract"
		isVerified = true
		tags = []string{"defi", "verified"}
		description = "A verified smart contract"
	case strings.HasSuffix(addressStr, "1111"):
		addressType = "exchange"
		label = "Mock Exchange"
		ensName = "exchange.eth"
		isVerified = true
		tags = []string{"exchange", "centralized"}
		description = "Cryptocurrency exchange hot wallet"
	case strings.HasSuffix(addressStr, "2222"):
		addressType = "eoa"
		label = "Mock User"
		ensName = "user.eth"
		tags = []string{"user"}
		description = "Regular user wallet"
	default:
		addressType = "eoa"
		label = "Unknown Address"
	}

	return &AddressInfo{
		Address:     address,
		Label:       label,
		Type:        addressType,
		ENSName:     ensName,
		IsVerified:  isVerified,
		Tags:        tags,
		Description: description,
	}, nil
}

func (m *MockAddressResolver) ResolveENS(ctx context.Context, address common.Address) (string, error) {
	// Mock ENS resolution
	addressStr := address.Hex()
	if strings.HasSuffix(addressStr, "1111") {
		return "exchange.eth", nil
	}
	if strings.HasSuffix(addressStr, "2222") {
		return "user.eth", nil
	}
	return "", fmt.Errorf("no ENS name found")
}

func (m *MockAddressResolver) GetAddressLabel(address common.Address) string {
	addressStr := address.Hex()
	switch {
	case strings.HasSuffix(addressStr, "0000"):
		return "Mock Contract"
	case strings.HasSuffix(addressStr, "1111"):
		return "Mock Exchange"
	case strings.HasSuffix(addressStr, "2222"):
		return "Mock User"
	default:
		return "Unknown"
	}
}

// MockContractAnalyzer provides mock contract analysis
type MockContractAnalyzer struct{}

func (m *MockContractAnalyzer) AnalyzeContract(ctx context.Context, address common.Address) (*ContractInfo, error) {
	// Mock contract analysis
	addressStr := address.Hex()

	var name, symbol, standard string
	var isProxy bool
	var implementation *common.Address
	var functions []FunctionInfo
	var events []EventInfo

	// Simulate different contract types
	switch {
	case strings.Contains(addressStr, "erc20"):
		name = "Mock Token"
		symbol = "MOCK"
		standard = "ERC20"
		functions = []FunctionInfo{
			{
				Name:        "transfer",
				Signature:   "transfer(address,uint256)",
				Selector:    []byte{0xa9, 0x05, 0x9c, 0xbb},
				Description: "Transfer tokens to another address",
				Parameters: []ParameterInfo{
					{Name: "to", Type: "address", Description: "Recipient address"},
					{Name: "amount", Type: "uint256", Description: "Amount to transfer"},
				},
			},
		}
		events = []EventInfo{
			{
				Name:        "Transfer",
				Signature:   "Transfer(address,address,uint256)",
				Topic0:      common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"),
				Description: "Emitted when tokens are transferred",
				Parameters: []ParameterInfo{
					{Name: "from", Type: "address", Description: "Sender address", Indexed: true},
					{Name: "to", Type: "address", Description: "Recipient address", Indexed: true},
					{Name: "value", Type: "uint256", Description: "Amount transferred"},
				},
			},
		}
	case strings.Contains(addressStr, "proxy"):
		name = "Proxy Contract"
		standard = "Proxy"
		isProxy = true
		implAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")
		implementation = &implAddr
	default:
		name = "Unknown Contract"
		standard = "Unknown"
	}

	return &ContractInfo{
		Address:        address,
		Name:           name,
		Symbol:         symbol,
		Standard:       standard,
		IsProxy:        isProxy,
		Implementation: implementation,
		ABI:            []byte("[]"), // Mock empty ABI
		Functions:      functions,
		Events:         events,
	}, nil
}

func (m *MockContractAnalyzer) DetectStandard(ctx context.Context, address common.Address) (string, error) {
	addressStr := address.Hex()
	switch {
	case strings.Contains(addressStr, "erc20"):
		return "ERC20", nil
	case strings.Contains(addressStr, "erc721"):
		return "ERC721", nil
	case strings.Contains(addressStr, "erc1155"):
		return "ERC1155", nil
	default:
		return "Unknown", nil
	}
}

func (m *MockContractAnalyzer) ResolveProxy(ctx context.Context, address common.Address) (*common.Address, error) {
	if strings.Contains(address.Hex(), "proxy") {
		impl := common.HexToAddress("0x1234567890123456789012345678901234567890")
		return &impl, nil
	}
	return nil, fmt.Errorf("not a proxy contract")
}

// MockTransactionParser provides mock transaction parsing
type MockTransactionParser struct{}

func (m *MockTransactionParser) ParseTransaction(ctx context.Context, tx *types.Transaction, receipt *types.Receipt) (*ParsedTransaction, error) {
	// Mock transaction parsing
	from := common.HexToAddress("0x1111111111111111111111111111111111111111")
	fromInfo := &AddressInfo{
		Address: from,
		Label:   "Sender",
		Type:    "eoa",
	}

	var toInfo *AddressInfo
	if tx.To() != nil {
		toInfo = &AddressInfo{
			Address: *tx.To(),
			Label:   "Recipient",
			Type:    "eoa",
		}
	}

	// Mock gas analysis
	gasAnalysis := &GasAnalysis{
		GasLimit:         tx.Gas(),
		GasUsed:          receipt.GasUsed,
		GasPrice:         decimal.NewFromBigInt(tx.GasPrice(), 0),
		GasCost:          decimal.NewFromBigInt(tx.GasPrice(), 0).Mul(decimal.NewFromInt(int64(receipt.GasUsed))),
		GasEfficiency:    decimal.NewFromFloat(0.8),
		EstimatedCostUSD: decimal.NewFromFloat(25.50),
	}

	// Mock method call if data exists
	var methodCall *MethodCall
	if len(tx.Data()) >= 4 {
		methodCall = &MethodCall{
			Function: &FunctionInfo{
				Name:        "transfer",
				Signature:   "transfer(address,uint256)",
				Description: "Transfer tokens",
			},
			Parameters: map[string]interface{}{
				"to":     tx.To().Hex(),
				"amount": "1000000000000000000",
			},
			DecodedParameters: map[string]string{
				"to":     tx.To().Hex(),
				"amount": "1.0 ETH",
			},
			Description: "Transfer 1.0 ETH to recipient",
		}
	}

	status := "success"
	if receipt.Status == types.ReceiptStatusFailed {
		status = "failed"
	}

	return &ParsedTransaction{
		Hash:           tx.Hash(),
		Type:           "transfer",
		From:           fromInfo,
		To:             toInfo,
		Value:          decimal.NewFromBigInt(tx.Value(), 0),
		MethodCall:     methodCall,
		TokenTransfers: []*TokenTransfer{},
		Events:         []*DecodedEvent{},
		GasAnalysis:    gasAnalysis,
		Status:         status,
		Timestamp:      time.Now(),
	}, nil
}

func (m *MockTransactionParser) DecodeMethodCall(ctx context.Context, data []byte, contractInfo *ContractInfo) (*MethodCall, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("insufficient data for method call")
	}

	// Mock method call decoding
	return &MethodCall{
		Function: &FunctionInfo{
			Name:        "mockMethod",
			Signature:   "mockMethod(uint256)",
			Description: "Mock method call",
		},
		Parameters: map[string]interface{}{
			"value": "123456789",
		},
		DecodedParameters: map[string]string{
			"value": "123.456789",
		},
		Description: "Mock method call with value 123.456789",
	}, nil
}

func (m *MockTransactionParser) AnalyzeGasUsage(tx *types.Transaction, receipt *types.Receipt) *GasAnalysis {
	gasPrice := decimal.NewFromBigInt(tx.GasPrice(), 0)
	gasUsed := decimal.NewFromInt(int64(receipt.GasUsed))
	gasCost := gasPrice.Mul(gasUsed)

	return &GasAnalysis{
		GasLimit:         tx.Gas(),
		GasUsed:          receipt.GasUsed,
		GasPrice:         gasPrice,
		GasCost:          gasCost,
		GasEfficiency:    gasUsed.Div(decimal.NewFromInt(int64(tx.Gas()))),
		EstimatedCostUSD: gasCost.Div(decimal.NewFromInt(1000000000000000000)).Mul(decimal.NewFromFloat(2000)), // Mock ETH price
	}
}

// MockEventDecoder provides mock event decoding
type MockEventDecoder struct{}

func (m *MockEventDecoder) DecodeEvent(ctx context.Context, log *types.Log, contractInfo *ContractInfo) (*DecodedEvent, error) {
	// Mock event decoding
	eventInfo := &EventInfo{
		Name:        "Transfer",
		Signature:   "Transfer(address,address,uint256)",
		Description: "Token transfer event",
		Parameters: []ParameterInfo{
			{Name: "from", Type: "address", Indexed: true},
			{Name: "to", Type: "address", Indexed: true},
			{Name: "value", Type: "uint256"},
		},
	}

	return &DecodedEvent{
		Event: eventInfo,
		Parameters: map[string]interface{}{
			"from":  "0x1111111111111111111111111111111111111111",
			"to":    "0x2222222222222222222222222222222222222222",
			"value": "1000000000000000000",
		},
		DecodedParameters: map[string]string{
			"from":  "0x1111...1111",
			"to":    "0x2222...2222",
			"value": "1.0 ETH",
		},
		Description: "Transfer of 1.0 ETH from 0x1111...1111 to 0x2222...2222",
	}, nil
}

func (m *MockEventDecoder) ResolveEventSignature(topic0 common.Hash) (*EventInfo, error) {
	// Mock event signature resolution
	return &EventInfo{
		Name:        "Transfer",
		Signature:   "Transfer(address,address,uint256)",
		Topic0:      topic0,
		Description: "Token transfer event",
	}, nil
}

// MockValueFormatter provides mock value formatting
type MockValueFormatter struct{}

func (m *MockValueFormatter) FormatValue(value decimal.Decimal, token string) string {
	if token == "ETH" {
		ethValue := value.Div(decimal.NewFromInt(1000000000000000000))
		return fmt.Sprintf("%.4f %s", ethValue.InexactFloat64(), token)
	}
	return fmt.Sprintf("%.2f %s", value.InexactFloat64(), token)
}

func (m *MockValueFormatter) FormatGas(gasUsed uint64, gasPrice decimal.Decimal) string {
	gasCost := gasPrice.Mul(decimal.NewFromInt(int64(gasUsed)))
	ethCost := gasCost.Div(decimal.NewFromInt(1000000000000000000))
	gweiPrice := gasPrice.Div(decimal.NewFromInt(1000000000))
	return fmt.Sprintf("%.6f ETH (%.1f gwei)", ethCost.InexactFloat64(), gweiPrice.InexactFloat64())
}

func (m *MockValueFormatter) FormatTime(timestamp time.Time) string {
	return timestamp.Format("2006-01-02 15:04:05 UTC")
}

func (m *MockValueFormatter) FormatAddress(address common.Address, info *AddressInfo) string {
	if info != nil && info.Label != "" {
		return fmt.Sprintf("%s (%s)", info.Label, address.Hex()[:10]+"...")
	}
	if info != nil && info.ENSName != "" {
		return fmt.Sprintf("%s (%s)", info.ENSName, address.Hex()[:10]+"...")
	}
	return address.Hex()[:10] + "..."
}

// MockTemplateEngine provides mock template rendering
type MockTemplateEngine struct{}

func (m *MockTemplateEngine) RenderTemplate(templateName string, data interface{}) (string, error) {
	// Mock template rendering
	return fmt.Sprintf("Rendered template '%s' with data", templateName), nil
}

func (m *MockTemplateEngine) RegisterTemplate(name string, template string) error {
	// Mock template registration
	return nil
}

func (m *MockTemplateEngine) GetAvailableTemplates() []string {
	return []string{"transaction", "address", "contract", "event"}
}

// MockContextAnalyzer provides mock context analysis
type MockContextAnalyzer struct{}

func (m *MockContextAnalyzer) AnalyzeContext(ctx context.Context, request *TranslationRequest) (*TranslationContext, error) {
	// Mock context analysis
	return &TranslationContext{
		TimeContext: &TimeContext{
			BlockTime:         time.Now().Add(-5 * time.Minute),
			CurrentTime:       time.Now(),
			TimeSinceBlock:    5 * time.Minute,
			NetworkCongestion: "medium",
		},
		MarketContext: &MarketContext{
			ETHPrice:           decimal.NewFromFloat(2000),
			GasPrice:           decimal.NewFromFloat(20000000000),
			NetworkUtilization: decimal.NewFromFloat(0.7),
			MarketSentiment:    "neutral",
		},
	}, nil
}

func (m *MockContextAnalyzer) EnrichContext(ctx context.Context, context *TranslationContext) error {
	// Mock context enrichment
	return nil
}

// MockSentenceBuilder provides mock sentence building
type MockSentenceBuilder struct{}

func (m *MockSentenceBuilder) BuildSummary(data interface{}, context *TranslationContext) string {
	switch v := data.(type) {
	case *ParsedTransaction:
		if !v.Value.IsZero() {
			ethValue := v.Value.Div(decimal.NewFromInt(1000000000000000000))
			return fmt.Sprintf("Transfer of %.4f ETH", ethValue.InexactFloat64())
		}
		return "Smart contract interaction"
	default:
		return "Blockchain operation"
	}
}

func (m *MockSentenceBuilder) BuildDescription(data interface{}, context *TranslationContext) string {
	return "Detailed description of blockchain operation with context"
}

func (m *MockSentenceBuilder) BuildKeyPoints(data interface{}, context *TranslationContext) []string {
	return []string{
		"Transaction completed successfully",
		"Gas fees were reasonable",
		"No security concerns detected",
	}
}

// MockEnrichmentEngine provides mock enrichment
type MockEnrichmentEngine struct{}

func (m *MockEnrichmentEngine) EnrichTranslation(ctx context.Context, result *TranslationResult, request *TranslationRequest) error {
	// Mock enrichment
	result.RelatedItems = []*RelatedItem{
		{
			Type:        "transaction",
			Reference:   "0xabcd1234...",
			Description: "Related transaction",
			Relevance:   decimal.NewFromFloat(0.8),
		},
	}
	return nil
}

func (m *MockEnrichmentEngine) GetPriceData(ctx context.Context, token string, timestamp time.Time) (*PriceData, error) {
	// Mock price data
	return &PriceData{
		Token:     token,
		Price:     decimal.NewFromFloat(2000),
		Currency:  "USD",
		Timestamp: timestamp,
		Source:    "mock_api",
	}, nil
}

func (m *MockEnrichmentEngine) GetRelatedTransactions(ctx context.Context, address common.Address, limit int) ([]common.Hash, error) {
	// Mock related transactions
	var hashes []common.Hash
	for i := 0; i < limit && i < 3; i++ {
		hash := common.HexToHash(fmt.Sprintf("0x%064x", i))
		hashes = append(hashes, hash)
	}
	return hashes, nil
}

// MockMetadataProvider provides mock metadata
type MockMetadataProvider struct{}

func (m *MockMetadataProvider) GetTransactionMetadata(ctx context.Context, hash common.Hash) (*TransactionMetadata, error) {
	return &TransactionMetadata{
		Category:    "transfer",
		Subcategory: "token_transfer",
		RiskLevel:   "low",
		Complexity:  "simple",
		Tags:        []string{"erc20", "transfer"},
	}, nil
}

func (m *MockMetadataProvider) GetAddressMetadata(ctx context.Context, address common.Address) (*AddressMetadata, error) {
	return &AddressMetadata{
		Category:      "user",
		Reputation:    decimal.NewFromFloat(0.8),
		ActivityLevel: "medium",
		Tags:          []string{"verified", "active"},
	}, nil
}

func (m *MockMetadataProvider) GetContractMetadata(ctx context.Context, address common.Address) (*ContractMetadata, error) {
	return &ContractMetadata{
		Category:        "token",
		SecurityScore:   decimal.NewFromFloat(0.9),
		PopularityScore: decimal.NewFromFloat(0.7),
		Tags:            []string{"erc20", "verified", "popular"},
	}, nil
}
