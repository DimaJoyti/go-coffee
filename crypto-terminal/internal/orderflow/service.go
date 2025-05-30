package orderflow

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/config"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/models"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

// Service handles order flow analysis operations
type Service struct {
	config           *config.Config
	db               *sql.DB
	redis            *redis.Client
	tickCollector    *TickCollector
	footprintEngine  *FootprintEngine
	volumeProfiler   *VolumeProfiler
	deltaAnalyzer    *DeltaAnalyzer
	imbalanceDetector *ImbalanceDetector
	isHealthy        bool
	mu               sync.RWMutex
	stopChan         chan struct{}
	subscribers      map[string][]chan *models.OrderFlowWebSocketMessage
}

// NewService creates a new order flow service
func NewService(cfg *config.Config, db *sql.DB, redis *redis.Client) (*Service, error) {
	service := &Service{
		config:      cfg,
		db:          db,
		redis:       redis,
		isHealthy:   true,
		stopChan:    make(chan struct{}),
		subscribers: make(map[string][]chan *models.OrderFlowWebSocketMessage),
	}

	// Initialize components
	var err error
	
	service.tickCollector, err = NewTickCollector(cfg, redis)
	if err != nil {
		return nil, fmt.Errorf("failed to create tick collector: %w", err)
	}

	service.footprintEngine = NewFootprintEngine(cfg)
	service.volumeProfiler = NewVolumeProfiler(cfg)
	service.deltaAnalyzer = NewDeltaAnalyzer(cfg)
	service.imbalanceDetector = NewImbalanceDetector(cfg)

	// Initialize database tables
	if err := service.initializeTables(); err != nil {
		return nil, fmt.Errorf("failed to initialize tables: %w", err)
	}

	return service, nil
}

// Start starts the order flow service
func (s *Service) Start(ctx context.Context) error {
	logrus.Info("Starting order flow service")

	// Start tick collection
	if err := s.tickCollector.Start(ctx); err != nil {
		return fmt.Errorf("failed to start tick collector: %w", err)
	}

	// Start real-time processing
	go s.startRealTimeProcessing(ctx)

	// Start order flow analysis
	go s.startOrderFlowAnalysis(ctx)

	// Start imbalance detection
	go s.startImbalanceDetection(ctx)

	logrus.Info("Order flow service started")
	return nil
}

// Stop stops the order flow service
func (s *Service) Stop() error {
	logrus.Info("Stopping order flow service")
	close(s.stopChan)
	
	if s.tickCollector != nil {
		s.tickCollector.Stop()
	}
	
	return nil
}

// IsHealthy returns the health status of the service
func (s *Service) IsHealthy() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isHealthy
}

// GetFootprintData returns footprint chart data for a symbol and timeframe
func (s *Service) GetFootprintData(ctx context.Context, symbol, timeframe string, startTime, endTime time.Time, config models.OrderFlowConfig) (*models.FootprintChartData, error) {
	// Get ticks for the time range
	ticks, err := s.getTicksForRange(ctx, symbol, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticks: %w", err)
	}

	// Generate footprint bars
	bars, err := s.footprintEngine.GenerateFootprintBars(ticks, config)
	if err != nil {
		return nil, fmt.Errorf("failed to generate footprint bars: %w", err)
	}

	return &models.FootprintChartData{
		Symbol:    symbol,
		Timeframe: timeframe,
		StartTime: startTime,
		EndTime:   endTime,
		Bars:      bars,
		Config:    config,
		Metadata: map[string]interface{}{
			"total_ticks": len(ticks),
			"total_bars":  len(bars),
		},
	}, nil
}

