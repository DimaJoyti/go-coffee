package translation

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// BlockchainDataTranslator provides NLP-based translation of blockchain data
type BlockchainDataTranslator struct {
	logger *logger.Logger
	config DataTranslatorConfig

	// Translation components
	addressResolver   AddressResolver
	contractAnalyzer  ContractAnalyzer
	transactionParser TransactionParser
	eventDecoder      EventDecoder
	valueFormatter    ValueFormatter

	// Language processing
	templateEngine  TemplateEngine
	contextAnalyzer ContextAnalyzer
	sentenceBuilder SentenceBuilder

	// Data enrichment
	enrichmentEngine EnrichmentEngine
	metadataProvider MetadataProvider

	// Caching and state
	translationCache map[string]*TranslationResult
	addressCache     map[common.Address]*AddressInfo
	contractCache    map[common.Address]*ContractInfo

	// State management
	isRunning   bool
	cacheTicker *time.Ticker
	stopChan    chan struct{}
	mutex       sync.RWMutex
	cacheMutex  sync.RWMutex
}

// DataTranslatorConfig holds configuration for blockchain data translation
type DataTranslatorConfig struct {
	Enabled                 bool                    `json:"enabled" yaml:"enabled"`
	Language                string                  `json:"language" yaml:"language"`
	DetailLevel             string                  `json:"detail_level" yaml:"detail_level"` // "basic", "detailed", "technical"
	IncludeMetadata         bool                    `json:"include_metadata" yaml:"include_metadata"`
	AddressResolverConfig   AddressResolverConfig   `json:"address_resolver_config" yaml:"address_resolver_config"`
	ContractAnalyzerConfig  ContractAnalyzerConfig  `json:"contract_analyzer_config" yaml:"contract_analyzer_config"`
	TransactionParserConfig TransactionParserConfig `json:"transaction_parser_config" yaml:"transaction_parser_config"`
	EventDecoderConfig      EventDecoderConfig      `json:"event_decoder_config" yaml:"event_decoder_config"`
	ValueFormatterConfig    ValueFormatterConfig    `json:"value_formatter_config" yaml:"value_formatter_config"`
	TemplateEngineConfig    TemplateEngineConfig    `json:"template_engine_config" yaml:"template_engine_config"`
	EnrichmentConfig        EnrichmentEngineConfig  `json:"enrichment_config" yaml:"enrichment_config"`
	CacheConfig             TranslationCacheConfig  `json:"cache_config" yaml:"cache_config"`
}

// Component configurations
type AddressResolverConfig struct {
	Enabled          bool              `json:"enabled" yaml:"enabled"`
	ResolveENS       bool              `json:"resolve_ens" yaml:"resolve_ens"`
	ResolveLabels    bool              `json:"resolve_labels" yaml:"resolve_labels"`
	ResolveContracts bool              `json:"resolve_contracts" yaml:"resolve_contracts"`
	CustomLabels     map[string]string `json:"custom_labels" yaml:"custom_labels"`
	UpdateInterval   time.Duration     `json:"update_interval" yaml:"update_interval"`
}

type ContractAnalyzerConfig struct {
	Enabled         bool `json:"enabled" yaml:"enabled"`
	AnalyzeABI      bool `json:"analyze_abi" yaml:"analyze_abi"`
	DetectStandards bool `json:"detect_standards" yaml:"detect_standards"`
	ResolveProxies  bool `json:"resolve_proxies" yaml:"resolve_proxies"`
	CacheResults    bool `json:"cache_results" yaml:"cache_results"`
}

type TransactionParserConfig struct {
	Enabled          bool `json:"enabled" yaml:"enabled"`
	ParseMethodCalls bool `json:"parse_method_calls" yaml:"parse_method_calls"`
	DecodeInputData  bool `json:"decode_input_data" yaml:"decode_input_data"`
	AnalyzeGasUsage  bool `json:"analyze_gas_usage" yaml:"analyze_gas_usage"`
	DetectPatterns   bool `json:"detect_patterns" yaml:"detect_patterns"`
}

type EventDecoderConfig struct {
	Enabled           bool `json:"enabled" yaml:"enabled"`
	DecodeKnownEvents bool `json:"decode_known_events" yaml:"decode_known_events"`
	ResolveTopics     bool `json:"resolve_topics" yaml:"resolve_topics"`
	IncludeRawData    bool `json:"include_raw_data" yaml:"include_raw_data"`
}

type ValueFormatterConfig struct {
	Enabled               bool   `json:"enabled" yaml:"enabled"`
	Currency              string `json:"currency" yaml:"currency"`
	DecimalPlaces         int    `json:"decimal_places" yaml:"decimal_places"`
	UseThousandsSeparator bool   `json:"use_thousands_separator" yaml:"use_thousands_separator"`
	ShowUSDValue          bool   `json:"show_usd_value" yaml:"show_usd_value"`
}

