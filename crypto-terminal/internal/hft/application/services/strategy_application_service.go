package services

import (
	"context"
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/hft/domain/entities"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/hft/domain/repositories"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/hft/domain/valueobjects"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// StrategyApplicationService handles strategy-related use cases
type StrategyApplicationService struct {
	strategyRepo repositories.StrategyRepository
	orderRepo    repositories.OrderRepository
	eventStore   repositories.EventStore
	tracer       trace.Tracer
}

// NewStrategyApplicationService creates a new strategy application service
func NewStrategyApplicationService(
	strategyRepo repositories.StrategyRepository,
	orderRepo repositories.OrderRepository,
	eventStore repositories.EventStore,
) *StrategyApplicationService {
	return &StrategyApplicationService{
		strategyRepo: strategyRepo,
		orderRepo:    orderRepo,
		eventStore:   eventStore,
		tracer:       otel.Tracer("hft.strategy.application"),
	}
}

// CreateStrategyCommand represents a command to create a strategy
type CreateStrategyCommand struct {
	Name                string                 `json:"name" validate:"required"`
	Type                string                 `json:"type" validate:"required"`
	Symbols             []string               `json:"symbols" validate:"required,min=1"`
	Exchanges           []string               `json:"exchanges" validate:"required,min=1"`
	Parameters          map[string]interface{} `json:"parameters"`
	MaxPositionSize     string                 `json:"max_position_size" validate:"required"`
	MaxDailyLoss        string                 `json:"max_daily_loss" validate:"required"`
	MaxDrawdown         string                 `json:"max_drawdown" validate:"required"`
	MaxOrderSize        string                 `json:"max_order_size" validate:"required"`
	MaxOrdersPerSecond  int                    `json:"max_orders_per_second" validate:"min=1"`
	MaxExposure         string                 `json:"max_exposure" validate:"required"`
	StopLossPercent     string                 `json:"stop_loss_percent"`
	TakeProfitPercent   string                 `json:"take_profit_percent"`
}

// CreateStrategyResult represents the result of creating a strategy
type CreateStrategyResult struct {
	StrategyID string    `json:"strategy_id"`
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	Message    string    `json:"message"`
}

// CreateStrategy handles the create strategy use case
func (s *StrategyApplicationService) CreateStrategy(ctx context.Context, cmd CreateStrategyCommand) (*CreateStrategyResult, error) {
	ctx, span := s.tracer.Start(ctx, "StrategyApplicationService.CreateStrategy")
	defer span.End()

	span.SetAttributes(
		attribute.String("strategy_name", cmd.Name),
		attribute.String("strategy_type", cmd.Type),
		attribute.StringSlice("symbols", cmd.Symbols),
		attribute.StringSlice("exchanges", cmd.Exchanges),
	)

	// Convert command to domain types
	symbols := make([]entities.Symbol, len(cmd.Symbols))
	for i, symbol := range cmd.Symbols {
		symbols[i] = entities.Symbol(symbol)
	}

	exchanges := make([]entities.Exchange, len(cmd.Exchanges))
	for i, exchange := range cmd.Exchanges {
		exchanges[i] = entities.Exchange(exchange)
	}

	// Parse risk limits
	riskLimits, err := s.parseRiskLimits(cmd)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to parse risk limits: %w", err)
	}

	// Create strategy entity
	strategy, err := entities.NewStrategy(
		cmd.Name,
		valueobjects.StrategyType(cmd.Type),
		symbols,
		exchanges,
		cmd.Parameters,
		riskLimits,
	)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to create strategy entity: %w", err)
	}

	// Save strategy to repository
	if err := s.strategyRepo.Save(ctx, strategy); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to save strategy: %w", err)
	}

	// Save strategy events to event store
	if err := s.eventStore.SaveEvents(ctx, string(strategy.GetID()), s.convertStrategyEventsToGeneric(strategy.GetEvents())); err != nil {
		// Log error but don't fail the strategy creation
		span.AddEvent("failed to save strategy events", trace.WithAttributes(
			attribute.String("error", err.Error()),
		))
	}

	span.SetAttributes(
		attribute.String("strategy_id", string(strategy.GetID())),
		attribute.String("strategy_status", string(strategy.GetStatus())),
	)

	return &CreateStrategyResult{
		StrategyID: string(strategy.GetID()),
		Name:       strategy.GetName(),
		Type:       string(strategy.GetType()),
		Status:     string(strategy.GetStatus()),
		CreatedAt:  strategy.GetCreatedAt(),
		Message:    "Strategy created successfully",
	}, nil
}

// StartStrategyCommand represents a command to start a strategy
type StartStrategyCommand struct {
	StrategyID string `json:"strategy_id" validate:"required"`
}

