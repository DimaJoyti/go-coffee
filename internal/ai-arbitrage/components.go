package aiarbitrage

import (
	"context"
	"fmt"
	"math"
	"time"

	pb "github.com/DimaJoyti/go-coffee/api/proto"
	"github.com/DimaJoyti/go-coffee/pkg/config"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
	"github.com/DimaJoyti/go-coffee/pkg/redis-mcp"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// OpportunityDetector detects arbitrage opportunities using AI
type OpportunityDetector struct {
	aiService *redismcp.AIService
	logger    *logger.Logger
	config    *config.Config
}

// NewOpportunityDetector creates a new opportunity detector
func NewOpportunityDetector(aiService *redismcp.AIService, logger *logger.Logger, cfg *config.Config) (*OpportunityDetector, error) {
	return &OpportunityDetector{
		aiService: aiService,
		logger:    logger,
		config:    cfg,
	}, nil
}

// DetectOpportunities detects arbitrage opportunities from market data
func (od *OpportunityDetector) DetectOpportunities(ctx context.Context, marketData []MarketData) ([]*pb.ArbitrageOpportunity, error) {
	var opportunities []*pb.ArbitrageOpportunity
	
	// Analyze market data for arbitrage opportunities
	for i := 0; i < len(marketData); i++ {
		for j := i + 1; j < len(marketData); j++ {
			if marketData[i].AssetSymbol == marketData[j].AssetSymbol {
				// Check for price difference
				priceDiff := math.Abs(marketData[i].Price - marketData[j].Price)
				profitMargin := (priceDiff / math.Min(marketData[i].Price, marketData[j].Price)) * 100
				
				if profitMargin > 1.0 { // Minimum 1% profit margin
					var buyMarket, sellMarket string
					var buyPrice, sellPrice float64
					
					if marketData[i].Price < marketData[j].Price {
						buyMarket = marketData[i].Exchange
						sellMarket = marketData[j].Exchange
						buyPrice = marketData[i].Price
						sellPrice = marketData[j].Price
					} else {
						buyMarket = marketData[j].Exchange
						sellMarket = marketData[i].Exchange
						buyPrice = marketData[j].Price
						sellPrice = marketData[i].Price
					}
					
					opportunity := &pb.ArbitrageOpportunity{
						Id:              generateID("detected"),
						AssetSymbol:     marketData[i].AssetSymbol,
						BuyPrice:        buyPrice,
						SellPrice:       sellPrice,
						ProfitMargin:    profitMargin,
						Volume:          math.Min(marketData[i].Volume, marketData[j].Volume),
						BuyMarket:       buyMarket,
						SellMarket:      sellMarket,
						Status:          pb.OpportunityStatus_OPPORTUNITY_ACTIVE,
						RiskScore:       od.calculateRiskScore(profitMargin),
						ConfidenceScore: od.calculateConfidenceScore(marketData[i], marketData[j]),
						CreatedAt:       timestamppb.Now(),
						ExpiresAt:       timestamppb.New(time.Now().Add(time.Hour)),
						Tags:            []string{"ai-detected"},
					}
					
					opportunities = append(opportunities, opportunity)
				}
			}
		}
	}
	
	return opportunities, nil
}

// calculateRiskScore calculates risk score based on profit margin and other factors
func (od *OpportunityDetector) calculateRiskScore(profitMargin float64) float64 {
	// Higher profit margins might indicate higher risk
	if profitMargin > 10 {
		return 0.8
	} else if profitMargin > 5 {
		return 0.6
	} else if profitMargin > 2 {
		return 0.4
	}
	return 0.2
}

// calculateConfidenceScore calculates confidence score based on market data quality
func (od *OpportunityDetector) calculateConfidenceScore(data1, data2 MarketData) float64 {
	// Base confidence
	confidence := 0.7
	
	// Adjust based on volume
	if data1.Volume > 1000 && data2.Volume > 1000 {
		confidence += 0.1
	}
	
	// Adjust based on data freshness
	if time.Since(data1.Timestamp) < time.Minute*5 && time.Since(data2.Timestamp) < time.Minute*5 {
		confidence += 0.1
	}
	
	return math.Min(confidence, 1.0)
}

// MatchingEngine matches buyers and sellers using AI algorithms
type MatchingEngine struct {
	aiService *redismcp.AIService
	logger    *logger.Logger
	config    *config.Config
}

// NewMatchingEngine creates a new matching engine
func NewMatchingEngine(aiService *redismcp.AIService, logger *logger.Logger, cfg *config.Config) (*MatchingEngine, error) {
	return &MatchingEngine{
		aiService: aiService,
		logger:    logger,
		config:    cfg,
	}, nil
}

// FindMatches finds optimal matches between buyers and sellers
func (me *MatchingEngine) FindMatches(ctx context.Context, opportunity *pb.ArbitrageOpportunity, participants []*pb.Participant) ([]*pb.ParticipantMatch, error) {
	var matches []*pb.ParticipantMatch
	
	// Separate buyers and sellers
	var buyers, sellers []*pb.Participant
	for _, participant := range participants {
		switch participant.Type {
		case pb.ParticipantType_PARTICIPANT_BUYER, pb.ParticipantType_PARTICIPANT_BOTH:
			buyers = append(buyers, participant)
		case pb.ParticipantType_PARTICIPANT_SELLER, pb.ParticipantType_PARTICIPANT_BOTH:
			sellers = append(sellers, participant)
		}
	}
	
	// Find optimal matches
	for _, buyer := range buyers {
		for _, seller := range sellers {
			if buyer.Id == seller.Id {
				continue // Skip same participant
			}
			
			matchScore := me.calculateMatchScore(opportunity, buyer, seller)
			if matchScore > 0.5 { // Minimum match threshold
				match := &pb.ParticipantMatch{
					BuyerId:           buyer.Id,
					SellerId:          seller.Id,
					MatchScore:        matchScore,
					SuggestedQuantity: me.calculateOptimalQuantity(opportunity, buyer, seller),
					SuggestedPrice:    (opportunity.BuyPrice + opportunity.SellPrice) / 2,
					MatchAnalysis: &pb.AIAnalysis{
						Confidence:      matchScore,
						Recommendation:  fmt.Sprintf("Good match between %s and %s", buyer.Name, seller.Name),
						RiskFactors:     []string{"Counterparty risk", "Execution risk"},
						Opportunities:   []string{"Profitable arbitrage"},
					},
				}
				matches = append(matches, match)
			}
		}
	}
	
	return matches, nil
}

// calculateMatchScore calculates compatibility score between buyer and seller
func (me *MatchingEngine) calculateMatchScore(opportunity *pb.ArbitrageOpportunity, buyer, seller *pb.Participant) float64 {
	score := 0.5 // Base score
	
	// Check asset preferences
	if me.hasAssetPreference(buyer, opportunity.AssetSymbol) {
		score += 0.2
	}
	if me.hasAssetPreference(seller, opportunity.AssetSymbol) {
		score += 0.2
	}
	
	// Check risk tolerance
	if buyer.RiskProfile != nil && buyer.RiskProfile.Tolerance == pb.RiskTolerance_RISK_HIGH {
		score += 0.1
	}
	if seller.RiskProfile != nil && seller.RiskProfile.Tolerance == pb.RiskTolerance_RISK_HIGH {
		score += 0.1
	}
	
	return math.Min(score, 1.0)
}

// hasAssetPreference checks if participant has preference for the asset
func (me *MatchingEngine) hasAssetPreference(participant *pb.Participant, assetSymbol string) bool {
	if participant.Preferences == nil {
		return false
	}
	
	for _, asset := range participant.Preferences.PreferredAssets {
		if asset == assetSymbol {
			return true
		}
	}
	return false
}

// calculateOptimalQuantity calculates optimal trade quantity
func (me *MatchingEngine) calculateOptimalQuantity(opportunity *pb.ArbitrageOpportunity, buyer, seller *pb.Participant) float64 {
	// Start with opportunity volume
	quantity := opportunity.Volume
	
	// Limit by buyer's max trade amount
	if buyer.Preferences != nil && buyer.Preferences.MaxTradeAmount > 0 {
		maxQuantity := buyer.Preferences.MaxTradeAmount / opportunity.BuyPrice
		quantity = math.Min(quantity, maxQuantity)
	}
	
	// Limit by seller's max trade amount
	if seller.Preferences != nil && seller.Preferences.MaxTradeAmount > 0 {
		maxQuantity := seller.Preferences.MaxTradeAmount / opportunity.SellPrice
		quantity = math.Min(quantity, maxQuantity)
	}
	
	return quantity
}

// RiskManager manages risk assessment and limits
type RiskManager struct {
	aiService *redismcp.AIService
	logger    *logger.Logger
	config    *config.Config
}

// NewRiskManager creates a new risk manager
func NewRiskManager(aiService *redismcp.AIService, logger *logger.Logger, cfg *config.Config) (*RiskManager, error) {
	return &RiskManager{
		aiService: aiService,
		logger:    logger,
		config:    cfg,
	}, nil
}

// RiskAssessment represents the result of risk assessment
type RiskAssessment struct {
	Approved    bool
	RiskScore   float64
	Reason      string
	Limitations []string
}

// AssessTradeRisk assesses risk for a trade
func (rm *RiskManager) AssessTradeRisk(ctx context.Context, opportunity *pb.ArbitrageOpportunity, buyer, seller *pb.Participant, quantity float64) (*RiskAssessment, error) {
	assessment := &RiskAssessment{
		Approved:  true,
		RiskScore: 0.0,
	}
	
	// Check opportunity risk score
	if opportunity.RiskScore > 0.8 {
		assessment.RiskScore += 0.3
		assessment.Limitations = append(assessment.Limitations, "High opportunity risk")
	}
	
	// Check participant risk profiles
	if buyer.RiskProfile != nil && buyer.RiskProfile.Tolerance == pb.RiskTolerance_RISK_LOW {
		if opportunity.ProfitMargin > 5.0 {
			assessment.RiskScore += 0.2
			assessment.Limitations = append(assessment.Limitations, "High profit margin for low-risk buyer")
		}
	}
	
	// Check trade amount limits
	tradeAmount := quantity * opportunity.BuyPrice
	if buyer.RiskProfile != nil && tradeAmount > buyer.RiskProfile.MaxExposure {
		assessment.Approved = false
		assessment.Reason = "Trade amount exceeds buyer's maximum exposure"
		return assessment, nil
	}
	
	// Final risk assessment
	if assessment.RiskScore > 0.7 {
		assessment.Approved = false
		assessment.Reason = "Overall risk score too high"
	} else {
		assessment.Reason = "Trade approved with acceptable risk"
	}
	
	return assessment, nil
}

// MarketAnalyzer provides AI-powered market analysis
type MarketAnalyzer struct {
	aiService *redismcp.AIService
	logger    *logger.Logger
	config    *config.Config
}

// NewMarketAnalyzer creates a new market analyzer
func NewMarketAnalyzer(aiService *redismcp.AIService, logger *logger.Logger, cfg *config.Config) (*MarketAnalyzer, error) {
	return &MarketAnalyzer{
		aiService: aiService,
		logger:    logger,
		config:    cfg,
	}, nil
}

// AnalyzeMarket performs comprehensive market analysis
func (ma *MarketAnalyzer) AnalyzeMarket(ctx context.Context, assetSymbol string, markets []string, timeframe string) (*pb.MarketAnalysis, error) {
	// Simulate market analysis (in real implementation, this would
	// fetch real market data and perform complex analysis)
	
	analysis := &pb.MarketAnalysis{
		AssetSymbol:     assetSymbol,
		CurrentPrice:    100.0, // Simulated current price
		PredictedPrice:  102.5, // AI prediction
		Volatility:      0.15,  // 15% volatility
		SentimentScore:  0.65,  // Slightly positive sentiment
		RiskFactors:     []string{"Market volatility", "Regulatory changes", "Liquidity risk"},
		SupportLevels:   []*pb.PriceLevel{{Price: 95.0, Strength: 0.8, Type: "support"}},
		ResistanceLevels: []*pb.PriceLevel{{Price: 110.0, Strength: 0.7, Type: "resistance"}},
		AnalysisTime:    timestamppb.Now(),
	}
	
	return analysis, nil
}

// MarketData represents market data from an exchange
type MarketData struct {
	AssetSymbol string
	Exchange    string
	Price       float64
	Volume      float64
	Timestamp   time.Time
}