type TemplateEngineConfig struct {
	Enabled           bool              `json:"enabled" yaml:"enabled"`
	TemplateDirectory string            `json:"template_directory" yaml:"template_directory"`
	CustomTemplates   map[string]string `json:"custom_templates" yaml:"custom_templates"`
	UseMarkdown       bool              `json:"use_markdown" yaml:"use_markdown"`
}

type EnrichmentEngineConfig struct {
	Enabled               bool `json:"enabled" yaml:"enabled"`
	IncludePriceData      bool `json:"include_price_data" yaml:"include_price_data"`
	IncludeHistoricalData bool `json:"include_historical_data" yaml:"include_historical_data"`
	IncludeRelatedTxs     bool `json:"include_related_txs" yaml:"include_related_txs"`
	MaxRelatedTxs         int  `json:"max_related_txs" yaml:"max_related_txs"`
}

type TranslationCacheConfig struct {
	Enabled         bool          `json:"enabled" yaml:"enabled"`
	MaxSize         int           `json:"max_size" yaml:"max_size"`
	TTL             time.Duration `json:"ttl" yaml:"ttl"`
	CleanupInterval time.Duration `json:"cleanup_interval" yaml:"cleanup_interval"`
}

// Data structures

// TranslationRequest represents a request to translate blockchain data
type TranslationRequest struct {
	Type    string              `json:"type"` // "transaction", "block", "address", "contract", "event"
	Data    interface{}         `json:"data"`
	Context *TranslationContext `json:"context"`
	Options *TranslationOptions `json:"options"`
}

// TranslationContext provides context for translation
type TranslationContext struct {
	UserAddress         *common.Address  `json:"user_address"`
	UserPreferences     *UserPreferences `json:"user_preferences"`
	RelatedTransactions []common.Hash    `json:"related_transactions"`
	TimeContext         *TimeContext     `json:"time_context"`
	MarketContext       *MarketContext   `json:"market_context"`
}

// TranslationOptions controls translation behavior
type TranslationOptions struct {
	DetailLevel     string `json:"detail_level"`
	IncludeMetadata bool   `json:"include_metadata"`
	IncludeRelated  bool   `json:"include_related"`
	Format          string `json:"format"` // "text", "markdown", "html"
	Language        string `json:"language"`
}

// TranslationResult contains the translated blockchain data
type TranslationResult struct {
	Summary             string               `json:"summary"`
	DetailedDescription string               `json:"detailed_description"`
	KeyPoints           []string             `json:"key_points"`
	Warnings            []string             `json:"warnings"`
	Recommendations     []string             `json:"recommendations"`
	Metadata            *TranslationMetadata `json:"metadata"`
	RelatedItems        []*RelatedItem       `json:"related_items"`
	Timestamp           time.Time            `json:"timestamp"`
}

// Supporting types
type AddressInfo struct {
	Address     common.Address `json:"address"`
	Label       string         `json:"label"`
	Type        string         `json:"type"` // "eoa", "contract", "exchange", "defi"
	ENSName     string         `json:"ens_name"`
	IsVerified  bool           `json:"is_verified"`
	Tags        []string       `json:"tags"`
	Description string         `json:"description"`
}

type ContractInfo struct {
	Address        common.Address  `json:"address"`
	Name           string          `json:"name"`
	Symbol         string          `json:"symbol"`
	Standard       string          `json:"standard"` // "ERC20", "ERC721", etc.
	IsProxy        bool            `json:"is_proxy"`
	Implementation *common.Address `json:"implementation"`
	ABI            []byte          `json:"abi"`
	Functions      []FunctionInfo  `json:"functions"`
	Events         []EventInfo     `json:"events"`
}

type FunctionInfo struct {
	Name        string          `json:"name"`
	Signature   string          `json:"signature"`
	Selector    []byte          `json:"selector"`
	Description string          `json:"description"`
	Parameters  []ParameterInfo `json:"parameters"`
}

type EventInfo struct {
	Name        string          `json:"name"`
	Signature   string          `json:"signature"`
	Topic0      common.Hash     `json:"topic0"`
	Description string          `json:"description"`
	Parameters  []ParameterInfo `json:"parameters"`
}

type ParameterInfo struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Indexed     bool   `json:"indexed"`
}

type UserPreferences struct {
	Language             string `json:"language"`
	DetailLevel          string `json:"detail_level"`
	Currency             string `json:"currency"`
	TimeZone             string `json:"time_zone"`
	ShowTechnicalDetails bool   `json:"show_technical_details"`
	ShowUSDValues        bool   `json:"show_usd_values"`
}