// StartStrategyResult represents the result of starting a strategy
type StartStrategyResult struct {
	StrategyID string    `json:"strategy_id"`
	Status     string    `json:"status"`
	StartedAt  time.Time `json:"started_at"`
	Message    string    `json:"message"`
}

// StartStrategy handles the start strategy use case
func (s *StrategyApplicationService) StartStrategy(ctx context.Context, cmd StartStrategyCommand) (*StartStrategyResult, error) {
	ctx, span := s.tracer.Start(ctx, "StrategyApplicationService.StartStrategy")
	defer span.End()

	span.SetAttributes(attribute.String("strategy_id", cmd.StrategyID))

	// Find strategy
	strategy, err := s.strategyRepo.FindByID(ctx, entities.StrategyID(cmd.StrategyID))
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to find strategy: %w", err)
	}

	// Start strategy
	if err := strategy.Start(); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to start strategy: %w", err)
	}

	// Update strategy in repository
	if err := s.strategyRepo.Update(ctx, strategy); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to update strategy: %w", err)
	}

	// Save strategy events to event store
	if err := s.eventStore.SaveEvents(ctx, string(strategy.GetID()), s.convertStrategyEventsToGeneric(strategy.GetEvents())); err != nil {
		// Log error but don't fail the operation
		span.AddEvent("failed to save strategy events", trace.WithAttributes(
			attribute.String("error", err.Error()),
		))
	}

	span.SetAttributes(
		attribute.String("strategy_status", string(strategy.GetStatus())),
	)

	return &StartStrategyResult{
		StrategyID: string(strategy.GetID()),
		Status:     string(strategy.GetStatus()),
		StartedAt:  *strategy.GetStartedAt(),
		Message:    "Strategy started successfully",
	}, nil
}

// StopStrategyCommand represents a command to stop a strategy
type StopStrategyCommand struct {
	StrategyID string `json:"strategy_id" validate:"required"`
}

// StopStrategyResult represents the result of stopping a strategy
type StopStrategyResult struct {
	StrategyID string    `json:"strategy_id"`
	Status     string    `json:"status"`
	StoppedAt  time.Time `json:"stopped_at"`
	Message    string    `json:"message"`
}

// StopStrategy handles the stop strategy use case
func (s *StrategyApplicationService) StopStrategy(ctx context.Context, cmd StopStrategyCommand) (*StopStrategyResult, error) {
	ctx, span := s.tracer.Start(ctx, "StrategyApplicationService.StopStrategy")
	defer span.End()

	span.SetAttributes(attribute.String("strategy_id", cmd.StrategyID))

	// Find strategy
	strategy, err := s.strategyRepo.FindByID(ctx, entities.StrategyID(cmd.StrategyID))
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to find strategy: %w", err)
	}

	// Stop strategy
	if err := strategy.Stop(); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to stop strategy: %w", err)
	}

	// Update strategy in repository
	if err := s.strategyRepo.Update(ctx, strategy); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to update strategy: %w", err)
	}

	// Save strategy events to event store
	if err := s.eventStore.SaveEvents(ctx, string(strategy.GetID()), s.convertStrategyEventsToGeneric(strategy.GetEvents())); err != nil {
		// Log error but don't fail the operation
		span.AddEvent("failed to save strategy events", trace.WithAttributes(
			attribute.String("error", err.Error()),
		))
	}

	span.SetAttributes(
		attribute.String("strategy_status", string(strategy.GetStatus())),
	)

	return &StopStrategyResult{
		StrategyID: string(strategy.GetID()),
		Status:     string(strategy.GetStatus()),
		StoppedAt:  *strategy.GetStoppedAt(),
		Message:    "Strategy stopped successfully",
	}, nil
}

// GetStrategyQuery represents a query to get strategy details
type GetStrategyQuery struct {
	StrategyID string `json:"strategy_id" validate:"required"`
}

