package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/internal/defi"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/config"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/shopspring/decimal"
)

// TradingStrategiesExample демонструє використання алгоритмічних торгових стратегій
func main() {
	ctx := context.Background()

	// Ініціалізація компонентів
	appLogger := logger.New("trading-example")

	// Create mock cache (simplified for example)
	cache := &mockRedisClient{}

	// Створення blockchain клієнтів (mock для прикладу)
	ethClient := &mockEthereumClient{}
	bscClient := &mockEthereumClient{}
	polygonClient := &mockEthereumClient{}
	solanaClient := &mockSolanaClient{}

	// Конфігурація DeFi
	defiConfig := config.DeFiConfig{
		OneInch: config.OneInchConfig{
			APIKey: "your-1inch-api-key",
		},
	}

	// Створення DeFi сервісу
	defiService := defi.NewService(
		ethClient,
		bscClient,
		polygonClient,
		solanaClient,
		nil, // raydiumClient
		nil, // jupiterClient
		cache,
		appLogger,
		defiConfig,
	)

	// Запуск сервісу
	if err := defiService.Start(ctx); err != nil {
		log.Fatal("Failed to start DeFi service:", err)
	}
	defer defiService.Stop()

	fmt.Println("🚀 DeFi Trading Strategies Example Started")

	// Демонстрація різних стратегій
	demonstrateArbitrageDetection(ctx, defiService)
	demonstrateYieldFarming(ctx, defiService)
	demonstrateOnChainAnalysis(ctx, defiService)
	demonstrateTradingBots(ctx, defiService)
}

// demonstrateArbitrageDetection демонструє виявлення арбітражних можливостей
func demonstrateArbitrageDetection(ctx context.Context, service *defi.Service) {
	fmt.Println("\n📊 === Арбітражні Можливості ===")

	// Отримання всіх арбітражних можливостей
	opportunities, err := service.GetArbitrageOpportunities(ctx)
	if err != nil {
		log.Printf("Error getting arbitrage opportunities: %v", err)
		return
	}

	fmt.Printf("Знайдено %d арбітражних можливостей:\n", len(opportunities))

	for i, opp := range opportunities {
		if i >= 3 { // Показуємо тільки перші 3
			break
		}

		fmt.Printf("\n🔄 Арбітраж #%d:\n", i+1)
		fmt.Printf("  Токен: %s\n", opp.Token.Symbol)
		fmt.Printf("  Джерело: %s (ціна: $%s)\n",
			opp.SourceExchange.Name, opp.SourcePrice)
		fmt.Printf("  Ціль: %s (ціна: $%s)\n",
			opp.TargetExchange.Name, opp.TargetPrice)
		fmt.Printf("  Прибуток: %s%% (впевненість: %s%%)\n",
			opp.ProfitMargin.Mul(decimal.NewFromInt(100)),
			opp.Confidence.Mul(decimal.NewFromInt(100)))
		fmt.Printf("  Ризик: %s\n", opp.Risk)
		fmt.Printf("  Чистий прибуток: $%s\n", opp.NetProfit)
	}

	// Пошук арбітражу для конкретного токена
	wethToken := defi.Token{
		Address: "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2",
		Symbol:  "WETH",
		Chain:   defi.ChainEthereum,
	}

	wethOpportunities, err := service.DetectArbitrageForToken(ctx, wethToken)
	if err != nil {
		log.Printf("Error detecting WETH arbitrage: %v", err)
		return
	}

	fmt.Printf("\n🔍 WETH арбітраж: знайдено %d можливостей\n", len(wethOpportunities))
}

// demonstrateYieldFarming демонструє yield farming стратегії
func demonstrateYieldFarming(ctx context.Context, service *defi.Service) {
	fmt.Println("\n🌾 === Yield Farming Стратегії ===")

	// Отримання найкращих yield можливостей
	opportunities, err := service.GetBestYieldOpportunities(ctx, 5)
	if err != nil {
		log.Printf("Error getting yield opportunities: %v", err)
		return
	}

	fmt.Printf("Топ %d yield farming можливостей:\n", len(opportunities))

	for i, opp := range opportunities {
		fmt.Printf("\n💰 Можливість #%d:\n", i+1)
		fmt.Printf("  Протокол: %s\n", opp.Protocol)
		fmt.Printf("  Стратегія: %s\n", opp.Strategy)
		fmt.Printf("  APY: %s%%\n", opp.APY.Mul(decimal.NewFromInt(100)))
		fmt.Printf("  TVL: $%s\n", opp.TVL)
		fmt.Printf("  Мін. депозит: $%s\n", opp.MinDeposit)
		fmt.Printf("  Ризик: %s\n", opp.Risk)
		if !opp.ImpermanentLoss.IsZero() {
			fmt.Printf("  Impermanent Loss: %s%%\n",
				opp.ImpermanentLoss.Mul(decimal.NewFromInt(100)))
		}
	}

	// Отримання оптимальної стратегії
	strategyRequest := &defi.OptimalStrategyRequest{
		InvestmentAmount: decimal.NewFromFloat(10000), // $10,000
		RiskTolerance:    defi.RiskLevelMedium,
		MinAPY:           decimal.NewFromFloat(0.08), // 8% мінімум
		MaxLockPeriod:    time.Hour * 24 * 30,        // 30 днів макс
		AutoCompound:     true,
		Diversification:  true,
	}

	strategy, err := service.GetOptimalYieldStrategy(ctx, strategyRequest)
	if err != nil {
		log.Printf("Error getting optimal strategy: %v", err)
		return
	}

	if strategy != nil {
		fmt.Printf("\n🎯 Оптимальна стратегія:\n")
		fmt.Printf("  Назва: %s\n", strategy.Name)
		fmt.Printf("  Тип: %s\n", strategy.Type)
		fmt.Printf("  Загальний APY: %s%%\n",
			strategy.TotalAPY.Mul(decimal.NewFromInt(100)))
		fmt.Printf("  Ризик: %s\n", strategy.Risk)
		fmt.Printf("  Мін. інвестиція: $%s\n", strategy.MinInvestment)
		fmt.Printf("  Автокомпаундинг: %t\n", strategy.AutoCompound)
		fmt.Printf("  Кількість можливостей: %d\n", len(strategy.Opportunities))
	}
}