type TimeContext struct {
	BlockTime         time.Time     `json:"block_time"`
	CurrentTime       time.Time     `json:"current_time"`
	TimeSinceBlock    time.Duration `json:"time_since_block"`
	NetworkCongestion string        `json:"network_congestion"`
}

type MarketContext struct {
	ETHPrice           decimal.Decimal `json:"eth_price"`
	GasPrice           decimal.Decimal `json:"gas_price"`
	NetworkUtilization decimal.Decimal `json:"network_utilization"`
	MarketSentiment    string          `json:"market_sentiment"`
}

type TranslationMetadata struct {
	TranslationType   string          `json:"translation_type"`
	Confidence        decimal.Decimal `json:"confidence"`
	ProcessingTime    time.Duration   `json:"processing_time"`
	DataSources       []string        `json:"data_sources"`
	EnrichmentApplied []string        `json:"enrichment_applied"`
}

type RelatedItem struct {
	Type        string          `json:"type"`
	Reference   string          `json:"reference"`
	Description string          `json:"description"`
	Relevance   decimal.Decimal `json:"relevance"`
}

// Component interfaces
type AddressResolver interface {
	ResolveAddress(ctx context.Context, address common.Address) (*AddressInfo, error)
	ResolveENS(ctx context.Context, address common.Address) (string, error)
	GetAddressLabel(address common.Address) string
}

type ContractAnalyzer interface {
	AnalyzeContract(ctx context.Context, address common.Address) (*ContractInfo, error)
	DetectStandard(ctx context.Context, address common.Address) (string, error)
	ResolveProxy(ctx context.Context, address common.Address) (*common.Address, error)
}

type TransactionParser interface {
	ParseTransaction(ctx context.Context, tx *types.Transaction, receipt *types.Receipt) (*ParsedTransaction, error)
	DecodeMethodCall(ctx context.Context, data []byte, contractInfo *ContractInfo) (*MethodCall, error)
	AnalyzeGasUsage(tx *types.Transaction, receipt *types.Receipt) *GasAnalysis
}

type EventDecoder interface {
	DecodeEvent(ctx context.Context, log *types.Log, contractInfo *ContractInfo) (*DecodedEvent, error)
	ResolveEventSignature(topic0 common.Hash) (*EventInfo, error)
}

type ValueFormatter interface {
	FormatValue(value decimal.Decimal, token string) string
	FormatGas(gasUsed uint64, gasPrice decimal.Decimal) string
	FormatTime(timestamp time.Time) string
	FormatAddress(address common.Address, info *AddressInfo) string
}

type TemplateEngine interface {
	RenderTemplate(templateName string, data interface{}) (string, error)
	RegisterTemplate(name string, template string) error
	GetAvailableTemplates() []string
}

type ContextAnalyzer interface {
	AnalyzeContext(ctx context.Context, request *TranslationRequest) (*TranslationContext, error)
	EnrichContext(ctx context.Context, context *TranslationContext) error
}

type SentenceBuilder interface {
	BuildSummary(data interface{}, context *TranslationContext) string
	BuildDescription(data interface{}, context *TranslationContext) string
	BuildKeyPoints(data interface{}, context *TranslationContext) []string
}

type EnrichmentEngine interface {
	EnrichTranslation(ctx context.Context, result *TranslationResult, request *TranslationRequest) error
	GetPriceData(ctx context.Context, token string, timestamp time.Time) (*PriceData, error)
	GetRelatedTransactions(ctx context.Context, address common.Address, limit int) ([]common.Hash, error)
}

type MetadataProvider interface {
	GetTransactionMetadata(ctx context.Context, hash common.Hash) (*TransactionMetadata, error)
	GetAddressMetadata(ctx context.Context, address common.Address) (*AddressMetadata, error)
	GetContractMetadata(ctx context.Context, address common.Address) (*ContractMetadata, error)
}

// Supporting types for parsing
type ParsedTransaction struct {
	Hash           common.Hash      `json:"hash"`
	Type           string           `json:"type"` // "transfer", "swap", "mint", etc.
	From           *AddressInfo     `json:"from"`
	To             *AddressInfo     `json:"to"`
	Value          decimal.Decimal  `json:"value"`
	MethodCall     *MethodCall      `json:"method_call"`
	TokenTransfers []*TokenTransfer `json:"token_transfers"`
	Events         []*DecodedEvent  `json:"events"`
	GasAnalysis    *GasAnalysis     `json:"gas_analysis"`
	Status         string           `json:"status"`
	Timestamp      time.Time        `json:"timestamp"`
}

type MethodCall struct {
	Function          *FunctionInfo          `json:"function"`
	Parameters        map[string]interface{} `json:"parameters"`
	DecodedParameters map[string]string      `json:"decoded_parameters"`
	Description       string                 `json:"description"`
}

