package translation

import (
	"fmt"
	"time"
)

// GetDefaultDataTranslatorConfig returns default data translator configuration
func GetDefaultDataTranslatorConfig() DataTranslatorConfig {
	return DataTranslatorConfig{
		Enabled:         true,
		Language:        "en",
		DetailLevel:     "detailed",
		IncludeMetadata: true,
		AddressResolverConfig: AddressResolverConfig{
			Enabled:        true,
			ResolveENS:     true,
			ResolveLabels:  true,
			ResolveContracts: true,
			CustomLabels: map[string]string{
				"0x0000000000000000000000000000000000000000": "Null Address",
				"0x000000000000000000000000000000000000dead": "Burn Address",
			},
			UpdateInterval: 1 * time.Hour,
		},
		ContractAnalyzerConfig: ContractAnalyzerConfig{
			Enabled:        true,
			AnalyzeABI:     true,
			DetectStandards: true,
			ResolveProxies: true,
			CacheResults:   true,
		},
		TransactionParserConfig: TransactionParserConfig{
			Enabled:         true,
			ParseMethodCalls: true,
			DecodeInputData: true,
			AnalyzeGasUsage: true,
			DetectPatterns:  true,
		},
		EventDecoderConfig: EventDecoderConfig{
			Enabled:         true,
			DecodeKnownEvents: true,
			ResolveTopics:   true,
			IncludeRawData:  false,
		},
		ValueFormatterConfig: ValueFormatterConfig{
			Enabled:              true,
			Currency:             "USD",
			DecimalPlaces:        4,
			UseThousandsSeparator: true,
			ShowUSDValue:         true,
		},
		TemplateEngineConfig: TemplateEngineConfig{
			Enabled:           true,
			TemplateDirectory: "./templates",
			CustomTemplates: map[string]string{
				"simple_transfer": "{{.From}} sent {{.Value}} to {{.To}}",
				"contract_call":   "{{.From}} called {{.Method}} on {{.Contract}}",
			},
			UseMarkdown: true,
		},
		EnrichmentConfig: EnrichmentEngineConfig{
			Enabled:              true,
			IncludePriceData:     true,
			IncludeHistoricalData: false,
			IncludeRelatedTxs:    true,
			MaxRelatedTxs:        5,
		},
		CacheConfig: TranslationCacheConfig{
			Enabled:         true,
			MaxSize:         10000,
			TTL:             1 * time.Hour,
			CleanupInterval: 15 * time.Minute,
		},
	}
}

// GetBasicDataTranslatorConfig returns configuration for basic translation
func GetBasicDataTranslatorConfig() DataTranslatorConfig {
	config := GetDefaultDataTranslatorConfig()
	
	// Simplify for basic use
	config.DetailLevel = "basic"
	config.IncludeMetadata = false
	
	// Disable advanced features
	config.ContractAnalyzerConfig.AnalyzeABI = false
	config.TransactionParserConfig.DecodeInputData = false
	config.EventDecoderConfig.DecodeKnownEvents = false
	config.EnrichmentConfig.IncludePriceData = false
	config.EnrichmentConfig.IncludeRelatedTxs = false
	
	// Smaller cache
	config.CacheConfig.MaxSize = 1000
	config.CacheConfig.TTL = 30 * time.Minute
	
	return config
}

// GetTechnicalDataTranslatorConfig returns configuration for technical users
func GetTechnicalDataTranslatorConfig() DataTranslatorConfig {
	config := GetDefaultDataTranslatorConfig()
	
	// Technical detail level
	config.DetailLevel = "technical"
	config.IncludeMetadata = true
	
	// Enable all analysis features
	config.ContractAnalyzerConfig.AnalyzeABI = true
	config.ContractAnalyzerConfig.ResolveProxies = true
	config.TransactionParserConfig.ParseMethodCalls = true
	config.TransactionParserConfig.DecodeInputData = true
	config.TransactionParserConfig.DetectPatterns = true
	config.EventDecoderConfig.DecodeKnownEvents = true
	config.EventDecoderConfig.IncludeRawData = true
	
	// Enhanced enrichment
	config.EnrichmentConfig.IncludePriceData = true
	config.EnrichmentConfig.IncludeHistoricalData = true
	config.EnrichmentConfig.IncludeRelatedTxs = true
	config.EnrichmentConfig.MaxRelatedTxs = 10
	
	// Larger cache for technical users
	config.CacheConfig.MaxSize = 50000
	config.CacheConfig.TTL = 4 * time.Hour
	
	return config
}

