package recovery

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go-coffee-ai-agents/internal/resilience/errors"
)

// RecoveryStrategy defines how to recover from failures
type RecoveryStrategy interface {
	// Recover attempts to recover from a failure
	Recover(ctx context.Context, err error) error
	
	// CanRecover determines if this strategy can handle the given error
	CanRecover(err error) bool
	
	// Name returns the name of the strategy
	Name() string
}

// FallbackStrategy defines fallback behavior when primary operations fail
type FallbackStrategy interface {
	// Execute executes the fallback operation
	Execute(ctx context.Context) (interface{}, error)
	
	// CanFallback determines if fallback is available for the given error
	CanFallback(err error) bool
	
	// Name returns the name of the fallback strategy
	Name() string
}

// RecoveryConfig holds configuration for recovery strategies
type RecoveryConfig struct {
	// Maximum number of recovery attempts
	MaxAttempts int
	
	// Delay between recovery attempts
	RecoveryDelay time.Duration
	
	// Whether to enable automatic recovery
	AutoRecovery bool
	
	// Strategies to try in order
	Strategies []RecoveryStrategy
	
	// Callbacks
	OnRecoveryAttempt func(strategy string, attempt int, err error)
	OnRecoverySuccess func(strategy string, attempt int)
	OnRecoveryFailure func(strategy string, attempts int, err error)
}

// DefaultRecoveryConfig returns a default recovery configuration
func DefaultRecoveryConfig() RecoveryConfig {
	return RecoveryConfig{
		MaxAttempts:   3,
		RecoveryDelay: 5 * time.Second,
		AutoRecovery:  true,
		Strategies:    []RecoveryStrategy{},
	}
}

// RecoveryManager manages recovery strategies
type RecoveryManager struct {
	config     RecoveryConfig
	strategies map[string]RecoveryStrategy
	mutex      sync.RWMutex
}

// NewRecoveryManager creates a new recovery manager
func NewRecoveryManager(config RecoveryConfig) *RecoveryManager {
	return &RecoveryManager{
		config:     config,
		strategies: make(map[string]RecoveryStrategy),
	}
}

// RegisterStrategy registers a recovery strategy
func (rm *RecoveryManager) RegisterStrategy(strategy RecoveryStrategy) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	rm.strategies[strategy.Name()] = strategy
}

// Recover attempts to recover from an error using registered strategies
func (rm *RecoveryManager) Recover(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}

	rm.mutex.RLock()
	strategies := make([]RecoveryStrategy, 0, len(rm.strategies))
	for _, strategy := range rm.strategies {
		if strategy.CanRecover(err) {
			strategies = append(strategies, strategy)
		}
	}
	rm.mutex.RUnlock()

	if len(strategies) == 0 {
		return errors.NewError(errors.CodeInternalError, "No recovery strategy available").
			WithCause(err).
			Build()
	}

	// Try each strategy
	for _, strategy := range strategies {
		for attempt := 1; attempt <= rm.config.MaxAttempts; attempt++ {
			if rm.config.OnRecoveryAttempt != nil {
				rm.config.OnRecoveryAttempt(strategy.Name(), attempt, err)
			}

			recoveryErr := strategy.Recover(ctx, err)
			if recoveryErr == nil {
				if rm.config.OnRecoverySuccess != nil {
					rm.config.OnRecoverySuccess(strategy.Name(), attempt)
				}
				return nil
			}

			// Wait before next attempt
			if attempt < rm.config.MaxAttempts {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(rm.config.RecoveryDelay):
				}
			}
		}

		if rm.config.OnRecoveryFailure != nil {
			rm.config.OnRecoveryFailure(strategy.Name(), rm.config.MaxAttempts, err)
		}
	}

	return errors.NewError(errors.CodeInternalError, "All recovery strategies failed").
		WithCause(err).
		Build()
}

// RestartStrategy implements a restart-based recovery strategy
type RestartStrategy struct {
	name        string
	restartFunc func(ctx context.Context) error
	canRecover  func(error) bool
}

// NewRestartStrategy creates a new restart strategy
func NewRestartStrategy(name string, restartFunc func(ctx context.Context) error, canRecover func(error) bool) *RestartStrategy {
	return &RestartStrategy{
		name:        name,
		restartFunc: restartFunc,
		canRecover:  canRecover,
	}
}

// Recover attempts to recover by restarting
func (rs *RestartStrategy) Recover(ctx context.Context, err error) error {
	return rs.restartFunc(ctx)
}

// CanRecover determines if this strategy can handle the error
func (rs *RestartStrategy) CanRecover(err error) bool {
	if rs.canRecover != nil {
		return rs.canRecover(err)
	}
	return true
}

// Name returns the strategy name
func (rs *RestartStrategy) Name() string {
	return rs.name
}