type TokenTransfer struct {
	From     common.Address  `json:"from"`
	To       common.Address  `json:"to"`
	Token    *TokenInfo      `json:"token"`
	Amount   decimal.Decimal `json:"amount"`
	USDValue decimal.Decimal `json:"usd_value"`
}

type TokenInfo struct {
	Address  common.Address `json:"address"`
	Name     string         `json:"name"`
	Symbol   string         `json:"symbol"`
	Decimals uint8          `json:"decimals"`
	Standard string         `json:"standard"`
}

type DecodedEvent struct {
	Event             *EventInfo             `json:"event"`
	Parameters        map[string]interface{} `json:"parameters"`
	DecodedParameters map[string]string      `json:"decoded_parameters"`
	Description       string                 `json:"description"`
}

type GasAnalysis struct {
	GasLimit         uint64          `json:"gas_limit"`
	GasUsed          uint64          `json:"gas_used"`
	GasPrice         decimal.Decimal `json:"gas_price"`
	GasCost          decimal.Decimal `json:"gas_cost"`
	GasEfficiency    decimal.Decimal `json:"gas_efficiency"`
	EstimatedCostUSD decimal.Decimal `json:"estimated_cost_usd"`
}

type PriceData struct {
	Token     string          `json:"token"`
	Price     decimal.Decimal `json:"price"`
	Currency  string          `json:"currency"`
	Timestamp time.Time       `json:"timestamp"`
	Source    string          `json:"source"`
}

type TransactionMetadata struct {
	Category    string   `json:"category"`
	Subcategory string   `json:"subcategory"`
	RiskLevel   string   `json:"risk_level"`
	Complexity  string   `json:"complexity"`
	Tags        []string `json:"tags"`
}

type AddressMetadata struct {
	Category      string          `json:"category"`
	Reputation    decimal.Decimal `json:"reputation"`
	ActivityLevel string          `json:"activity_level"`
	Tags          []string        `json:"tags"`
}

type ContractMetadata struct {
	Category        string          `json:"category"`
	SecurityScore   decimal.Decimal `json:"security_score"`
	PopularityScore decimal.Decimal `json:"popularity_score"`
	Tags            []string        `json:"tags"`
}

// NewBlockchainDataTranslator creates a new blockchain data translator
func NewBlockchainDataTranslator(logger *logger.Logger, config DataTranslatorConfig) *BlockchainDataTranslator {
	bdt := &BlockchainDataTranslator{
		logger:           logger.Named("blockchain-data-translator"),
		config:           config,
		translationCache: make(map[string]*TranslationResult),
		addressCache:     make(map[common.Address]*AddressInfo),
		contractCache:    make(map[common.Address]*ContractInfo),
		stopChan:         make(chan struct{}),
	}

	// Initialize components (mock implementations for this example)
	bdt.initializeComponents()

	return bdt
}

// initializeComponents initializes all translation components
func (bdt *BlockchainDataTranslator) initializeComponents() {
	// Initialize components with mock implementations
	// In production, these would be real implementations
	bdt.addressResolver = &MockAddressResolver{}
	bdt.contractAnalyzer = &MockContractAnalyzer{}
	bdt.transactionParser = &MockTransactionParser{}
	bdt.eventDecoder = &MockEventDecoder{}
	bdt.valueFormatter = &MockValueFormatter{}
	bdt.templateEngine = &MockTemplateEngine{}
	bdt.contextAnalyzer = &MockContextAnalyzer{}
	bdt.sentenceBuilder = &MockSentenceBuilder{}
	bdt.enrichmentEngine = &MockEnrichmentEngine{}
	bdt.metadataProvider = &MockMetadataProvider{}
}

// Start starts the blockchain data translator
func (bdt *BlockchainDataTranslator) Start(ctx context.Context) error {
	bdt.mutex.Lock()
	defer bdt.mutex.Unlock()

	if bdt.isRunning {
		return fmt.Errorf("blockchain data translator is already running")
	}

	if !bdt.config.Enabled {
		bdt.logger.Info("Blockchain data translator is disabled")
		return nil
	}

	bdt.logger.Info("Starting blockchain data translator",
		zap.String("language", bdt.config.Language),
		zap.String("detail_level", bdt.config.DetailLevel))

	// Start cache cleanup routine
	if bdt.config.CacheConfig.Enabled {
		bdt.cacheTicker = time.NewTicker(bdt.config.CacheConfig.CleanupInterval)
		go bdt.cacheCleanupLoop(ctx)
	}

	bdt.isRunning = true
	bdt.logger.Info("Blockchain data translator started successfully")
	return nil
}

// Stop stops the blockchain data translator
func (bdt *BlockchainDataTranslator) Stop() error {
	bdt.mutex.Lock()
	defer bdt.mutex.Unlock()

	if !bdt.isRunning {
		return nil
	}

	bdt.logger.Info("Stopping blockchain data translator")

	// Stop cache cleanup
	if bdt.cacheTicker != nil {
		bdt.cacheTicker.Stop()
	}
	close(bdt.stopChan)

	bdt.isRunning = false
	bdt.logger.Info("Blockchain data translator stopped")
	return nil
}

