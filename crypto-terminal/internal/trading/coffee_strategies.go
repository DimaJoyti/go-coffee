package trading

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// CoffeeStrategyFactory creates coffee-themed trading strategies
type CoffeeStrategyFactory struct{}

// NewCoffeeStrategyFactory creates a new strategy factory
func NewCoffeeStrategyFactory() *CoffeeStrategyFactory {
	return &CoffeeStrategyFactory{}
}

// CreateEspressoStrategy creates a high-frequency scalping strategy
func (csf *CoffeeStrategyFactory) CreateEspressoStrategy(symbol string) *CoffeeStrategy {
	return &CoffeeStrategy{
		ID:          uuid.New().String(),
		Name:        "Espresso Scalper",
		Type:        StrategyEspresso,
		Description: "High-frequency scalping strategy for quick profits, like a shot of espresso - fast and intense",
		Status:      StatusPaused,
		Config: StrategyConfig{
			Symbol:              symbol,
			Timeframe:           "1m",
			MaxPositionSize:     decimal.NewFromFloat(0.05), // 5% of portfolio
			StopLossPercent:     decimal.NewFromFloat(0.005), // 0.5%
			TakeProfitPercent:   decimal.NewFromFloat(0.01),  // 1%
			MinConfidence:       decimal.NewFromFloat(0.8),   // 80%
			UseTrailingStop:     true,
			TrailingStopPercent: decimal.NewFromFloat(0.003), // 0.3%
			MaxDailyTrades:      50,
			TradingHours: TradingHours{
				Enabled:   true,
				StartHour: 8,  // 8 AM
				EndHour:   18, // 6 PM
				Timezone:  "UTC",
				Weekdays:  []int{1, 2, 3, 4, 5}, // Monday to Friday
			},
			Indicators:        []string{"RSI", "MACD", "EMA_9", "EMA_21"},
			CoffeeCorrelation: true,
		},
		Performance: StrategyPerformance{
			TotalTrades:   0,
			WinningTrades: 0,
			LosingTrades:  0,
			WinRate:       decimal.Zero,
			TotalPnL:      decimal.Zero,
			LastUpdated:   time.Now(),
		},
		RiskManagement: RiskConfig{
			MaxPortfolioRisk:  decimal.NewFromFloat(0.01), // 1%
			MaxPositionRisk:   decimal.NewFromFloat(0.005), // 0.5%
			MaxCorrelation:    decimal.NewFromFloat(0.3),   // 30%
			EmergencyStopLoss: decimal.NewFromFloat(0.02),  // 2%
			DailyLossLimit:    decimal.NewFromFloat(0.005), // 0.5%
			MaxLeverage:       decimal.NewFromFloat(3),     // 3x
			RiskRewardRatio:   decimal.NewFromFloat(2),     // 1:2
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// CreateLatteStrategy creates a smooth swing trading strategy
func (csf *CoffeeStrategyFactory) CreateLatteStrategy(symbol string) *CoffeeStrategy {
	return &CoffeeStrategy{
		ID:          uuid.New().String(),
		Name:        "Latte Swing",
		Type:        StrategyLatte,
		Description: "Smooth swing trading strategy for balanced returns, like a perfect latte - smooth and satisfying",
		Status:      StatusPaused,
		Config: StrategyConfig{
			Symbol:              symbol,
			Timeframe:           "15m",
			MaxPositionSize:     decimal.NewFromFloat(0.2),  // 20% of portfolio
			StopLossPercent:     decimal.NewFromFloat(0.02), // 2%
			TakeProfitPercent:   decimal.NewFromFloat(0.04), // 4%
			MinConfidence:       decimal.NewFromFloat(0.7),  // 70%
			UseTrailingStop:     true,
			TrailingStopPercent: decimal.NewFromFloat(0.015), // 1.5%
			MaxDailyTrades:      10,
			TradingHours: TradingHours{
				Enabled:   true,
				StartHour: 6,  // 6 AM
				EndHour:   22, // 10 PM
				Timezone:  "UTC",
				Weekdays:  []int{1, 2, 3, 4, 5, 6, 7}, // All week
			},
			Indicators:        []string{"RSI", "MACD", "BB", "EMA_50", "EMA_200"},
			CoffeeCorrelation: true,
		},
		Performance: StrategyPerformance{
			TotalTrades:   0,
			WinningTrades: 0,
			LosingTrades:  0,
			WinRate:       decimal.Zero,
			TotalPnL:      decimal.Zero,
			LastUpdated:   time.Now(),
		},
		RiskManagement: RiskConfig{
			MaxPortfolioRisk:  decimal.NewFromFloat(0.03), // 3%
			MaxPositionRisk:   decimal.NewFromFloat(0.02), // 2%
			MaxCorrelation:    decimal.NewFromFloat(0.5),  // 50%
			EmergencyStopLoss: decimal.NewFromFloat(0.05), // 5%
			DailyLossLimit:    decimal.NewFromFloat(0.02), // 2%
			MaxLeverage:       decimal.NewFromFloat(5),    // 5x
			RiskRewardRatio:   decimal.NewFromFloat(2),    // 1:2
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// CreateColdBrewStrategy creates a patient position trading strategy
func (csf *CoffeeStrategyFactory) CreateColdBrewStrategy(symbol string) *CoffeeStrategy {
	return &CoffeeStrategy{
		ID:          uuid.New().String(),
		Name:        "Cold Brew Position",
		Type:        StrategyColdBrew,
		Description: "Patient position trading strategy for long-term gains, like cold brew - slow extraction, rich rewards",
		Status:      StatusPaused,
		Config: StrategyConfig{
			Symbol:              symbol,
			Timeframe:           "4h",
			MaxPositionSize:     decimal.NewFromFloat(0.5),  // 50% of portfolio
			StopLossPercent:     decimal.NewFromFloat(0.05), // 5%
			TakeProfitPercent:   decimal.NewFromFloat(0.15), // 15%
			MinConfidence:       decimal.NewFromFloat(0.6),  // 60%
			UseTrailingStop:     true,
			TrailingStopPercent: decimal.NewFromFloat(0.03), // 3%
			MaxDailyTrades:      3,
			TradingHours: TradingHours{
				Enabled:   false, // 24/7 trading
				StartHour: 0,
				EndHour:   23,
				Timezone:  "UTC",
				Weekdays:  []int{1, 2, 3, 4, 5, 6, 7},
			},
			Indicators:        []string{"RSI", "MACD", "BB", "EMA_100", "EMA_200", "VWAP"},
			CoffeeCorrelation: false, // Less correlation for long-term
		},
		Performance: StrategyPerformance{
			TotalTrades:   0,
			WinningTrades: 0,
			LosingTrades:  0,
			WinRate:       decimal.Zero,
			TotalPnL:      decimal.Zero,
			LastUpdated:   time.Now(),
		},
		RiskManagement: RiskConfig{
			MaxPortfolioRisk:  decimal.NewFromFloat(0.05), // 5%
			MaxPositionRisk:   decimal.NewFromFloat(0.03), // 3%
			MaxCorrelation:    decimal.NewFromFloat(0.7),  // 70%
			EmergencyStopLoss: decimal.NewFromFloat(0.1),  // 10%
			DailyLossLimit:    decimal.NewFromFloat(0.03), // 3%
			MaxLeverage:       decimal.NewFromFloat(10),   // 10x
			RiskRewardRatio:   decimal.NewFromFloat(3),    // 1:3
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// CreateCappuccinoStrategy creates a frothy momentum trading strategy
func (csf *CoffeeStrategyFactory) CreateCappuccinoStrategy(symbol string) *CoffeeStrategy {
	return &CoffeeStrategy{
		ID:          uuid.New().String(),
		Name:        "Cappuccino Momentum",
		Type:        StrategyCappuccino,
		Description: "Frothy momentum trading strategy for dynamic markets, like cappuccino - light, airy, and energetic",
		Status:      StatusPaused,
		Config: StrategyConfig{
			Symbol:              symbol,
			Timeframe:           "5m",
			MaxPositionSize:     decimal.NewFromFloat(0.15), // 15% of portfolio
			StopLossPercent:     decimal.NewFromFloat(0.015), // 1.5%
			TakeProfitPercent:   decimal.NewFromFloat(0.03),  // 3%
			MinConfidence:       decimal.NewFromFloat(0.75),  // 75%
			UseTrailingStop:     true,
			TrailingStopPercent: decimal.NewFromFloat(0.01), // 1%
			MaxDailyTrades:      20,
			TradingHours: TradingHours{
				Enabled:   true,
				StartHour: 7,  // 7 AM
				EndHour:   19, // 7 PM
				Timezone:  "UTC",
				Weekdays:  []int{1, 2, 3, 4, 5}, // Weekdays only
			},
			Indicators:        []string{"RSI", "MACD", "BB", "EMA_12", "EMA_26", "Volume"},
			CoffeeCorrelation: true,
		},
		Performance: StrategyPerformance{
			TotalTrades:   0,
			WinningTrades: 0,
			LosingTrades:  0,
			WinRate:       decimal.Zero,
			TotalPnL:      decimal.Zero,
			LastUpdated:   time.Now(),
		},
		RiskManagement: RiskConfig{
			MaxPortfolioRisk:  decimal.NewFromFloat(0.025), // 2.5%
			MaxPositionRisk:   decimal.NewFromFloat(0.015), // 1.5%
			MaxCorrelation:    decimal.NewFromFloat(0.4),   // 40%
			EmergencyStopLoss: decimal.NewFromFloat(0.03),  // 3%
			DailyLossLimit:    decimal.NewFromFloat(0.015), // 1.5%
			MaxLeverage:       decimal.NewFromFloat(4),     // 4x
			RiskRewardRatio:   decimal.NewFromFloat(2),     // 1:2
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// CreateCustomStrategy creates a custom strategy with specified parameters
func (csf *CoffeeStrategyFactory) CreateCustomStrategy(name, symbol string, strategyType CoffeeStrategyType, config StrategyConfig) *CoffeeStrategy {
	return &CoffeeStrategy{
		ID:          uuid.New().String(),
		Name:        name,
		Type:        strategyType,
		Description: "Custom coffee-themed trading strategy",
		Status:      StatusPaused,
		Config:      config,
		Performance: StrategyPerformance{
			TotalTrades:   0,
			WinningTrades: 0,
			LosingTrades:  0,
			WinRate:       decimal.Zero,
			TotalPnL:      decimal.Zero,
			LastUpdated:   time.Now(),
		},
		RiskManagement: RiskConfig{
			MaxPortfolioRisk:  decimal.NewFromFloat(0.02),
			MaxPositionRisk:   decimal.NewFromFloat(0.01),
			MaxCorrelation:    decimal.NewFromFloat(0.5),
			EmergencyStopLoss: decimal.NewFromFloat(0.05),
			DailyLossLimit:    decimal.NewFromFloat(0.02),
			MaxLeverage:       decimal.NewFromFloat(5),
			RiskRewardRatio:   decimal.NewFromFloat(2),
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// GetStrategyRecommendation recommends a strategy based on market conditions
func (csf *CoffeeStrategyFactory) GetStrategyRecommendation(marketCondition string, volatility decimal.Decimal) CoffeeStrategyType {
	switch marketCondition {
	case "trending_up", "trending_down":
		if volatility.GreaterThan(decimal.NewFromFloat(0.05)) {
			return StrategyEspresso // High volatility trending market
		}
		return StrategyCappuccino // Moderate volatility trending market
	case "sideways", "ranging":
		if volatility.LessThan(decimal.NewFromFloat(0.02)) {
			return StrategyColdBrew // Low volatility ranging market
		}
		return StrategyLatte // Moderate volatility ranging market
	case "volatile", "uncertain":
		return StrategyEspresso // High volatility market
	default:
		return StrategyLatte // Default balanced approach
	}
}

// GetCoffeeThemeDescription returns a coffee-themed description for the strategy
func GetCoffeeThemeDescription(strategyType CoffeeStrategyType) string {
	switch strategyType {
	case StrategyEspresso:
		return "☕ Like a perfect espresso shot - quick, intense, and energizing. This strategy captures rapid market movements with precision timing."
	case StrategyLatte:
		return "🥛 Like a smooth latte - balanced, creamy, and satisfying. This strategy provides steady returns with controlled risk."
	case StrategyColdBrew:
		return "🧊 Like cold brew coffee - patient extraction for rich rewards. This strategy waits for the best opportunities for maximum profit."
	case StrategyCappuccino:
		return "🫖 Like a frothy cappuccino - light, airy, and dynamic. This strategy rides momentum waves with style and energy."
	default:
		return "☕ A custom coffee blend tailored to your trading taste."
	}
}

// GetOptimalTimeframes returns optimal timeframes for each strategy type
func GetOptimalTimeframes(strategyType CoffeeStrategyType) []string {
	switch strategyType {
	case StrategyEspresso:
		return []string{"1m", "3m", "5m"}
	case StrategyLatte:
		return []string{"15m", "30m", "1h"}
	case StrategyColdBrew:
		return []string{"4h", "1d", "3d"}
	case StrategyCappuccino:
		return []string{"5m", "15m", "30m"}
	default:
		return []string{"15m", "1h", "4h"}
	}
}

// GetRecommendedIndicators returns recommended indicators for each strategy type
func GetRecommendedIndicators(strategyType CoffeeStrategyType) []string {
	switch strategyType {
	case StrategyEspresso:
		return []string{"RSI", "MACD", "EMA_9", "EMA_21", "Volume", "ATR"}
	case StrategyLatte:
		return []string{"RSI", "MACD", "BB", "EMA_50", "EMA_200", "VWAP"}
	case StrategyColdBrew:
		return []string{"RSI", "MACD", "BB", "EMA_100", "EMA_200", "VWAP", "Ichimoku"}
	case StrategyCappuccino:
		return []string{"RSI", "MACD", "BB", "EMA_12", "EMA_26", "Volume", "Momentum"}
	default:
		return []string{"RSI", "MACD", "EMA_50", "EMA_200"}
	}
}
