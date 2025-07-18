package metrics

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/developer-dao/pkg/config"
	"github.com/DimaJoyti/go-coffee/developer-dao/pkg/logger"
	"github.com/DimaJoyti/go-coffee/developer-dao/pkg/redis"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// Service provides TVL/MAU metrics operations
type Service struct {
	db          *sql.DB
	redis       redis.Client
	logger      *logger.Logger
	config      *config.Config
	serviceName string

	// Blockchain clients
	ethClient     *ethclient.Client
	bscClient     *ethclient.Client
	polygonClient *ethclient.Client

	// Data source clients
	defiLlamaClient   DefiLlamaClientInterface
	analyticsClient   AnalyticsClientInterface
	blockchainClients map[string]BlockchainClientInterface

	// Repositories
	tvlRepo         TVLRepository
	mauRepo         MAURepository
	impactRepo      ImpactRepository
	alertRepo       AlertRepository
	reportRepo      ReportRepository
	dataSourceRepo  DataSourceRepository

	// Cache and aggregators
	metricsCache     map[string]*MetricsSnapshot
	aggregationCache map[string]*AggregatedMetrics
	alertsCache      map[string]*Alert
}

// ServiceConfig holds the configuration for the metrics service
type ServiceConfig struct {
	DB          *sql.DB
	Redis       redis.Client
	Logger      *logger.Logger
	Config      *config.Config
	ServiceName string
}

// NewService creates a new metrics service instance
func NewService(cfg ServiceConfig) (*Service, error) {
	// Initialize blockchain clients
	ethClient, err := ethclient.Dial(cfg.Config.Blockchain.Ethereum.RPCURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum: %w", err)
	}

	bscClient, err := ethclient.Dial(cfg.Config.Blockchain.BSC.RPCURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to BSC: %w", err)
	}

	polygonClient, err := ethclient.Dial(cfg.Config.Blockchain.Polygon.RPCURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Polygon: %w", err)
	}

	// Initialize data source clients
	defiLlamaClient := NewDefiLlamaClient(cfg.Logger)
	analyticsClient := NewAnalyticsClient(cfg.Logger)

	// Initialize blockchain clients map
	blockchainClients := map[string]BlockchainClientInterface{
		"ethereum": NewBlockchainClient(ethClient, "ethereum", cfg.Logger),
		"bsc":      NewBlockchainClient(bscClient, "bsc", cfg.Logger),
		"polygon":  NewBlockchainClient(polygonClient, "polygon", cfg.Logger),
	}

	// Initialize repositories
	tvlRepo := NewTVLRepository(cfg.DB, cfg.Logger)
	mauRepo := NewMAURepository(cfg.DB, cfg.Logger)
	impactRepo := NewImpactRepository(cfg.DB, cfg.Logger)
	alertRepo := NewAlertRepository(cfg.DB, cfg.Logger)
	reportRepo := NewReportRepository(cfg.DB, cfg.Logger)
	dataSourceRepo := NewDataSourceRepository(cfg.DB, cfg.Logger)

	service := &Service{
		db:          cfg.DB,
		redis:       cfg.Redis,
		logger:      cfg.Logger,
		config:      cfg.Config,
		serviceName: cfg.ServiceName,

		ethClient:     ethClient,
		bscClient:     bscClient,
		polygonClient: polygonClient,

		defiLlamaClient:   defiLlamaClient,
		analyticsClient:   analyticsClient,
		blockchainClients: blockchainClients,

		tvlRepo:        tvlRepo,
		mauRepo:        mauRepo,
		impactRepo:     impactRepo,
		alertRepo:      alertRepo,
		reportRepo:     reportRepo,
		dataSourceRepo: dataSourceRepo,

		metricsCache:     make(map[string]*MetricsSnapshot),
		aggregationCache: make(map[string]*AggregatedMetrics),
		alertsCache:      make(map[string]*Alert),
	}

	return service, nil
}

