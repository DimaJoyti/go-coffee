package main

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/DimaJoyti/go-coffee/api/proto"
)

const (
	arbitrageServiceAddr  = "localhost:50054"
	marketDataServiceAddr = "localhost:50055"
)

func main() {
	fmt.Println("🤖 AI Arbitrage System Demo")
	fmt.Println("===========================")
	fmt.Println("Connecting buyers and sellers through intelligent arbitrage")
	fmt.Println()

	// Connect to AI Arbitrage Service
	arbitrageConn, err := grpc.Dial(arbitrageServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to arbitrage service: %v", err)
	}
	defer arbitrageConn.Close()

	arbitrageClient := pb.NewArbitrageServiceClient(arbitrageConn)

	// Connect to Market Data Service
	marketDataConn, err := grpc.Dial(marketDataServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to market data service: %v", err)
	}
	defer marketDataConn.Close()

	marketDataClient := pb.NewMarketDataServiceClient(marketDataConn)

	ctx := context.Background()

	// Demo scenarios
	fmt.Println("🎬 Running AI Arbitrage Demo Scenarios...")
	fmt.Println()

	// Scenario 1: Get current market prices
	runMarketDataDemo(ctx, marketDataClient)

	// Scenario 2: Create arbitrage opportunities
	runArbitrageOpportunityDemo(ctx, arbitrageClient)

	// Scenario 3: Match participants
	runParticipantMatchingDemo(ctx, arbitrageClient)

	// Scenario 4: Execute trades
	runTradeExecutionDemo(ctx, arbitrageClient)

	// Scenario 5: Market analysis
	runMarketAnalysisDemo(ctx, arbitrageClient)

	fmt.Println("✅ AI Arbitrage Demo completed successfully!")
	fmt.Println()
	fmt.Println("🌟 Key Features Demonstrated:")
	fmt.Println("  • Real-time market data aggregation")
	fmt.Println("  • AI-powered opportunity detection")
	fmt.Println("  • Intelligent buyer-seller matching")
	fmt.Println("  • Risk-assessed trade execution")
	fmt.Println("  • Comprehensive market analysis")
}

func runMarketDataDemo(ctx context.Context, client pb.MarketDataServiceClient) {
	fmt.Println("📊 Scenario 1: Market Data Aggregation")
	fmt.Println("--------------------------------------")

	// Get current market prices
	pricesReq := &pb.GetMarketPricesRequest{
		AssetSymbols: []string{"COFFEE", "BTC", "ETH"},
		Markets:      []string{"Exchange_A", "Exchange_B", "Exchange_C"},
	}

	pricesResp, err := client.GetMarketPrices(ctx, pricesReq)
	if err != nil {
		fmt.Printf("❌ Failed to get market prices: %v\n", err)
		return
	}

	fmt.Printf("📈 Retrieved %d market prices:\n", len(pricesResp.Prices))
	for i, price := range pricesResp.Prices {
		if i >= 3 { // Limit output for demo
			break
		}
		fmt.Printf("  %s on %s: $%.2f (24h change: %.2f%%)\n",
			price.AssetSymbol, price.Market, price.LastPrice, price.Change24H*100)
	}
	fmt.Println()
}

