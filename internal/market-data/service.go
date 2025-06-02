package marketdata

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/DimaJoyti/go-coffee/api/proto"
	"github.com/DimaJoyti/go-coffee/pkg/config"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
	redismcp "github.com/DimaJoyti/go-coffee/pkg/redis-mcp"
)

// Service implements the Market Data service
type Service struct {
	pb.UnimplementedMarketDataServiceServer
	
	// Dependencies
	redisClient *redismcp.RedisClient
	logger      *logger.Logger
	config      *config.Config
	
	// Market data storage
	currentPrices map[string]*pb.MarketPrice
	historicalData map[string][]*pb.HistoricalPrice
	
	// Subscribers for real-time updates
	priceSubscribers map[string][]chan *pb.PriceUpdate
	
	// Concurrency control
	mutex sync.RWMutex
	
	// Background services
	isRunning bool
	stopChan  chan struct{}
}

// NewService creates a new Market Data service
func NewService(
	redisClient *redismcp.RedisClient,
	logger *logger.Logger,
	cfg *config.Config,
) (*Service, error) {
	service := &Service{
		redisClient:      redisClient,
		logger:           logger,
		config:           cfg,
		currentPrices:    make(map[string]*pb.MarketPrice),
		historicalData:   make(map[string][]*pb.HistoricalPrice),
		priceSubscribers: make(map[string][]chan *pb.PriceUpdate),
		stopChan:         make(chan struct{}),
	}
	
	// Initialize with sample data
	service.initializeSampleData()
	
	return service, nil
}

// Start starts the background services
func (s *Service) Start(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if s.isRunning {
		return fmt.Errorf("service is already running")
	}
	
	s.logger.Info("Starting Market Data Service background processes")
	
	// Start price update simulation
	go s.priceUpdateLoop(ctx)
	
	s.isRunning = true
	s.logger.Info("Market Data Service started successfully")
	
	return nil
}

// Stop stops the background services
func (s *Service) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if !s.isRunning {
		return
	}
	
	s.logger.Info("Stopping Market Data Service")
	close(s.stopChan)
	s.isRunning = false
}