// demonstrateOnChainAnalysis демонструє он-чейн аналітику
func demonstrateOnChainAnalysis(ctx context.Context, service *defi.Service) {
	fmt.Println("\n🔗 === Он-чейн Аналітика ===")

	// Аналіз WETH токена
	wethAddress := "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"

	metrics, err := service.GetOnChainMetrics(ctx, wethAddress)
	if err != nil {
		log.Printf("Error getting WETH metrics: %v", err)
	} else {
		fmt.Printf("📈 WETH Метрики:\n")
		fmt.Printf("  Ціна: $%s\n", metrics.Price)
		fmt.Printf("  Об'єм 24h: $%s\n", metrics.Volume24h)
		fmt.Printf("  Ліквідність: $%s\n", metrics.Liquidity)
		fmt.Printf("  Холдери: %d\n", metrics.Holders)
		fmt.Printf("  Транзакції 24h: %d\n", metrics.Transactions24h)
		fmt.Printf("  Волатильність: %s%%\n",
			metrics.Volatility.Mul(decimal.NewFromInt(100)))
	}

	// Ринкові сигнали
	signals, err := service.GetMarketSignals(ctx)
	if err != nil {
		log.Printf("Error getting market signals: %v", err)
	} else {
		fmt.Printf("\n📡 Ринкові сигнали (%d):\n", len(signals))

		for i, signal := range signals {
			if i >= 3 { // Показуємо тільки перші 3
				break
			}

			fmt.Printf("\n🚨 Сигнал #%d:\n", i+1)
			fmt.Printf("  Тип: %s\n", signal.Type)
			fmt.Printf("  Токен: %s\n", signal.Token.Symbol)
			fmt.Printf("  Напрямок: %s\n", signal.Direction)
			fmt.Printf("  Сила: %s%%\n",
				signal.Strength.Mul(decimal.NewFromInt(100)))
			fmt.Printf("  Впевненість: %s%%\n",
				signal.Confidence.Mul(decimal.NewFromInt(100)))
			fmt.Printf("  Причина: %s\n", signal.Reason)
			fmt.Printf("  Закінчується: %s\n",
				signal.ExpiresAt.Format("15:04:05"))
		}
	}

	// Активність китів
	whales, err := service.GetWhaleActivity(ctx)
	if err != nil {
		log.Printf("Error getting whale activity: %v", err)
	} else {
		fmt.Printf("\n🐋 Активність китів (%d):\n", len(whales))

		for i, whale := range whales {
			if i >= 2 { // Показуємо тільки перших 2
				break
			}

			fmt.Printf("\n🐋 Кит #%d:\n", i+1)
			fmt.Printf("  Мітка: %s\n", whale.Label)
			fmt.Printf("  Баланс: $%s\n", whale.Balance)
			fmt.Printf("  Транзакції 24h: %d\n", whale.TxCount24h)
			fmt.Printf("  Об'єм 24h: $%s\n", whale.Volume24h)
			fmt.Printf("  Остання транзакція: %s\n",
				whale.LastTx.Format("15:04:05"))
		}
	}
}