// GetVolumeProfile returns volume profile data for a symbol and time range
func (s *Service) GetVolumeProfile(ctx context.Context, symbol, profileType string, startTime, endTime time.Time, config models.OrderFlowConfig) (*models.VolumeProfileChartData, error) {
	// Get ticks for the time range
	ticks, err := s.getTicksForRange(ctx, symbol, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticks: %w", err)
	}

	// Generate volume profile
	profile, err := s.volumeProfiler.GenerateVolumeProfile(ticks, profileType, config)
	if err != nil {
		return nil, fmt.Errorf("failed to generate volume profile: %w", err)
	}

	return &models.VolumeProfileChartData{
		Symbol:      symbol,
		ProfileType: profileType,
		StartTime:   startTime,
		EndTime:     endTime,
		Profile:     *profile,
		PriceLevels: profile.PriceLevels,
		Config:      config,
	}, nil
}

// GetDeltaAnalysis returns delta analysis data for a symbol and timeframe
func (s *Service) GetDeltaAnalysis(ctx context.Context, symbol, timeframe string, startTime, endTime time.Time, config models.OrderFlowConfig) (*models.DeltaAnalysisData, error) {
	// Get ticks for the time range
	ticks, err := s.getTicksForRange(ctx, symbol, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticks: %w", err)
	}

	// Generate delta analysis
	deltaProfile, err := s.deltaAnalyzer.AnalyzeDelta(ticks, config)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze delta: %w", err)
	}

	// Get historical delta data
	deltaHistory, err := s.getDeltaHistory(ctx, symbol, timeframe, startTime, endTime)
	if err != nil {
		logrus.Warnf("Failed to get delta history: %v", err)
		deltaHistory = []models.DeltaProfile{}
	}

	// Detect divergences
	divergences, err := s.imbalanceDetector.DetectDeltaDivergences(ticks, *deltaProfile, config)
	if err != nil {
		logrus.Warnf("Failed to detect divergences: %v", err)
		divergences = []models.OrderFlowImbalance{}
	}

	return &models.DeltaAnalysisData{
		Symbol:       symbol,
		Timeframe:    timeframe,
		StartTime:    startTime,
		EndTime:      endTime,
		DeltaProfile: *deltaProfile,
		DeltaHistory: deltaHistory,
		Divergences:  divergences,
		Config:       config,
	}, nil
}

// GetOrderFlowMetrics returns real-time order flow metrics
func (s *Service) GetOrderFlowMetrics(ctx context.Context, symbol string) (*models.OrderFlowMetrics, error) {
	// Get latest tick data
	latestTick, err := s.getLatestTick(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest tick: %w", err)
	}

	// Calculate real-time metrics
	metrics := &models.OrderFlowMetrics{
		Symbol:        symbol,
		Timestamp:     time.Now(),
		CurrentPrice:  latestTick.Price,
		BidPrice:      latestTick.BidPrice,
		AskPrice:      latestTick.AskPrice,
		LastTradeVolume: latestTick.Volume,
		LastTradeSide: latestTick.Side,
	}

	// Get session delta
	sessionDelta, err := s.getSessionDelta(ctx, symbol)
	if err != nil {
		logrus.Warnf("Failed to get session delta: %v", err)
	} else {
		metrics.SessionDelta = sessionDelta
	}

	// Get cumulative delta
	cumulativeDelta, err := s.getCumulativeDelta(ctx, symbol)
	if err != nil {
		logrus.Warnf("Failed to get cumulative delta: %v", err)
	} else {
		metrics.CumulativeDelta = cumulativeDelta
	}

	// Get active imbalances count
	activeImbalances, err := s.getActiveImbalancesCount(ctx, symbol)
	if err != nil {
		logrus.Warnf("Failed to get active imbalances: %v", err)
	} else {
		metrics.ActiveImbalances = activeImbalances
	}

	return metrics, nil
}

// GetActiveImbalances returns currently active order flow imbalances
func (s *Service) GetActiveImbalances(ctx context.Context, symbol string) ([]models.OrderFlowImbalance, error) {
	// Implementation placeholder - would query database for active imbalances
	return []models.OrderFlowImbalance{}, nil
}