// ReconnectStrategy implements a reconnection-based recovery strategy
type ReconnectStrategy struct {
	name          string
	reconnectFunc func(ctx context.Context) error
	canRecover    func(error) bool
}

// NewReconnectStrategy creates a new reconnect strategy
func NewReconnectStrategy(name string, reconnectFunc func(ctx context.Context) error, canRecover func(error) bool) *ReconnectStrategy {
	return &ReconnectStrategy{
		name:          name,
		reconnectFunc: reconnectFunc,
		canRecover:    canRecover,
	}
}

// Recover attempts to recover by reconnecting
func (rs *ReconnectStrategy) Recover(ctx context.Context, err error) error {
	return rs.reconnectFunc(ctx)
}

// CanRecover determines if this strategy can handle the error
func (rs *ReconnectStrategy) CanRecover(err error) bool {
	if rs.canRecover != nil {
		return rs.canRecover(err)
	}
	
	// Default: handle network and connection errors
	category := errors.GetErrorCategory(err)
	return category == errors.CategoryNetwork || category == errors.CategoryDatabase
}

// Name returns the strategy name
func (rs *ReconnectStrategy) Name() string {
	return rs.name
}

// CacheInvalidationStrategy implements cache invalidation recovery
type CacheInvalidationStrategy struct {
	name           string
	invalidateFunc func(ctx context.Context, key string) error
	extractKey     func(error) string
}

// NewCacheInvalidationStrategy creates a new cache invalidation strategy
func NewCacheInvalidationStrategy(name string, invalidateFunc func(ctx context.Context, key string) error, extractKey func(error) string) *CacheInvalidationStrategy {
	return &CacheInvalidationStrategy{
		name:           name,
		invalidateFunc: invalidateFunc,
		extractKey:     extractKey,
	}
}

// Recover attempts to recover by invalidating cache
func (cis *CacheInvalidationStrategy) Recover(ctx context.Context, err error) error {
	key := cis.extractKey(err)
	if key == "" {
		return fmt.Errorf("cannot extract cache key from error")
	}
	return cis.invalidateFunc(ctx, key)
}

// CanRecover determines if this strategy can handle the error
func (cis *CacheInvalidationStrategy) CanRecover(err error) bool {
	category := errors.GetErrorCategory(err)
	return category == errors.CategoryCache
}

// Name returns the strategy name
func (cis *CacheInvalidationStrategy) Name() string {
	return cis.name
}

// FallbackManager manages fallback strategies
type FallbackManager struct {
	fallbacks map[string]FallbackStrategy
	mutex     sync.RWMutex
}

// NewFallbackManager creates a new fallback manager
func NewFallbackManager() *FallbackManager {
	return &FallbackManager{
		fallbacks: make(map[string]FallbackStrategy),
	}
}

// RegisterFallback registers a fallback strategy
func (fm *FallbackManager) RegisterFallback(name string, fallback FallbackStrategy) {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()
	fm.fallbacks[name] = fallback
}

// ExecuteWithFallback executes a primary function with fallback support
func (fm *FallbackManager) ExecuteWithFallback(ctx context.Context, primary func() (interface{}, error), fallbackName string) (interface{}, error) {
	// Try primary function first
	result, err := primary()
	if err == nil {
		return result, nil
	}

	// Try fallback
	fm.mutex.RLock()
	fallback, exists := fm.fallbacks[fallbackName]
	fm.mutex.RUnlock()

	if !exists {
		return nil, errors.NewError(errors.CodeInternalError, "Fallback strategy not found").
			WithCause(err).
			WithDetail("fallback_name", fallbackName).
			Build()
	}

	if !fallback.CanFallback(err) {
		return nil, errors.NewError(errors.CodeInternalError, "Fallback strategy cannot handle error").
			WithCause(err).
			WithDetail("fallback_name", fallbackName).
			Build()
	}

	return fallback.Execute(ctx)
}

// CacheFallback implements cache-based fallback
type CacheFallback struct {
	name     string
	getCache func(ctx context.Context, key string) (interface{}, error)
	keyFunc  func(error) string
}

// NewCacheFallback creates a new cache fallback strategy
func NewCacheFallback(name string, getCache func(ctx context.Context, key string) (interface{}, error), keyFunc func(error) string) *CacheFallback {
	return &CacheFallback{
		name:     name,
		getCache: getCache,
		keyFunc:  keyFunc,
	}
}

// Execute executes the cache fallback
func (cf *CacheFallback) Execute(ctx context.Context) (interface{}, error) {
	// This is a simplified implementation
	// In practice, you'd need to pass the error context
	return nil, fmt.Errorf("cache fallback requires error context")
}

// ExecuteWithError executes the cache fallback with error context
func (cf *CacheFallback) ExecuteWithError(ctx context.Context, err error) (interface{}, error) {
	key := cf.keyFunc(err)
	if key == "" {
		return nil, fmt.Errorf("cannot extract cache key from error")
	}
	return cf.getCache(ctx, key)
}