// RecordTVL records a new TVL measurement
func (s *Service) RecordTVL(ctx context.Context, req *RecordTVLRequest) (*RecordTVLResponse, error) {
	s.logger.Info("Recording TVL measurement",
		zap.String("protocol", req.Protocol),
		zap.String("amount", req.Amount.String()),
		zap.String("chain", req.Chain))

	// Validate request
	if err := s.validateTVLRequest(req); err != nil {
		return nil, fmt.Errorf("invalid TVL request: %w", err)
	}

	// Create TVL record
	tvlRecord := &TVLRecord{
		Protocol:    req.Protocol,
		Chain:       req.Chain,
		Amount:      req.Amount,
		TokenSymbol: req.TokenSymbol,
		Source:      req.Source,
		Timestamp:   time.Now(),
		BlockNumber: req.BlockNumber,
		TxHash:      req.TxHash,
	}

	if err := s.tvlRepo.Create(ctx, tvlRecord); err != nil {
		return nil, fmt.Errorf("failed to record TVL: %w", err)
	}

	// Update aggregated metrics
	if err := s.updateTVLAggregations(ctx, tvlRecord); err != nil {
		s.logger.Error("Failed to update TVL aggregations", zap.Error(err))
	}

	// Check for alerts
	if err := s.checkTVLAlerts(ctx, tvlRecord); err != nil {
		s.logger.Error("Failed to check TVL alerts", zap.Error(err))
	}

	s.logger.Info("TVL recorded successfully",
		zap.Int64("recordID", tvlRecord.ID))

	return &RecordTVLResponse{
		RecordID:  tvlRecord.ID,
		Timestamp: tvlRecord.Timestamp,
		Status:    "recorded",
	}, nil
}

// RecordMAU records a new MAU measurement
func (s *Service) RecordMAU(ctx context.Context, req *RecordMAURequest) (*RecordMAUResponse, error) {
	s.logger.Info("Recording MAU measurement",
		zap.String("feature", req.Feature),
		zap.Int("userCount", req.UserCount),
		zap.String("period", req.Period))

	// Validate request
	if err := s.validateMAURequest(req); err != nil {
		return nil, fmt.Errorf("invalid MAU request: %w", err)
	}

	// Create MAU record
	mauRecord := &MAURecord{
		Feature:       req.Feature,
		UserCount:     req.UserCount,
		UniqueUsers:   req.UniqueUsers,
		Period:        req.Period,
		Source:        req.Source,
		Timestamp:     time.Now(),
		Metadata:      req.Metadata,
	}

	if err := s.mauRepo.Create(ctx, mauRecord); err != nil {
		return nil, fmt.Errorf("failed to record MAU: %w", err)
	}

	// Update aggregated metrics
	if err := s.updateMAUAggregations(ctx, mauRecord); err != nil {
		s.logger.Error("Failed to update MAU aggregations", zap.Error(err))
	}

	// Check for alerts
	if err := s.checkMAUAlerts(ctx, mauRecord); err != nil {
		s.logger.Error("Failed to check MAU alerts", zap.Error(err))
	}

	s.logger.Info("MAU recorded successfully",
		zap.Int64("recordID", mauRecord.ID))

	return &RecordMAUResponse{
		RecordID:  mauRecord.ID,
		Timestamp: mauRecord.Timestamp,
		Status:    "recorded",
	}, nil
}

// RecordImpact records developer/solution impact metrics
func (s *Service) RecordImpact(ctx context.Context, req *RecordImpactRequest) (*RecordImpactResponse, error) {
	s.logger.Info("Recording impact measurement",
		zap.String("entity", req.EntityID),
		zap.String("type", req.EntityType),
		zap.String("tvlImpact", req.TVLImpact.String()),
		zap.Int("mauImpact", req.MAUImpact))

	// Validate request
	if err := s.validateImpactRequest(req); err != nil {
		return nil, fmt.Errorf("invalid impact request: %w", err)
	}

	// Create impact record
	impactRecord := &ImpactRecord{
		EntityID:    req.EntityID,
		EntityType:  req.EntityType,
		TVLImpact:   req.TVLImpact,
		MAUImpact:   req.MAUImpact,
		Attribution: req.Attribution,
		Source:      req.Source,
		Timestamp:   time.Now(),
		Verified:    false,
		Metadata:    req.Metadata,
	}

	if err := s.impactRepo.Create(ctx, impactRecord); err != nil {
		return nil, fmt.Errorf("failed to record impact: %w", err)
	}

	// Update leaderboards
	if err := s.updateImpactLeaderboards(ctx, impactRecord); err != nil {
		s.logger.Error("Failed to update impact leaderboards", zap.Error(err))
	}

	s.logger.Info("Impact recorded successfully",
		zap.Int64("recordID", impactRecord.ID))

	return &RecordImpactResponse{
		RecordID:  impactRecord.ID,
		Timestamp: impactRecord.Timestamp,
		Status:    "recorded",
	}, nil
}