// GetMultiLanguageConfig returns configuration for multiple languages
func GetMultiLanguageConfig(language string) DataTranslatorConfig {
	config := GetDefaultDataTranslatorConfig()
	
	// Set language
	config.Language = language
	
	// Language-specific customizations
	switch language {
	case "es": // Spanish
		config.ValueFormatterConfig.Currency = "EUR"
		config.TemplateEngineConfig.CustomTemplates = map[string]string{
			"simple_transfer": "{{.From}} envió {{.Value}} a {{.To}}",
			"contract_call":   "{{.From}} llamó {{.Method}} en {{.Contract}}",
		}
	case "fr": // French
		config.ValueFormatterConfig.Currency = "EUR"
		config.TemplateEngineConfig.CustomTemplates = map[string]string{
			"simple_transfer": "{{.From}} a envoyé {{.Value}} à {{.To}}",
			"contract_call":   "{{.From}} a appelé {{.Method}} sur {{.Contract}}",
		}
	case "de": // German
		config.ValueFormatterConfig.Currency = "EUR"
		config.TemplateEngineConfig.CustomTemplates = map[string]string{
			"simple_transfer": "{{.From}} hat {{.Value}} an {{.To}} gesendet",
			"contract_call":   "{{.From}} hat {{.Method}} auf {{.Contract}} aufgerufen",
		}
	case "ja": // Japanese
		config.ValueFormatterConfig.Currency = "JPY"
		config.TemplateEngineConfig.CustomTemplates = map[string]string{
			"simple_transfer": "{{.From}}が{{.To}}に{{.Value}}を送信しました",
			"contract_call":   "{{.From}}が{{.Contract}}で{{.Method}}を呼び出しました",
		}
	case "zh": // Chinese
		config.ValueFormatterConfig.Currency = "CNY"
		config.TemplateEngineConfig.CustomTemplates = map[string]string{
			"simple_transfer": "{{.From}}向{{.To}}发送了{{.Value}}",
			"contract_call":   "{{.From}}在{{.Contract}}上调用了{{.Method}}",
		}
	}
	
	return config
}

// ValidateDataTranslatorConfig validates data translator configuration
func ValidateDataTranslatorConfig(config DataTranslatorConfig) error {
	if !config.Enabled {
		return nil
	}
	
	// Validate language
	supportedLanguages := GetSupportedLanguages()
	isValidLanguage := false
	for _, lang := range supportedLanguages {
		if config.Language == lang {
			isValidLanguage = true
			break
		}
	}
	if !isValidLanguage {
		return fmt.Errorf("unsupported language: %s", config.Language)
	}
	
	// Validate detail level
	supportedDetailLevels := GetSupportedDetailLevels()
	isValidDetailLevel := false
	for _, level := range supportedDetailLevels {
		if config.DetailLevel == level {
			isValidDetailLevel = true
			break
		}
	}
	if !isValidDetailLevel {
		return fmt.Errorf("unsupported detail level: %s", config.DetailLevel)
	}
	
	// Validate address resolver config
	if config.AddressResolverConfig.Enabled {
		if config.AddressResolverConfig.UpdateInterval <= 0 {
			return fmt.Errorf("address resolver update interval must be positive")
		}
	}
	
	// Validate value formatter config
	if config.ValueFormatterConfig.Enabled {
		if config.ValueFormatterConfig.DecimalPlaces < 0 {
			return fmt.Errorf("decimal places must be non-negative")
		}
		
		supportedCurrencies := GetSupportedCurrencies()
		isValidCurrency := false
		for _, currency := range supportedCurrencies {
			if config.ValueFormatterConfig.Currency == currency {
				isValidCurrency = true
				break
			}
		}
		if !isValidCurrency {
			return fmt.Errorf("unsupported currency: %s", config.ValueFormatterConfig.Currency)
		}
	}
	
	// Validate enrichment config
	if config.EnrichmentConfig.Enabled {
		if config.EnrichmentConfig.MaxRelatedTxs < 0 {
			return fmt.Errorf("max related transactions must be non-negative")
		}
		if config.EnrichmentConfig.MaxRelatedTxs > 100 {
			return fmt.Errorf("max related transactions cannot exceed 100")
		}
	}
	
	// Validate cache config
	if config.CacheConfig.Enabled {
		if config.CacheConfig.MaxSize <= 0 {
			return fmt.Errorf("cache max size must be positive")
		}
		if config.CacheConfig.TTL <= 0 {
			return fmt.Errorf("cache TTL must be positive")
		}
		if config.CacheConfig.CleanupInterval <= 0 {
			return fmt.Errorf("cache cleanup interval must be positive")
		}
	}
	
	return nil
}