// SubscribeToOrderFlow subscribes to real-time order flow updates
func (s *Service) SubscribeToOrderFlow(symbol string) <-chan *models.OrderFlowWebSocketMessage {
	s.mu.Lock()
	defer s.mu.Unlock()

	ch := make(chan *models.OrderFlowWebSocketMessage, 100)
	if s.subscribers[symbol] == nil {
		s.subscribers[symbol] = make([]chan *models.OrderFlowWebSocketMessage, 0)
	}
	s.subscribers[symbol] = append(s.subscribers[symbol], ch)

	return ch
}

// UnsubscribeFromOrderFlow unsubscribes from real-time order flow updates
func (s *Service) UnsubscribeFromOrderFlow(symbol string, ch <-chan *models.OrderFlowWebSocketMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if subscribers, exists := s.subscribers[symbol]; exists {
		for i, subscriber := range subscribers {
			if subscriber == ch {
				// Remove the channel from the slice
				s.subscribers[symbol] = append(subscribers[:i], subscribers[i+1:]...)
				close(subscriber)
				break
			}
		}
	}
}

// broadcastOrderFlowUpdate broadcasts order flow updates to subscribers
func (s *Service) broadcastOrderFlowUpdate(symbol string, message *models.OrderFlowWebSocketMessage) {
	s.mu.RLock()
	subscribers := s.subscribers[symbol]
	s.mu.RUnlock()

	for _, ch := range subscribers {
		select {
		case ch <- message:
		default:
			// Channel is full, skip this subscriber
			logrus.Warnf("Order flow subscriber channel full for symbol %s", symbol)
		}
	}
}

// Helper methods

func (s *Service) getTicksForRange(ctx context.Context, symbol string, startTime, endTime time.Time) ([]models.Tick, error) {
	// Implementation placeholder - would query database for ticks in time range
	return []models.Tick{}, nil
}

func (s *Service) getLatestTick(ctx context.Context, symbol string) (*models.Tick, error) {
	// Implementation placeholder - would get latest tick from cache or database
	return &models.Tick{
		Symbol:    symbol,
		Price:     decimal.NewFromFloat(50000),
		Volume:    decimal.NewFromFloat(0.1),
		Side:      "BUY",
		Timestamp: time.Now(),
	}, nil
}

func (s *Service) getSessionDelta(ctx context.Context, symbol string) (decimal.Decimal, error) {
	// Implementation placeholder
	return decimal.NewFromFloat(1500), nil
}

func (s *Service) getCumulativeDelta(ctx context.Context, symbol string) (decimal.Decimal, error) {
	// Implementation placeholder
	return decimal.NewFromFloat(2500), nil
}

func (s *Service) getActiveImbalancesCount(ctx context.Context, symbol string) (int, error) {
	// Implementation placeholder
	return 3, nil
}

func (s *Service) getDeltaHistory(ctx context.Context, symbol, timeframe string, startTime, endTime time.Time) ([]models.DeltaProfile, error) {
	// Implementation placeholder
	return []models.DeltaProfile{}, nil
}

func (s *Service) initializeTables() error {
	// Implementation placeholder - would create order flow tables
	return nil
}

func (s *Service) startRealTimeProcessing(ctx context.Context) {
	ticker := time.NewTicker(100 * time.Millisecond) // High frequency updates
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.processRealTimeUpdates(ctx)
		}
	}
}

func (s *Service) startOrderFlowAnalysis(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.performOrderFlowAnalysis(ctx)
		}
	}
}

func (s *Service) startImbalanceDetection(ctx context.Context) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.detectImbalances(ctx)
		}
	}
}

func (s *Service) processRealTimeUpdates(ctx context.Context) {
	// Implementation placeholder - process real-time tick updates
}

func (s *Service) performOrderFlowAnalysis(ctx context.Context) {
	// Implementation placeholder - perform order flow analysis
}

func (s *Service) detectImbalances(ctx context.Context) {
	// Implementation placeholder - detect order flow imbalances
}