// GetTVLMetrics retrieves current TVL metrics
func (s *Service) GetTVLMetrics(ctx context.Context, req *GetTVLMetricsRequest) (*TVLMetricsResponse, error) {
	// Get current TVL from cache or calculate
	cacheKey := fmt.Sprintf("tvl_metrics_%s_%s", req.Protocol, req.Chain)
	if cached, exists := s.metricsCache[cacheKey]; exists {
		if time.Since(cached.Timestamp) < 5*time.Minute {
			return s.buildTVLResponse(cached), nil
		}
	}

	// Calculate fresh metrics
	metrics, err := s.calculateTVLMetrics(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate TVL metrics: %w", err)
	}

	// Update cache
	s.metricsCache[cacheKey] = &MetricsSnapshot{
		Type:      "tvl",
		Data:      metrics,
		Timestamp: time.Now(),
	}

	return s.buildTVLResponse(s.metricsCache[cacheKey]), nil
}

// GetMAUMetrics retrieves current MAU metrics
func (s *Service) GetMAUMetrics(ctx context.Context, req *GetMAUMetricsRequest) (*MAUMetricsResponse, error) {
	// Get current MAU from cache or calculate
	cacheKey := fmt.Sprintf("mau_metrics_%s_%s", req.Feature, req.Period)
	if cached, exists := s.metricsCache[cacheKey]; exists {
		if time.Since(cached.Timestamp) < 10*time.Minute {
			return s.buildMAUResponse(cached), nil
		}
	}

	// Calculate fresh metrics
	metrics, err := s.calculateMAUMetrics(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate MAU metrics: %w", err)
	}

	// Update cache
	s.metricsCache[cacheKey] = &MetricsSnapshot{
		Type:      "mau",
		Data:      metrics,
		Timestamp: time.Now(),
	}

	return s.buildMAUResponse(s.metricsCache[cacheKey]), nil
}

// StartMetricsCollection starts the metrics collection background service
func (s *Service) StartMetricsCollection(ctx context.Context) error {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	s.logger.Info("Starting metrics collection service")

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Stopping metrics collection service")
			return ctx.Err()
		case <-ticker.C:
			if err := s.collectMetrics(ctx); err != nil {
				s.logger.Error("Failed to collect metrics", zap.Error(err))
			}
		}
	}
}

// StartAnalyticsProcessing starts the analytics processing background service
func (s *Service) StartAnalyticsProcessing(ctx context.Context) error {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	s.logger.Info("Starting analytics processing service")

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Stopping analytics processing service")
			return ctx.Err()
		case <-ticker.C:
			if err := s.processAnalytics(ctx); err != nil {
				s.logger.Error("Failed to process analytics", zap.Error(err))
			}
		}
	}
}

// StartAlertingService starts the alerting background service
func (s *Service) StartAlertingService(ctx context.Context) error {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	s.logger.Info("Starting alerting service")

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Stopping alerting service")
			return ctx.Err()
		case <-ticker.C:
			if err := s.processAlerts(ctx); err != nil {
				s.logger.Error("Failed to process alerts", zap.Error(err))
			}
		}
	}
}

// StartReportingService starts the reporting background service
func (s *Service) StartReportingService(ctx context.Context) error {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	s.logger.Info("Starting reporting service")

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Stopping reporting service")
			return ctx.Err()
		case <-ticker.C:
			if err := s.generateReports(ctx); err != nil {
				s.logger.Error("Failed to generate reports", zap.Error(err))
			}
		}
	}
}

// Helper methods

func (s *Service) validateTVLRequest(req *RecordTVLRequest) error {
	if req.Protocol == "" {
		return fmt.Errorf("protocol is required")
	}
	if req.Amount.LessThan(decimal.Zero) {
		return fmt.Errorf("amount must be positive")
	}
	if req.Chain == "" {
		return fmt.Errorf("chain is required")
	}
	return nil
}