// TranslateTransaction translates a transaction into human-readable text
func (bdt *BlockchainDataTranslator) TranslateTransaction(ctx context.Context, tx *types.Transaction, receipt *types.Receipt, options *TranslationOptions) (*TranslationResult, error) {
	startTime := time.Now()

	bdt.logger.Debug("Translating transaction",
		zap.String("hash", tx.Hash().Hex()),
		zap.String("detail_level", options.DetailLevel))

	// Check cache first
	cacheKey := fmt.Sprintf("tx:%s:%s", tx.Hash().Hex(), options.DetailLevel)
	if cached := bdt.getFromCache(cacheKey); cached != nil {
		return cached, nil
	}

	// Parse transaction
	parsedTx, err := bdt.transactionParser.ParseTransaction(ctx, tx, receipt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse transaction: %w", err)
	}

	// Build translation context
	context := &TranslationContext{
		TimeContext: &TimeContext{
			BlockTime:   time.Unix(int64(receipt.BlockNumber.Uint64()*13), 0), // Mock block time
			CurrentTime: time.Now(),
		},
		MarketContext: &MarketContext{
			ETHPrice: decimal.NewFromFloat(2000), // Mock ETH price
			GasPrice: decimal.NewFromBigInt(tx.GasPrice(), 0),
		},
	}

	// Build translation result
	result := &TranslationResult{
		Timestamp: time.Now(),
		Metadata: &TranslationMetadata{
			TranslationType: "transaction",
			ProcessingTime:  time.Since(startTime),
			DataSources:     []string{"blockchain", "contract_abi", "address_labels"},
		},
	}

	// Generate summary
	result.Summary = bdt.sentenceBuilder.BuildSummary(parsedTx, context)

	// Generate detailed description based on detail level
	switch options.DetailLevel {
	case "basic":
		result.DetailedDescription = bdt.buildBasicTransactionDescription(parsedTx, context)
	case "detailed":
		result.DetailedDescription = bdt.buildDetailedTransactionDescription(parsedTx, context)
	case "technical":
		result.DetailedDescription = bdt.buildTechnicalTransactionDescription(parsedTx, context)
	default:
		result.DetailedDescription = bdt.buildDetailedTransactionDescription(parsedTx, context)
	}

	// Generate key points
	result.KeyPoints = bdt.sentenceBuilder.BuildKeyPoints(parsedTx, context)

	// Add warnings and recommendations
	result.Warnings = bdt.generateTransactionWarnings(parsedTx)
	result.Recommendations = bdt.generateTransactionRecommendations(parsedTx)

	// Enrich with additional data if requested
	if options.IncludeMetadata {
		if err := bdt.enrichmentEngine.EnrichTranslation(ctx, result, &TranslationRequest{
			Type:    "transaction",
			Data:    parsedTx,
			Options: options,
		}); err != nil {
			bdt.logger.Warn("Failed to enrich translation", zap.Error(err))
		}
	}

	// Cache the result
	bdt.addToCache(cacheKey, result)

	bdt.logger.Debug("Transaction translated successfully",
		zap.String("hash", tx.Hash().Hex()),
		zap.Duration("processing_time", time.Since(startTime)))

	return result, nil
}

// TranslateAddress translates an address into human-readable information
func (bdt *BlockchainDataTranslator) TranslateAddress(ctx context.Context, address common.Address, options *TranslationOptions) (*TranslationResult, error) {
	startTime := time.Now()

	bdt.logger.Debug("Translating address",
		zap.String("address", address.Hex()),
		zap.String("detail_level", options.DetailLevel))

	// Check cache first
	cacheKey := fmt.Sprintf("addr:%s:%s", address.Hex(), options.DetailLevel)
	if cached := bdt.getFromCache(cacheKey); cached != nil {
		return cached, nil
	}

	// Resolve address information
	addressInfo, err := bdt.addressResolver.ResolveAddress(ctx, address)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve address: %w", err)
	}

	// Build translation result
	result := &TranslationResult{
		Timestamp: time.Now(),
		Metadata: &TranslationMetadata{
			TranslationType: "address",
			ProcessingTime:  time.Since(startTime),
			DataSources:     []string{"address_resolver", "ens", "labels"},
		},
	}

	// Generate summary
	result.Summary = bdt.buildAddressSummary(addressInfo)

	// Generate detailed description
	result.DetailedDescription = bdt.buildAddressDescription(addressInfo, options.DetailLevel)

	// Generate key points
	result.KeyPoints = bdt.buildAddressKeyPoints(addressInfo)

	// Cache the result
	bdt.addToCache(cacheKey, result)

	return result, nil
}