// GetMarketPrices retrieves current market prices
func (s *Service) GetMarketPrices(ctx context.Context, req *pb.GetMarketPricesRequest) (*pb.GetMarketPricesResponse, error) {
	s.logger.WithFields(map[string]interface{}{
		"assets":  req.AssetSymbols,
		"markets": req.Markets,
	}).Info("Getting market prices")
	
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	var prices []*pb.MarketPrice
	
	for key, price := range s.currentPrices {
		// Filter by asset symbols
		if len(req.AssetSymbols) > 0 {
			found := false
			for _, symbol := range req.AssetSymbols {
				if price.AssetSymbol == symbol {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		
		// Filter by markets
		if len(req.Markets) > 0 {
			found := false
			for _, market := range req.Markets {
				if price.Market == market {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		
		prices = append(prices, price)
	}
	
	return &pb.GetMarketPricesResponse{
		Prices:  prices,
		Success: true,
		Message: fmt.Sprintf("Retrieved %d prices", len(prices)),
	}, nil
}

// SubscribeToPrices provides real-time price updates
func (s *Service) SubscribeToPrices(req *pb.SubscribePricesRequest, stream pb.MarketDataService_SubscribeToPricesServer) error {
	s.logger.Info("Client subscribed to price updates",
		logger.String("participant_id", req.ParticipantId),
		logger.Strings("assets", req.AssetSymbols),
	)
	
	// Create subscription channel
	subscription := make(chan *pb.PriceUpdate, 100)
	defer close(subscription)
	
	// Register subscription
	subscriptionKey := fmt.Sprintf("%s_%s", req.ParticipantId, time.Now().Format("20060102150405"))
	s.mutex.Lock()
	s.priceSubscribers[subscriptionKey] = append(s.priceSubscribers[subscriptionKey], subscription)
	s.mutex.Unlock()
	
	// Cleanup on disconnect
	defer func() {
		s.mutex.Lock()
		delete(s.priceSubscribers, subscriptionKey)
		s.mutex.Unlock()
	}()
	
	// Send current prices first
	s.mutex.RLock()
	for _, price := range s.currentPrices {
		if s.shouldSendPriceUpdate(price, req) {
			update := &pb.PriceUpdate{
				Price:      price,
				UpdateType: "CURRENT",
				Timestamp:  timestamppb.Now(),
			}
			select {
			case subscription <- update:
			default:
				// Channel full, skip
			}
		}
	}
	s.mutex.RUnlock()
	
	// Send real-time updates
	for {
		select {
		case update := <-subscription:
			if err := stream.Send(update); err != nil {
				s.logger.Error("Failed to send price update", logger.Error(err))
				return err
			}
		case <-stream.Context().Done():
			s.logger.Info("Client disconnected from price stream",
				logger.String("participant_id", req.ParticipantId),
			)
			return nil
		}
	}
}

// GetHistoricalData retrieves historical price data
func (s *Service) GetHistoricalData(ctx context.Context, req *pb.GetHistoricalDataRequest) (*pb.GetHistoricalDataResponse, error) {
	s.logger.Info("Getting historical data",
		logger.String("asset", req.AssetSymbol),
		logger.String("market", req.Market),
		logger.String("interval", req.Interval),
	)
	
	key := fmt.Sprintf("%s_%s", req.AssetSymbol, req.Market)
	
	s.mutex.RLock()
	data, exists := s.historicalData[key]
	s.mutex.RUnlock()
	
	if !exists {
		// Generate sample historical data
		data = s.generateHistoricalData(req.AssetSymbol, req.StartTime.AsTime(), req.EndTime.AsTime())
		
		s.mutex.Lock()
		s.historicalData[key] = data
		s.mutex.Unlock()
	}
	
	// Filter by time range
	var filteredData []*pb.HistoricalPrice
	for _, price := range data {
		if price.Timestamp.AsTime().After(req.StartTime.AsTime()) && 
		   price.Timestamp.AsTime().Before(req.EndTime.AsTime()) {
			filteredData = append(filteredData, price)
		}
	}
	
	return &pb.GetHistoricalDataResponse{
		Prices:  filteredData,
		Success: true,
		Message: fmt.Sprintf("Retrieved %d historical prices", len(filteredData)),
	}, nil
}

// GetMarketDepth retrieves market depth (order book)
func (s *Service) GetMarketDepth(ctx context.Context, req *pb.GetMarketDepthRequest) (*pb.GetMarketDepthResponse, error) {
	s.logger.Info("Getting market depth",
		logger.String("asset", req.AssetSymbol),
		logger.String("market", req.Market),
		logger.Int32("depth", req.Depth),
	)
	
	// Generate sample market depth data
	bids := s.generateOrderBookLevels(100.0, false, int(req.Depth))
	asks := s.generateOrderBookLevels(100.5, true, int(req.Depth))
	
	return &pb.GetMarketDepthResponse{
		Bids:    bids,
		Asks:    asks,
		Success: true,
		Message: "Market depth retrieved successfully",
	}, nil
}

// Helper functions

// initializeSampleData initializes the service with sample market data
func (s *Service) initializeSampleData() {
	assets := []string{"COFFEE", "BTC", "ETH", "USDT"}
	markets := []string{"Exchange_A", "Exchange_B", "Exchange_C"}
	
	for _, asset := range assets {
		for _, market := range markets {
			basePrice := s.getBasePrice(asset)
			// Add some variation between markets
			variation := (rand.Float64() - 0.5) * 0.1 // ±5% variation
			price := basePrice * (1 + variation)
			
			marketPrice := &pb.MarketPrice{
				AssetSymbol: asset,
				Market:      market,
				BidPrice:    price * 0.999,
				AskPrice:    price * 1.001,
				LastPrice:   price,
				Volume24H:   rand.Float64() * 1000000,
				Change24H:   (rand.Float64() - 0.5) * 0.2, // ±10% change
				Timestamp:   timestamppb.Now(),
			}
			
			key := fmt.Sprintf("%s_%s", asset, market)
			s.currentPrices[key] = marketPrice
		}
	}
}

// getBasePrice returns base price for an asset
func (s *Service) getBasePrice(asset string) float64 {
	switch asset {
	case "COFFEE":
		return 100.0
	case "BTC":
		return 45000.0
	case "ETH":
		return 3000.0
	case "USDT":
		return 1.0
	default:
		return 100.0
	}
}

// shouldSendPriceUpdate determines if a price update should be sent to subscriber
func (s *Service) shouldSendPriceUpdate(price *pb.MarketPrice, req *pb.SubscribePricesRequest) bool {
	// Filter by asset symbols
	if len(req.AssetSymbols) > 0 {
		found := false
		for _, symbol := range req.AssetSymbols {
			if price.AssetSymbol == symbol {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	
	// Filter by markets
	if len(req.Markets) > 0 {
		found := false
		for _, market := range req.Markets {
			if price.Market == market {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	
	return true
}

// generateHistoricalData generates sample historical data
func (s *Service) generateHistoricalData(asset string, start, end time.Time) []*pb.HistoricalPrice {
	var data []*pb.HistoricalPrice
	
	basePrice := s.getBasePrice(asset)
	currentPrice := basePrice
	
	for t := start; t.Before(end); t = t.Add(time.Hour) {
		// Simulate price movement
		change := (rand.Float64() - 0.5) * 0.02 // ±1% change per hour
		currentPrice *= (1 + change)
		
		high := currentPrice * (1 + rand.Float64()*0.01)
		low := currentPrice * (1 - rand.Float64()*0.01)
		
		data = append(data, &pb.HistoricalPrice{
			Timestamp: timestamppb.New(t),
			Open:      currentPrice,
			High:      high,
			Low:       low,
			Close:     currentPrice,
			Volume:    rand.Float64() * 10000,
		})
	}
	
	return data
}

// generateOrderBookLevels generates sample order book levels
func (s *Service) generateOrderBookLevels(basePrice float64, isAsk bool, depth int) []*pb.OrderBookLevel {
	var levels []*pb.OrderBookLevel
	
	for i := 0; i < depth; i++ {
		var price float64
		if isAsk {
			price = basePrice + float64(i)*0.01
		} else {
			price = basePrice - float64(i)*0.01
		}
		
		levels = append(levels, &pb.OrderBookLevel{
			Price:      price,
			Quantity:   rand.Float64() * 1000,
			OrderCount: int32(rand.Intn(10) + 1),
		})
	}
	
	return levels
}

// priceUpdateLoop simulates real-time price updates
func (s *Service) priceUpdateLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 5) // Update every 5 seconds
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			s.updatePrices()
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		}
	}
}

// updatePrices updates current prices and notifies subscribers
func (s *Service) updatePrices() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	for key, price := range s.currentPrices {
		// Simulate price movement
		change := (rand.Float64() - 0.5) * 0.01 // ±0.5% change
		newPrice := price.LastPrice * (1 + change)
		
		price.LastPrice = newPrice
		price.BidPrice = newPrice * 0.999
		price.AskPrice = newPrice * 1.001
		price.Timestamp = timestamppb.Now()
		
		// Notify subscribers
		update := &pb.PriceUpdate{
			Price:      price,
			UpdateType: "UPDATE",
			Timestamp:  timestamppb.Now(),
		}
		
		for _, subscribers := range s.priceSubscribers {
			for _, subscriber := range subscribers {
				select {
				case subscriber <- update:
				default:
					// Channel full, skip
				}
			}
		}
	}
}