// GetSupportedLanguages returns supported languages
func GetSupportedLanguages() []string {
	return []string{
		"en", // English
		"es", // Spanish
		"fr", // French
		"de", // German
		"it", // Italian
		"pt", // Portuguese
		"ru", // Russian
		"ja", // Japanese
		"ko", // Korean
		"zh", // Chinese
		"ar", // Arabic
		"hi", // Hindi
	}
}

// GetSupportedDetailLevels returns supported detail levels
func GetSupportedDetailLevels() []string {
	return []string{
		"basic",
		"detailed",
		"technical",
	}
}

// GetSupportedCurrencies returns supported currencies
func GetSupportedCurrencies() []string {
	return []string{
		"USD", "EUR", "GBP", "JPY", "CNY", "KRW", "CAD", "AUD", "CHF", "SEK",
		"NOK", "DKK", "PLN", "CZK", "HUF", "RUB", "BRL", "MXN", "INR", "SGD",
	}
}

// GetSupportedFormats returns supported output formats
func GetSupportedFormats() []string {
	return []string{
		"text",
		"markdown",
		"html",
		"json",
	}
}

// GetOptimalConfigForUseCase returns optimal configuration for specific use cases
func GetOptimalConfigForUseCase(useCase string) (DataTranslatorConfig, error) {
	switch useCase {
	case "basic":
		return GetBasicDataTranslatorConfig(), nil
	case "technical":
		return GetTechnicalDataTranslatorConfig(), nil
	case "default":
		return GetDefaultDataTranslatorConfig(), nil
	default:
		return DataTranslatorConfig{}, fmt.Errorf("unsupported use case: %s", useCase)
	}
}

// GetLanguageDescription returns descriptions for supported languages
func GetLanguageDescription() map[string]string {
	return map[string]string{
		"en": "English - Default language with full feature support",
		"es": "Spanish - Español con soporte completo de características",
		"fr": "French - Français avec support complet des fonctionnalités",
		"de": "German - Deutsch mit vollständiger Funktionsunterstützung",
		"it": "Italian - Italiano con supporto completo delle funzionalità",
		"pt": "Portuguese - Português com suporte completo de recursos",
		"ru": "Russian - Русский с полной поддержкой функций",
		"ja": "Japanese - 日本語、完全な機能サポート付き",
		"ko": "Korean - 한국어, 전체 기능 지원",
		"zh": "Chinese - 中文，具有完整的功能支持",
		"ar": "Arabic - العربية مع دعم كامل للميزات",
		"hi": "Hindi - हिंदी पूर्ण सुविधा समर्थन के साथ",
	}
}

// GetDetailLevelDescription returns descriptions for detail levels
func GetDetailLevelDescription() map[string]string {
	return map[string]string{
		"basic":     "Basic information with simple explanations suitable for beginners",
		"detailed":  "Comprehensive information with context and explanations for general users",
		"technical": "Technical details with raw data and advanced information for developers",
	}
}

// GetUseCaseDescription returns descriptions for use cases
func GetUseCaseDescription() map[string]string {
	return map[string]string{
		"basic":     "Simplified translation for casual users and beginners",
		"technical": "Advanced translation with technical details for developers and analysts",
		"default":   "Balanced translation suitable for most users",
	}
}

// GetTranslationTypeDescription returns descriptions for translation types
func GetTranslationTypeDescription() map[string]string {
	return map[string]string{
		"transaction": "Convert transaction data into human-readable summaries",
		"address":     "Resolve and explain address information with labels and context",
		"contract":    "Analyze and explain smart contract functionality and purpose",
		"event":       "Decode and explain blockchain events and their significance",
		"block":       "Summarize block information and contained transactions",
	}
}