// TranslateContract translates a smart contract into human-readable information
func (bdt *BlockchainDataTranslator) TranslateContract(ctx context.Context, address common.Address, options *TranslationOptions) (*TranslationResult, error) {
	startTime := time.Now()

	// Check cache first
	cacheKey := fmt.Sprintf("contract:%s:%s", address.Hex(), options.DetailLevel)
	if cached := bdt.getFromCache(cacheKey); cached != nil {
		return cached, nil
	}

	// Analyze contract
	contractInfo, err := bdt.contractAnalyzer.AnalyzeContract(ctx, address)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze contract: %w", err)
	}

	// Build translation result
	result := &TranslationResult{
		Timestamp: time.Now(),
		Metadata: &TranslationMetadata{
			TranslationType: "contract",
			ProcessingTime:  time.Since(startTime),
			DataSources:     []string{"contract_analyzer", "abi", "bytecode"},
		},
	}

	// Generate summary
	result.Summary = bdt.buildContractSummary(contractInfo)

	// Generate detailed description
	result.DetailedDescription = bdt.buildContractDescription(contractInfo, options.DetailLevel)

	// Generate key points
	result.KeyPoints = bdt.buildContractKeyPoints(contractInfo)

	// Cache the result
	bdt.addToCache(cacheKey, result)

	return result, nil
}

// Helper methods for building descriptions

func (bdt *BlockchainDataTranslator) buildBasicTransactionDescription(tx *ParsedTransaction, context *TranslationContext) string {
	var parts []string

	// Basic transaction info
	if tx.From != nil && tx.To != nil {
		fromLabel := bdt.valueFormatter.FormatAddress(tx.From.Address, tx.From)
		toLabel := bdt.valueFormatter.FormatAddress(tx.To.Address, tx.To)
		parts = append(parts, fmt.Sprintf("Transaction from %s to %s", fromLabel, toLabel))
	}

	// Value transfer
	if !tx.Value.IsZero() {
		valueStr := bdt.valueFormatter.FormatValue(tx.Value, "ETH")
		parts = append(parts, fmt.Sprintf("transferring %s", valueStr))
	}

	// Gas cost
	if tx.GasAnalysis != nil {
		gasCost := bdt.valueFormatter.FormatGas(tx.GasAnalysis.GasUsed, tx.GasAnalysis.GasPrice)
		parts = append(parts, fmt.Sprintf("with gas cost of %s", gasCost))
	}

	// Status
	parts = append(parts, fmt.Sprintf("Status: %s", tx.Status))

	return strings.Join(parts, ". ") + "."
}

func (bdt *BlockchainDataTranslator) buildDetailedTransactionDescription(tx *ParsedTransaction, context *TranslationContext) string {
	var parts []string

	// Add basic description
	parts = append(parts, bdt.buildBasicTransactionDescription(tx, context))

	// Add method call information
	if tx.MethodCall != nil {
		parts = append(parts, fmt.Sprintf("Called method: %s", tx.MethodCall.Function.Name))
		if tx.MethodCall.Description != "" {
			parts = append(parts, tx.MethodCall.Description)
		}
	}

	// Add token transfers
	if len(tx.TokenTransfers) > 0 {
		parts = append(parts, "Token transfers:")
		for _, transfer := range tx.TokenTransfers {
			transferDesc := fmt.Sprintf("- %s %s from %s to %s",
				bdt.valueFormatter.FormatValue(transfer.Amount, transfer.Token.Symbol),
				transfer.Token.Symbol,
				transfer.From.Hex()[:10]+"...",
				transfer.To.Hex()[:10]+"...")
			parts = append(parts, transferDesc)
		}
	}

	// Add events
	if len(tx.Events) > 0 {
		parts = append(parts, "Events emitted:")
		for _, event := range tx.Events {
			eventDesc := fmt.Sprintf("- %s", event.Event.Name)
			if event.Description != "" {
				eventDesc += ": " + event.Description
			}
			parts = append(parts, eventDesc)
		}
	}

	return strings.Join(parts, "\n")
}