func runArbitrageOpportunityDemo(ctx context.Context, client pb.ArbitrageServiceClient) {
	fmt.Println("🔍 Scenario 2: AI-Powered Opportunity Detection")
	fmt.Println("-----------------------------------------------")

	// Create a sample arbitrage opportunity
	createReq := &pb.CreateOpportunityRequest{
		AssetSymbol: "COFFEE",
		BuyPrice:    100.0,
		SellPrice:   105.0,
		Volume:      1000.0,
		BuyMarket:   "Exchange_A",
		SellMarket:  "Exchange_B",
		Tags:        []string{"demo", "coffee-arbitrage"},
	}

	createResp, err := client.CreateOpportunity(ctx, createReq)
	if err != nil {
		fmt.Printf("❌ Failed to create opportunity: %v\n", err)
		return
	}

	if createResp.Success {
		opp := createResp.Opportunity
		fmt.Printf("✅ Created arbitrage opportunity:\n")
		fmt.Printf("  ID: %s\n", opp.Id)
		fmt.Printf("  Asset: %s\n", opp.AssetSymbol)
		fmt.Printf("  Buy: $%.2f on %s\n", opp.BuyPrice, opp.BuyMarket)
		fmt.Printf("  Sell: $%.2f on %s\n", opp.SellPrice, opp.SellMarket)
		fmt.Printf("  Profit Margin: %.2f%%\n", opp.ProfitMargin)
		fmt.Printf("  Risk Score: %.2f\n", opp.RiskScore)
		fmt.Printf("  AI Confidence: %.2f\n", opp.ConfidenceScore)
		if opp.AiAnalysis != nil {
			fmt.Printf("  AI Recommendation: %s\n", opp.AiAnalysis.Recommendation)
		}
	}

	// Get available opportunities
	getReq := &pb.GetOpportunitiesRequest{
		AssetSymbol:     "COFFEE",
		MinProfitMargin: 2.0,
		MaxRiskScore:    0.8,
		Limit:           5,
	}

	getResp, err := client.GetOpportunities(ctx, getReq)
	if err != nil {
		fmt.Printf("❌ Failed to get opportunities: %v\n", err)
		return
	}

	fmt.Printf("\n📋 Found %d arbitrage opportunities:\n", len(getResp.Opportunities))
	for i, opp := range getResp.Opportunities {
		if i >= 2 { // Limit output for demo
			break
		}
		fmt.Printf("  %d. %s: %.2f%% profit (Risk: %.2f)\n",
			i+1, opp.AssetSymbol, opp.ProfitMargin, opp.RiskScore)
	}
	fmt.Println()
}

func runParticipantMatchingDemo(ctx context.Context, client pb.ArbitrageServiceClient) {
	fmt.Println("🤝 Scenario 3: Intelligent Participant Matching")
	fmt.Println("-----------------------------------------------")

	// First, get an opportunity to match against
	getReq := &pb.GetOpportunitiesRequest{
		Limit: 1,
	}

	getResp, err := client.GetOpportunities(ctx, getReq)
	if err != nil || len(getResp.Opportunities) == 0 {
		fmt.Println("❌ No opportunities available for matching demo")
		return
	}

	opportunity := getResp.Opportunities[0]

	// Simulate participant matching
	matchReq := &pb.MatchParticipantsRequest{
		OpportunityId:  opportunity.Id,
		ParticipantIds: []string{"buyer_001", "seller_001", "buyer_002", "seller_002"},
		AutoExecute:    false,
	}

	matchResp, err := client.MatchParticipants(ctx, matchReq)
	if err != nil {
		fmt.Printf("❌ Failed to match participants: %v\n", err)
		return
	}

	if matchResp.Success {
		fmt.Printf("✅ Found %d participant matches:\n", len(matchResp.Matches))
		for i, match := range matchResp.Matches {
			if i >= 2 { // Limit output for demo
				break
			}
			fmt.Printf("  Match %d:\n", i+1)
			fmt.Printf("    Buyer: %s\n", match.BuyerId)
			fmt.Printf("    Seller: %s\n", match.SellerId)
			fmt.Printf("    Match Score: %.2f\n", match.MatchScore)
			fmt.Printf("    Suggested Quantity: %.2f\n", match.SuggestedQuantity)
			fmt.Printf("    Suggested Price: $%.2f\n", match.SuggestedPrice)
		}
	}
	fmt.Println()
}