// demonstrateTradingBots демонструє торгові боти
func demonstrateTradingBots(ctx context.Context, service *defi.Service) {
	fmt.Println("\n🤖 === Торгові Боти ===")

	// Конфігурація для арбітражного бота
	arbitrageBotConfig := defi.TradingBotConfig{
		MaxPositionSize:   decimal.NewFromFloat(5000),  // $5,000
		MinProfitMargin:   decimal.NewFromFloat(0.01),  // 1%
		MaxSlippage:       decimal.NewFromFloat(0.005), // 0.5%
		RiskTolerance:     defi.RiskLevelMedium,
		AutoCompound:      true,
		MaxDailyTrades:    10,
		StopLossPercent:   decimal.NewFromFloat(0.05), // 5%
		TakeProfitPercent: decimal.NewFromFloat(0.15), // 15%
		ExecutionDelay:    time.Second * 5,
	}

	// Створення арбітражного бота
	arbitrageBot, err := service.CreateTradingBot(ctx,
		"Coffee Arbitrage Bot",
		defi.StrategyTypeArbitrage,
		arbitrageBotConfig)
	if err != nil {
		log.Printf("Error creating arbitrage bot: %v", err)
		return
	}

	fmt.Printf("✅ Створено арбітражний бот: %s (ID: %s)\n",
		arbitrageBot.Name, arbitrageBot.ID)

	// Конфігурація для yield farming бота
	yieldBotConfig := defi.TradingBotConfig{
		MaxPositionSize: decimal.NewFromFloat(10000), // $10,000
		MinProfitMargin: decimal.NewFromFloat(0.05),  // 5% APY мінімум
		RiskTolerance:   defi.RiskLevelLow,
		AutoCompound:    true,
		MaxDailyTrades:  3,
		ExecutionDelay:  time.Second * 10,
	}

	// Створення yield farming бота
	yieldBot, err := service.CreateTradingBot(ctx,
		"Coffee Yield Bot",
		defi.StrategyTypeYieldFarming,
		yieldBotConfig)
	if err != nil {
		log.Printf("Error creating yield bot: %v", err)
		return
	}

	fmt.Printf("✅ Створено yield farming бот: %s (ID: %s)\n",
		yieldBot.Name, yieldBot.ID)

	// Створення DCA бота
	dcaBot, err := service.CreateTradingBot(ctx,
		"Coffee DCA Bot",
		defi.StrategyTypeDCA,
		defi.TradingBotConfig{
			MaxPositionSize: decimal.NewFromFloat(1000), // $1,000
			RiskTolerance:   defi.RiskLevelLow,
			AutoCompound:    false,
			MaxDailyTrades:  1, // Одна покупка на день
			ExecutionDelay:  time.Minute,
		})
	if err != nil {
		log.Printf("Error creating DCA bot: %v", err)
		return
	}

	fmt.Printf("✅ Створено DCA бот: %s (ID: %s)\n",
		dcaBot.Name, dcaBot.ID)

	// Запуск ботів (в реальному застосунку це робилося б обережно)
	fmt.Println("\n🚀 Запуск ботів...")

	if err := service.StartTradingBot(ctx, arbitrageBot.ID); err != nil {
		log.Printf("Error starting arbitrage bot: %v", err)
	} else {
		fmt.Printf("▶️  Арбітражний бот запущено\n")
	}

	if err := service.StartTradingBot(ctx, yieldBot.ID); err != nil {
		log.Printf("Error starting yield bot: %v", err)
	} else {
		fmt.Printf("▶️  Yield farming бот запущено\n")
	}

	// Симуляція роботи протягом короткого часу
	fmt.Println("\n⏱️  Симуляція роботи протягом 10 секунд...")
	time.Sleep(10 * time.Second)

	// Перевірка статусу та продуктивності ботів
	fmt.Println("\n📊 Статус ботів:")

	bots, err := service.GetAllTradingBots(ctx)
	if err != nil {
		log.Printf("Error getting bots: %v", err)
		return
	}

	for _, bot := range bots {
		performance, err := service.GetTradingBotPerformance(ctx, bot.ID)
		if err != nil {
			log.Printf("Error getting bot performance: %v", err)
			continue
		}

		fmt.Printf("\n🤖 %s:\n", bot.Name)
		fmt.Printf("  Статус: %s\n", bot.GetStatus())
		fmt.Printf("  Стратегія: %s\n", bot.Strategy)
		fmt.Printf("  Загальні угоди: %d\n", performance.TotalTrades)
		fmt.Printf("  Прибуткові угоди: %d\n", performance.WinningTrades)
		if performance.TotalTrades > 0 {
			fmt.Printf("  Win Rate: %s%%\n",
				performance.WinRate.Mul(decimal.NewFromInt(100)))
		}
		fmt.Printf("  Чистий прибуток: $%s\n", performance.NetProfit)

		// Активні позиції
		positions := bot.GetActivePositions()
		fmt.Printf("  Активні позиції: %d\n", len(positions))
	}

	// Зупинка ботів
	fmt.Println("\n⏹️  Зупинка ботів...")
	for _, bot := range bots {
		if err := service.StopTradingBot(ctx, bot.ID); err != nil {
			log.Printf("Error stopping bot %s: %v", bot.Name, err)
		} else {
			fmt.Printf("⏹️  %s зупинено\n", bot.Name)
		}
	}

	fmt.Println("\n✅ Демонстрація торгових стратегій завершена!")
}

// Mock types for compilation
type mockRedisClient struct{}

func (m *mockRedisClient) Get(key string) (string, error)                        { return "", nil }
func (m *mockRedisClient) Set(key, value string, expiration time.Duration) error { return nil }
func (m *mockRedisClient) Del(keys ...string) error                              { return nil }
func (m *mockRedisClient) Close() error                                          { return nil }

type mockEthereumClient struct{}
type mockSolanaClient struct{}