// GetTemplateExamples returns example templates for different languages
func GetTemplateExamples() map[string]map[string]string {
	return map[string]map[string]string{
		"en": {
			"simple_transfer": "{{.From}} sent {{.Value}} to {{.To}}",
			"contract_call":   "{{.From}} called {{.Method}} on {{.Contract}}",
			"token_transfer":  "{{.Amount}} {{.Token}} transferred from {{.From}} to {{.To}}",
			"failed_tx":       "Transaction failed: {{.Reason}}",
		},
		"es": {
			"simple_transfer": "{{.From}} envió {{.Value}} a {{.To}}",
			"contract_call":   "{{.From}} llamó {{.Method}} en {{.Contract}}",
			"token_transfer":  "{{.Amount}} {{.Token}} transferido de {{.From}} a {{.To}}",
			"failed_tx":       "Transacción falló: {{.Reason}}",
		},
		"fr": {
			"simple_transfer": "{{.From}} a envoyé {{.Value}} à {{.To}}",
			"contract_call":   "{{.From}} a appelé {{.Method}} sur {{.Contract}}",
			"token_transfer":  "{{.Amount}} {{.Token}} transféré de {{.From}} à {{.To}}",
			"failed_tx":       "Transaction échouée: {{.Reason}}",
		},
		"de": {
			"simple_transfer": "{{.From}} hat {{.Value}} an {{.To}} gesendet",
			"contract_call":   "{{.From}} hat {{.Method}} auf {{.Contract}} aufgerufen",
			"token_transfer":  "{{.Amount}} {{.Token}} von {{.From}} an {{.To}} übertragen",
			"failed_tx":       "Transaktion fehlgeschlagen: {{.Reason}}",
		},
		"ja": {
			"simple_transfer": "{{.From}}が{{.To}}に{{.Value}}を送信しました",
			"contract_call":   "{{.From}}が{{.Contract}}で{{.Method}}を呼び出しました",
			"token_transfer":  "{{.Amount}} {{.Token}}が{{.From}}から{{.To}}に転送されました",
			"failed_tx":       "トランザクションが失敗しました: {{.Reason}}",
		},
		"zh": {
			"simple_transfer": "{{.From}}向{{.To}}发送了{{.Value}}",
			"contract_call":   "{{.From}}在{{.Contract}}上调用了{{.Method}}",
			"token_transfer":  "{{.Amount}} {{.Token}}从{{.From}}转移到{{.To}}",
			"failed_tx":       "交易失败: {{.Reason}}",
		},
	}
}

// GetCommonAddressLabels returns common address labels
func GetCommonAddressLabels() map[string]string {
	return map[string]string{
		"0x0000000000000000000000000000000000000000": "Null Address",
		"0x000000000000000000000000000000000000dead": "Burn Address",
		"0xdac17f958d2ee523a2206206994597c13d831ec7": "Tether USD (USDT)",
		"0xa0b86a33e6c3c8c95c5c8c8c8c8c8c8c8c8c8c8c": "Uniswap V2 Router",
		"0x7a250d5630b4cf539739df2c5dacb4c659f2488d": "Uniswap V2 Router 02",
		"0x1f9840a85d5af5bf1d1762f925bdaddc4201f984": "Uniswap Token (UNI)",
		"0x6b175474e89094c44da98b954eedeac495271d0f": "Dai Stablecoin (DAI)",
		"0xa0b73e1ff0b80914ab6fe0444e65848c4c34450b": "Compound cETH",
		"0x5d3a536e4d6dbd6114cc1ead35777bab948e3643": "Compound cDAI",
	}
}

// GetEventSignatures returns common event signatures
func GetEventSignatures() map[string]string {
	return map[string]string{
		"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef": "Transfer(address,address,uint256)",
		"0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925": "Approval(address,address,uint256)",
		"0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31": "ApprovalForAll(address,address,bool)",
		"0xc3d58168c5ae7397731d063d5bbf3d657854427343f4c083240f7aacaa2d0f62": "TransferSingle(address,address,address,uint256,uint256)",
		"0x4a39dc06d4c0dbc64b70af90fd698a233a518aa5d07e595d983b8c0526c8f7fb": "TransferBatch(address,address,address,uint256[],uint256[])",
	}
}

// GetMethodSignatures returns common method signatures
func GetMethodSignatures() map[string]string {
	return map[string]string{
		"0xa9059cbb": "transfer(address,uint256)",
		"0x23b872dd": "transferFrom(address,address,uint256)",
		"0x095ea7b3": "approve(address,uint256)",
		"0x70a08231": "balanceOf(address)",
		"0x18160ddd": "totalSupply()",
		"0x06fdde03": "name()",
		"0x95d89b41": "symbol()",
		"0x313ce567": "decimals()",
		"0x42842e0e": "safeTransferFrom(address,address,uint256)",
		"0xb88d4fde": "safeTransferFrom(address,address,uint256,bytes)",
	}
}