// CanFallback determines if cache fallback is available
func (cf *CacheFallback) CanFallback(err error) bool {
	// Can fallback for most errors except cache errors themselves
	category := errors.GetErrorCategory(err)
	return category != errors.CategoryCache
}

// Name returns the fallback name
func (cf *CacheFallback) Name() string {
	return cf.name
}

// DefaultValueFallback implements default value fallback
type DefaultValueFallback struct {
	name         string
	defaultValue interface{}
	canFallback  func(error) bool
}

// NewDefaultValueFallback creates a new default value fallback
func NewDefaultValueFallback(name string, defaultValue interface{}, canFallback func(error) bool) *DefaultValueFallback {
	return &DefaultValueFallback{
		name:         name,
		defaultValue: defaultValue,
		canFallback:  canFallback,
	}
}

// Execute returns the default value
func (dvf *DefaultValueFallback) Execute(ctx context.Context) (interface{}, error) {
	return dvf.defaultValue, nil
}

// CanFallback determines if default value fallback is available
func (dvf *DefaultValueFallback) CanFallback(err error) bool {
	if dvf.canFallback != nil {
		return dvf.canFallback(err)
	}
	return true
}

// Name returns the fallback name
func (dvf *DefaultValueFallback) Name() string {
	return dvf.name
}

// ServiceFallback implements service-to-service fallback
type ServiceFallback struct {
	name           string
	fallbackFunc   func(ctx context.Context) (interface{}, error)
	canFallback    func(error) bool
}

// NewServiceFallback creates a new service fallback
func NewServiceFallback(name string, fallbackFunc func(ctx context.Context) (interface{}, error), canFallback func(error) bool) *ServiceFallback {
	return &ServiceFallback{
		name:         name,
		fallbackFunc: fallbackFunc,
		canFallback:  canFallback,
	}
}

// Execute executes the service fallback
func (sf *ServiceFallback) Execute(ctx context.Context) (interface{}, error) {
	return sf.fallbackFunc(ctx)
}

// CanFallback determines if service fallback is available
func (sf *ServiceFallback) CanFallback(err error) bool {
	if sf.canFallback != nil {
		return sf.canFallback(err)
	}
	
	// Default: fallback for external service errors
	category := errors.GetErrorCategory(err)
	return category == errors.CategoryExternal || category == errors.CategoryAI
}

// Name returns the fallback name
func (sf *ServiceFallback) Name() string {
	return sf.name
}

// GracefulDegradationManager manages graceful degradation strategies
type GracefulDegradationManager struct {
	degradationLevels map[string]int
	currentLevel      int
	maxLevel          int
	mutex             sync.RWMutex
}

// NewGracefulDegradationManager creates a new graceful degradation manager
func NewGracefulDegradationManager(maxLevel int) *GracefulDegradationManager {
	return &GracefulDegradationManager{
		degradationLevels: make(map[string]int),
		currentLevel:      0,
		maxLevel:          maxLevel,
	}
}

// SetDegradationLevel sets the degradation level for a component
func (gdm *GracefulDegradationManager) SetDegradationLevel(component string, level int) {
	gdm.mutex.Lock()
	defer gdm.mutex.Unlock()
	
	gdm.degradationLevels[component] = level
	
	// Update current level to the maximum
	maxLevel := 0
	for _, l := range gdm.degradationLevels {
		if l > maxLevel {
			maxLevel = l
		}
	}
	gdm.currentLevel = maxLevel
}

// GetDegradationLevel returns the current degradation level
func (gdm *GracefulDegradationManager) GetDegradationLevel() int {
	gdm.mutex.RLock()
	defer gdm.mutex.RUnlock()
	return gdm.currentLevel
}

// ShouldDegrade determines if a feature should be degraded
func (gdm *GracefulDegradationManager) ShouldDegrade(featureLevel int) bool {
	return gdm.GetDegradationLevel() >= featureLevel
}

// Global recovery and fallback managers
var (
	globalRecoveryManager *RecoveryManager
	globalFallbackManager *FallbackManager
)

// InitGlobalManagers initializes global recovery and fallback managers
func InitGlobalManagers(recoveryConfig RecoveryConfig) {
	globalRecoveryManager = NewRecoveryManager(recoveryConfig)
	globalFallbackManager = NewFallbackManager()
}

// GetGlobalRecoveryManager returns the global recovery manager
func GetGlobalRecoveryManager() *RecoveryManager {
	if globalRecoveryManager == nil {
		globalRecoveryManager = NewRecoveryManager(DefaultRecoveryConfig())
	}
	return globalRecoveryManager
}

// GetGlobalFallbackManager returns the global fallback manager
func GetGlobalFallbackManager() *FallbackManager {
	if globalFallbackManager == nil {
		globalFallbackManager = NewFallbackManager()
	}
	return globalFallbackManager
}