func (bdt *BlockchainDataTranslator) buildTechnicalTransactionDescription(tx *ParsedTransaction, context *TranslationContext) string {
	var parts []string

	// Add detailed description
	parts = append(parts, bdt.buildDetailedTransactionDescription(tx, context))

	// Add technical details
	parts = append(parts, "\nTechnical Details:")
	parts = append(parts, fmt.Sprintf("- Transaction Hash: %s", tx.Hash.Hex()))
	parts = append(parts, fmt.Sprintf("- Block Timestamp: %s", bdt.valueFormatter.FormatTime(tx.Timestamp)))

	// Gas analysis
	if tx.GasAnalysis != nil {
		parts = append(parts, fmt.Sprintf("- Gas Limit: %d", tx.GasAnalysis.GasLimit))
		parts = append(parts, fmt.Sprintf("- Gas Used: %d (%.1f%%)",
			tx.GasAnalysis.GasUsed,
			tx.GasAnalysis.GasEfficiency.Mul(decimal.NewFromFloat(100)).InexactFloat64()))
		parts = append(parts, fmt.Sprintf("- Gas Price: %s gwei",
			tx.GasAnalysis.GasPrice.Div(decimal.NewFromInt(1000000000)).StringFixed(2)))
	}

	// Method call details
	if tx.MethodCall != nil {
		parts = append(parts, fmt.Sprintf("- Function Signature: %s", tx.MethodCall.Function.Signature))
		if len(tx.MethodCall.Parameters) > 0 {
			parts = append(parts, "- Parameters:")
			for name, value := range tx.MethodCall.DecodedParameters {
				parts = append(parts, fmt.Sprintf("  - %s: %s", name, value))
			}
		}
	}

	return strings.Join(parts, "\n")
}

func (bdt *BlockchainDataTranslator) buildAddressSummary(info *AddressInfo) string {
	if info.Label != "" {
		return fmt.Sprintf("%s (%s)", info.Label, info.Type)
	}
	if info.ENSName != "" {
		return fmt.Sprintf("%s (%s)", info.ENSName, info.Type)
	}
	return fmt.Sprintf("%s address", strings.Title(info.Type))
}

func (bdt *BlockchainDataTranslator) buildAddressDescription(info *AddressInfo, detailLevel string) string {
	var parts []string

	parts = append(parts, fmt.Sprintf("Address: %s", info.Address.Hex()))

	if info.Label != "" {
		parts = append(parts, fmt.Sprintf("Label: %s", info.Label))
	}

	if info.ENSName != "" {
		parts = append(parts, fmt.Sprintf("ENS Name: %s", info.ENSName))
	}

	parts = append(parts, fmt.Sprintf("Type: %s", strings.Title(info.Type)))

	if info.IsVerified {
		parts = append(parts, "✓ Verified")
	}

	if len(info.Tags) > 0 {
		parts = append(parts, fmt.Sprintf("Tags: %s", strings.Join(info.Tags, ", ")))
	}

	if info.Description != "" && detailLevel != "basic" {
		parts = append(parts, fmt.Sprintf("Description: %s", info.Description))
	}

	return strings.Join(parts, "\n")
}

func (bdt *BlockchainDataTranslator) buildAddressKeyPoints(info *AddressInfo) []string {
	var points []string

	if info.Type == "contract" {
		points = append(points, "This is a smart contract address")
	} else {
		points = append(points, "This is an externally owned account (EOA)")
	}

	if info.IsVerified {
		points = append(points, "Address is verified and trusted")
	}

	if len(info.Tags) > 0 {
		for _, tag := range info.Tags {
			if tag == "exchange" {
				points = append(points, "This address belongs to a cryptocurrency exchange")
			} else if tag == "defi" {
				points = append(points, "This address is associated with DeFi protocols")
			}
		}
	}

	return points
}

func (bdt *BlockchainDataTranslator) buildContractSummary(info *ContractInfo) string {
	if info.Name != "" {
		if info.Symbol != "" {
			return fmt.Sprintf("%s (%s) - %s contract", info.Name, info.Symbol, info.Standard)
		}
		return fmt.Sprintf("%s - %s contract", info.Name, info.Standard)
	}
	return fmt.Sprintf("%s smart contract", info.Standard)
}

func (bdt *BlockchainDataTranslator) buildContractDescription(info *ContractInfo, detailLevel string) string {
	var parts []string

	parts = append(parts, fmt.Sprintf("Contract Address: %s", info.Address.Hex()))

	if info.Name != "" {
		parts = append(parts, fmt.Sprintf("Name: %s", info.Name))
	}

	if info.Symbol != "" {
		parts = append(parts, fmt.Sprintf("Symbol: %s", info.Symbol))
	}

	if info.Standard != "" {
		parts = append(parts, fmt.Sprintf("Standard: %s", info.Standard))
	}

	if info.IsProxy {
		parts = append(parts, "⚠️ This is a proxy contract")
		if info.Implementation != nil {
			parts = append(parts, fmt.Sprintf("Implementation: %s", info.Implementation.Hex()))
		}
	}

	if detailLevel == "technical" {
		if len(info.Functions) > 0 {
			parts = append(parts, fmt.Sprintf("Functions: %d", len(info.Functions)))
		}
		if len(info.Events) > 0 {
			parts = append(parts, fmt.Sprintf("Events: %d", len(info.Events)))
		}
	}

	return strings.Join(parts, "\n")
}