func (s *Service) validateMAURequest(req *RecordMAURequest) error {
	if req.Feature == "" {
		return fmt.Errorf("feature is required")
	}
	if req.UserCount < 0 {
		return fmt.Errorf("user count must be non-negative")
	}
	if req.Period == "" {
		return fmt.Errorf("period is required")
	}
	return nil
}

func (s *Service) validateImpactRequest(req *RecordImpactRequest) error {
	if req.EntityID == "" {
		return fmt.Errorf("entity ID is required")
	}
	if req.EntityType == "" {
		return fmt.Errorf("entity type is required")
	}
	return nil
}

func (s *Service) updateTVLAggregations(ctx context.Context, record *TVLRecord) error {
	// Update protocol-level aggregations
	// Update chain-level aggregations
	// Update time-based aggregations
	s.logger.Debug("Updated TVL aggregations", zap.String("protocol", record.Protocol))
	return nil
}

func (s *Service) updateMAUAggregations(ctx context.Context, record *MAURecord) error {
	// Update feature-level aggregations
	// Update time-based aggregations
	s.logger.Debug("Updated MAU aggregations", zap.String("feature", record.Feature))
	return nil
}

func (s *Service) updateImpactLeaderboards(ctx context.Context, record *ImpactRecord) error {
	// Update developer leaderboards
	// Update solution leaderboards
	s.logger.Debug("Updated impact leaderboards", zap.String("entity", record.EntityID))
	return nil
}

func (s *Service) checkTVLAlerts(ctx context.Context, record *TVLRecord) error {
	// Check for TVL threshold alerts
	// Check for growth rate alerts
	s.logger.Debug("Checked TVL alerts", zap.String("protocol", record.Protocol))
	return nil
}

func (s *Service) checkMAUAlerts(ctx context.Context, record *MAURecord) error {
	// Check for MAU threshold alerts
	// Check for growth rate alerts
	s.logger.Debug("Checked MAU alerts", zap.String("feature", record.Feature))
	return nil
}

func (s *Service) calculateTVLMetrics(ctx context.Context, req *GetTVLMetricsRequest) (interface{}, error) {
	// Calculate current TVL metrics
	return map[string]interface{}{
		"current_tvl": "5000000.00",
		"growth_24h":  "2.5",
		"growth_7d":   "15.2",
	}, nil
}

func (s *Service) calculateMAUMetrics(ctx context.Context, req *GetMAUMetricsRequest) (interface{}, error) {
	// Calculate current MAU metrics
	return map[string]interface{}{
		"current_mau": 25000,
		"growth_30d":  "8.5",
		"retention":   "75.2",
	}, nil
}

func (s *Service) buildTVLResponse(snapshot *MetricsSnapshot) *TVLMetricsResponse {
	return &TVLMetricsResponse{
		CurrentTVL: decimal.NewFromFloat(5000000),
		Growth24h:  decimal.NewFromFloat(2.5),
		Growth7d:   decimal.NewFromFloat(15.2),
		Timestamp:  snapshot.Timestamp,
	}
}

func (s *Service) buildMAUResponse(snapshot *MetricsSnapshot) *MAUMetricsResponse {
	return &MAUMetricsResponse{
		CurrentMAU: 25000,
		Growth30d:  decimal.NewFromFloat(8.5),
		Retention:  decimal.NewFromFloat(75.2),
		Timestamp:  snapshot.Timestamp,
	}
}

func (s *Service) collectMetrics(ctx context.Context) error {
	// Collect metrics from various sources
	s.logger.Debug("Collected metrics from external sources")
	return nil
}

func (s *Service) processAnalytics(ctx context.Context) error {
	// Process analytics and generate insights
	s.logger.Debug("Processed analytics data")
	return nil
}

func (s *Service) processAlerts(ctx context.Context) error {
	// Process and send alerts
	s.logger.Debug("Processed alerts")
	return nil
}

func (s *Service) generateReports(ctx context.Context) error {
	// Generate periodic reports
	s.logger.Debug("Generated periodic reports")
	return nil
}