// StrategyDTO represents strategy data transfer object
type StrategyDTO struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Status      string                 `json:"status"`
	Symbols     []string               `json:"symbols"`
	Exchanges   []string               `json:"exchanges"`
	Parameters  map[string]interface{} `json:"parameters"`
	RiskLimits  RiskLimitsDTO          `json:"risk_limits"`
	Performance PerformanceDTO         `json:"performance"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	StoppedAt   *time.Time             `json:"stopped_at,omitempty"`
}

// RiskLimitsDTO represents risk limits data transfer object
type RiskLimitsDTO struct {
	MaxPositionSize    string `json:"max_position_size"`
	MaxDailyLoss       string `json:"max_daily_loss"`
	MaxDrawdown        string `json:"max_drawdown"`
	MaxOrderSize       string `json:"max_order_size"`
	MaxOrdersPerSecond int    `json:"max_orders_per_second"`
	MaxExposure        string `json:"max_exposure"`
	StopLossPercent    string `json:"stop_loss_percent"`
	TakeProfitPercent  string `json:"take_profit_percent"`
}

// PerformanceDTO represents performance data transfer object
type PerformanceDTO struct {
	TotalPnL      string `json:"total_pnl"`
	DailyPnL      string `json:"daily_pnl"`
	TotalTrades   int64  `json:"total_trades"`
	WinningTrades int64  `json:"winning_trades"`
	LosingTrades  int64  `json:"losing_trades"`
	WinRate       string `json:"win_rate"`
	AvgWin        string `json:"avg_win"`
	AvgLoss       string `json:"avg_loss"`
	ProfitFactor  string `json:"profit_factor"`
	SharpeRatio   string `json:"sharpe_ratio"`
	MaxDrawdown   string `json:"max_drawdown"`
	VolumeTraded  string `json:"volume_traded"`
	AvgLatency    string `json:"avg_latency"`
	LastUpdated   time.Time `json:"last_updated"`
}

// GetStrategy handles the get strategy query
func (s *StrategyApplicationService) GetStrategy(ctx context.Context, query GetStrategyQuery) (*StrategyDTO, error) {
	ctx, span := s.tracer.Start(ctx, "StrategyApplicationService.GetStrategy")
	defer span.End()

	span.SetAttributes(attribute.String("strategy_id", query.StrategyID))

	// Find strategy
	strategy, err := s.strategyRepo.FindByID(ctx, entities.StrategyID(query.StrategyID))
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to find strategy: %w", err)
	}

	// Convert to DTO
	dto := s.strategyToDTO(strategy)
	return dto, nil
}

// ListStrategiesQuery represents a query to list strategies
type ListStrategiesQuery struct {
	Status *string `json:"status,omitempty"`
	Type   *string `json:"type,omitempty"`
	Limit  int     `json:"limit"`
	Offset int     `json:"offset"`
}

// ListStrategiesResult represents the result of listing strategies
type ListStrategiesResult struct {
	Strategies []StrategyDTO `json:"strategies"`
	Total      int           `json:"total"`
	Limit      int           `json:"limit"`
	Offset     int           `json:"offset"`
}

// ListStrategies handles the list strategies query
func (s *StrategyApplicationService) ListStrategies(ctx context.Context, query ListStrategiesQuery) (*ListStrategiesResult, error) {
	ctx, span := s.tracer.Start(ctx, "StrategyApplicationService.ListStrategies")
	defer span.End()

	var strategies []*entities.Strategy
	var err error

	// Query based on filters
	if query.Status != nil && query.Type != nil {
		strategies, err = s.strategyRepo.FindStrategiesByTypeAndStatus(
			ctx,
			valueobjects.StrategyType(*query.Type),
			valueobjects.StrategyStatus(*query.Status),
		)
	} else if query.Status != nil {
		strategies, err = s.strategyRepo.FindByStatus(ctx, valueobjects.StrategyStatus(*query.Status))
	} else if query.Type != nil {
		strategies, err = s.strategyRepo.FindByType(ctx, valueobjects.StrategyType(*query.Type))
	} else {
		strategies, err = s.strategyRepo.FindAll(ctx)
	}

	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to find strategies: %w", err)
	}

	// Convert to DTOs
	dtos := make([]StrategyDTO, len(strategies))
	for i, strategy := range strategies {
		dtos[i] = *s.strategyToDTO(strategy)
	}

	span.SetAttributes(
		attribute.Int("strategies_count", len(dtos)),
	)

	return &ListStrategiesResult{
		Strategies: dtos,
		Total:      len(dtos),
		Limit:      query.Limit,
		Offset:     query.Offset,
	}, nil
}

// Helper methods

// parseRiskLimits parses risk limits from command
func (s *StrategyApplicationService) parseRiskLimits(cmd CreateStrategyCommand) (valueobjects.RiskLimits, error) {
	maxPositionSize, err := valueobjects.NewPriceFromString(cmd.MaxPositionSize)
	if err != nil {
		return valueobjects.RiskLimits{}, fmt.Errorf("invalid max position size: %w", err)
	}

	maxDailyLoss, err := valueobjects.NewPriceFromString(cmd.MaxDailyLoss)
	if err != nil {
		return valueobjects.RiskLimits{}, fmt.Errorf("invalid max daily loss: %w", err)
	}

	maxDrawdown, err := valueobjects.NewPriceFromString(cmd.MaxDrawdown)
	if err != nil {
		return valueobjects.RiskLimits{}, fmt.Errorf("invalid max drawdown: %w", err)
	}

	maxOrderSize, err := valueobjects.NewPriceFromString(cmd.MaxOrderSize)
	if err != nil {
		return valueobjects.RiskLimits{}, fmt.Errorf("invalid max order size: %w", err)
	}

	maxExposure, err := valueobjects.NewPriceFromString(cmd.MaxExposure)
	if err != nil {
		return valueobjects.RiskLimits{}, fmt.Errorf("invalid max exposure: %w", err)
	}

	stopLossPercent, err := valueobjects.NewPriceFromString(cmd.StopLossPercent)
	if err != nil {
		return valueobjects.RiskLimits{}, fmt.Errorf("invalid stop loss percent: %w", err)
	}

	takeProfitPercent, err := valueobjects.NewPriceFromString(cmd.TakeProfitPercent)
	if err != nil {
		return valueobjects.RiskLimits{}, fmt.Errorf("invalid take profit percent: %w", err)
	}

	return valueobjects.NewRiskLimits(
		maxPositionSize.Decimal,
		maxDailyLoss.Decimal,
		maxDrawdown.Decimal,
		maxOrderSize.Decimal,
		maxExposure.Decimal,
		cmd.MaxOrdersPerSecond,
		stopLossPercent.Decimal,
		takeProfitPercent.Decimal,
	)
}

// strategyToDTO converts strategy entity to DTO
func (s *StrategyApplicationService) strategyToDTO(strategy *entities.Strategy) *StrategyDTO {
	symbols := make([]string, len(strategy.GetSymbols()))
	for i, symbol := range strategy.GetSymbols() {
		symbols[i] = string(symbol)
	}

	exchanges := make([]string, len(strategy.GetExchanges()))
	for i, exchange := range strategy.GetExchanges() {
		exchanges[i] = string(exchange)
	}

	riskLimits := strategy.GetRiskLimits()
	performance := strategy.GetPerformance()

	return &StrategyDTO{
		ID:        string(strategy.GetID()),
		Name:      strategy.GetName(),
		Type:      string(strategy.GetType()),
		Status:    string(strategy.GetStatus()),
		Symbols:   symbols,
		Exchanges: exchanges,
		Parameters: strategy.GetParameters(),
		RiskLimits: RiskLimitsDTO{
			MaxPositionSize:    riskLimits.MaxPositionSize.String(),
			MaxDailyLoss:       riskLimits.MaxDailyLoss.String(),
			MaxDrawdown:        riskLimits.MaxDrawdown.String(),
			MaxOrderSize:       riskLimits.MaxOrderSize.String(),
			MaxOrdersPerSecond: riskLimits.MaxOrdersPerSecond,
			MaxExposure:        riskLimits.MaxExposure.String(),
			StopLossPercent:    riskLimits.StopLossPercent.String(),
			TakeProfitPercent:  riskLimits.TakeProfitPercent.String(),
		},
		Performance: PerformanceDTO{
			TotalPnL:      performance.TotalPnL.String(),
			DailyPnL:      performance.DailyPnL.String(),
			TotalTrades:   performance.TotalTrades,
			WinningTrades: performance.WinningTrades,
			LosingTrades:  performance.LosingTrades,
			WinRate:       performance.WinRate.String(),
			AvgWin:        performance.AvgWin.String(),
			AvgLoss:       performance.AvgLoss.String(),
			ProfitFactor:  performance.ProfitFactor.String(),
			SharpeRatio:   performance.SharpeRatio.String(),
			MaxDrawdown:   performance.MaxDrawdown.String(),
			VolumeTraded:  performance.VolumeTraded.String(),
			AvgLatency:    performance.AvgLatency.String(),
			LastUpdated:   performance.LastUpdated,
		},
		CreatedAt: strategy.GetCreatedAt(),
		UpdatedAt: strategy.GetUpdatedAt(),
		StartedAt: strategy.GetStartedAt(),
		StoppedAt: strategy.GetStoppedAt(),
	}
}

// convertStrategyEventsToGeneric converts strategy events to generic domain events
func (s *StrategyApplicationService) convertStrategyEventsToGeneric(events []valueobjects.StrategyEvent) []repositories.DomainEvent {
	genericEvents := make([]repositories.DomainEvent, len(events))
	for i, event := range events {
		genericEvents[i] = repositories.DomainEvent{
			ID:          fmt.Sprintf("strategy_event_%d", i),
			EventType:   string(event.Type),
			EventData:   event.Data,
			Timestamp:   event.Timestamp,
			Version:     i + 1,
		}
	}
	return genericEvents
}