func runTradeExecutionDemo(ctx context.Context, client pb.ArbitrageServiceClient) {
	fmt.Println("⚡ Scenario 4: Risk-Assessed Trade Execution")
	fmt.Println("--------------------------------------------")

	// Get an opportunity for trade execution
	getReq := &pb.GetOpportunitiesRequest{
		Limit: 1,
	}

	getResp, err := client.GetOpportunities(ctx, getReq)
	if err != nil || len(getResp.Opportunities) == 0 {
		fmt.Println("❌ No opportunities available for execution demo")
		return
	}

	opportunity := getResp.Opportunities[0]

	// Execute a trade
	executeReq := &pb.ExecuteTradeRequest{
		OpportunityId: opportunity.Id,
		BuyerId:       "buyer_001",
		SellerId:      "seller_001",
		Quantity:      100.0,
		Price:         opportunity.SellPrice,
		ForceExecute:  false,
	}

	executeResp, err := client.ExecuteTrade(ctx, executeReq)
	if err != nil {
		fmt.Printf("❌ Failed to execute trade: %v\n", err)
		return
	}

	if executeResp.Success {
		trade := executeResp.Trade
		fmt.Printf("✅ Trade executed successfully:\n")
		fmt.Printf("  Trade ID: %s\n", trade.Id)
		fmt.Printf("  Asset: %s\n", trade.AssetSymbol)
		fmt.Printf("  Quantity: %.2f\n", trade.Quantity)
		fmt.Printf("  Buy Price: $%.2f\n", trade.BuyPrice)
		fmt.Printf("  Sell Price: $%.2f\n", trade.SellPrice)
		fmt.Printf("  Profit: $%.2f\n", trade.Profit)
		fmt.Printf("  Status: %v\n", trade.Status)
		fmt.Printf("  Transaction ID: %s\n", executeResp.TransactionId)
	} else {
		fmt.Printf("❌ Trade execution failed: %s\n", executeResp.Message)
	}
	fmt.Println()
}

func runMarketAnalysisDemo(ctx context.Context, client pb.ArbitrageServiceClient) {
	fmt.Println("📈 Scenario 5: AI-Powered Market Analysis")
	fmt.Println("-----------------------------------------")

	// Get market analysis
	analysisReq := &pb.GetMarketAnalysisRequest{
		AssetSymbol: "COFFEE",
		Markets:     []string{"Exchange_A", "Exchange_B", "Exchange_C"},
		Timeframe:   "1h",
	}

	analysisResp, err := client.GetMarketAnalysis(ctx, analysisReq)
	if err != nil {
		fmt.Printf("❌ Failed to get market analysis: %v\n", err)
		return
	}

	if analysisResp.Success {
		analysis := analysisResp.Analysis
		fmt.Printf("✅ Market Analysis for %s:\n", analysis.AssetSymbol)
		fmt.Printf("  Current Price: $%.2f\n", analysis.CurrentPrice)
		fmt.Printf("  Predicted Price: $%.2f\n", analysis.PredictedPrice)
		fmt.Printf("  Volatility: %.2f%%\n", analysis.Volatility*100)
		fmt.Printf("  Sentiment Score: %.2f\n", analysis.SentimentScore)

		if len(analysis.RiskFactors) > 0 {
			fmt.Printf("  Risk Factors:\n")
			for i, factor := range analysis.RiskFactors {
				if i >= 3 { // Limit output
					break
				}
				fmt.Printf("    • %s\n", factor)
			}
		}

		if len(analysis.SupportLevels) > 0 {
			fmt.Printf("  Support Levels:\n")
			for _, level := range analysis.SupportLevels {
				fmt.Printf("    • $%.2f (strength: %.2f)\n", level.Price, level.Strength)
			}
		}

		if len(analysis.ResistanceLevels) > 0 {
			fmt.Printf("  Resistance Levels:\n")
			for _, level := range analysis.ResistanceLevels {
				fmt.Printf("    • $%.2f (strength: %.2f)\n", level.Price, level.Strength)
			}
		}
	}
	fmt.Println()
}

func printSeparator() {
	fmt.Println("================================================")
}