func (bdt *BlockchainDataTranslator) buildContractKeyPoints(info *ContractInfo) []string {
	var points []string

	if info.Standard != "" {
		switch info.Standard {
		case "ERC20":
			points = append(points, "This is a fungible token contract")
		case "ERC721":
			points = append(points, "This is an NFT (Non-Fungible Token) contract")
		case "ERC1155":
			points = append(points, "This is a multi-token contract")
		}
	}

	if info.IsProxy {
		points = append(points, "Contract uses proxy pattern for upgradability")
	}

	return points
}

func (bdt *BlockchainDataTranslator) generateTransactionWarnings(tx *ParsedTransaction) []string {
	var warnings []string

	// Check for failed transaction
	if tx.Status == "failed" {
		warnings = append(warnings, "⚠️ This transaction failed and consumed gas")
	}

	// Check for high gas usage
	if tx.GasAnalysis != nil && tx.GasAnalysis.GasEfficiency.LessThan(decimal.NewFromFloat(0.5)) {
		warnings = append(warnings, "⚠️ This transaction used a high amount of gas")
	}

	// Check for contract interaction
	if tx.To != nil && tx.To.Type == "contract" && tx.MethodCall == nil {
		warnings = append(warnings, "⚠️ Contract interaction without recognized method call")
	}

	return warnings
}

func (bdt *BlockchainDataTranslator) generateTransactionRecommendations(tx *ParsedTransaction) []string {
	var recommendations []string

	// Gas optimization recommendations
	if tx.GasAnalysis != nil {
		if tx.GasAnalysis.GasPrice.GreaterThan(decimal.NewFromFloat(50000000000)) { // > 50 gwei
			recommendations = append(recommendations, "💡 Consider using lower gas price during off-peak hours")
		}
	}

	// Security recommendations
	if tx.To != nil && tx.To.Type == "contract" && !tx.To.IsVerified {
		recommendations = append(recommendations, "🔒 Be cautious when interacting with unverified contracts")
	}

	return recommendations
}

// Cache management methods

func (bdt *BlockchainDataTranslator) getFromCache(key string) *TranslationResult {
	if !bdt.config.CacheConfig.Enabled {
		return nil
	}

	bdt.cacheMutex.RLock()
	defer bdt.cacheMutex.RUnlock()

	if result, exists := bdt.translationCache[key]; exists {
		// Check if cache entry is still valid
		if time.Since(result.Timestamp) < bdt.config.CacheConfig.TTL {
			return result
		}
		// Remove expired entry
		delete(bdt.translationCache, key)
	}

	return nil
}

func (bdt *BlockchainDataTranslator) addToCache(key string, result *TranslationResult) {
	if !bdt.config.CacheConfig.Enabled {
		return
	}

	bdt.cacheMutex.Lock()
	defer bdt.cacheMutex.Unlock()

	// Check cache size limit
	if len(bdt.translationCache) >= bdt.config.CacheConfig.MaxSize {
		// Remove oldest entries (simple FIFO for this example)
		for k := range bdt.translationCache {
			delete(bdt.translationCache, k)
			break
		}
	}

	bdt.translationCache[key] = result
}

func (bdt *BlockchainDataTranslator) cacheCleanupLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-bdt.stopChan:
			return
		case <-bdt.cacheTicker.C:
			bdt.cleanupExpiredCache()
		}
	}
}

func (bdt *BlockchainDataTranslator) cleanupExpiredCache() {
	bdt.cacheMutex.Lock()
	defer bdt.cacheMutex.Unlock()

	now := time.Now()
	for key, result := range bdt.translationCache {
		if now.Sub(result.Timestamp) > bdt.config.CacheConfig.TTL {
			delete(bdt.translationCache, key)
		}
	}
}

// IsRunning returns whether the translator is running
func (bdt *BlockchainDataTranslator) IsRunning() bool {
	bdt.mutex.RLock()
	defer bdt.mutex.RUnlock()
	return bdt.isRunning
}

// GetMetrics returns translator metrics
func (bdt *BlockchainDataTranslator) GetMetrics() map[string]interface{} {
	bdt.cacheMutex.RLock()
	defer bdt.cacheMutex.RUnlock()

	return map[string]interface{}{
		"is_running":                bdt.IsRunning(),
		"cache_size":                len(bdt.translationCache),
		"address_cache_size":        len(bdt.addressCache),
		"contract_cache_size":       len(bdt.contractCache),
		"language":                  bdt.config.Language,
		"detail_level":              bdt.config.DetailLevel,
		"cache_enabled":             bdt.config.CacheConfig.Enabled,
		"enrichment_enabled":        bdt.config.EnrichmentConfig.Enabled,
		"address_resolver_enabled":  bdt.config.AddressResolverConfig.Enabled,
		"contract_analyzer_enabled": bdt.config.ContractAnalyzerConfig.Enabled,
	}
}
