package ai

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"go-coffee-ai-agents/internal/common"
)

// Ensure all strategies implement the ProviderStrategy interface
var (
	_ common.ProviderStrategy = (*RoundRobinStrategy)(nil)
	_ common.ProviderStrategy = (*RandomStrategy)(nil)
	_ common.ProviderStrategy = (*CostOptimizedStrategy)(nil)
	_ common.ProviderStrategy = (*PerformanceOptimizedStrategy)(nil)
	_ common.ProviderStrategy = (*WeightedStrategy)(nil)
	_ common.ProviderStrategy = (*FailoverStrategy)(nil)
	_ common.ProviderStrategy = (*AdaptiveStrategy)(nil)
	_ common.ProviderStrategy = (*CompositeStrategy)(nil)
)

// RoundRobinStrategy implements round-robin provider selection
type RoundRobinStrategy struct {
	counters map[common.ModelType]int
	mutex    sync.Mutex
}

// NewRoundRobinStrategy creates a new round-robin strategy
func NewRoundRobinStrategy() *RoundRobinStrategy {
	return &RoundRobinStrategy{
		counters: make(map[common.ModelType]int),
	}
}

// SelectProvider selects a provider using round-robin algorithm
func (s *RoundRobinStrategy) SelectProvider(providers []common.Provider, modelType common.ModelType) (common.Provider, error) {
	if len(providers) == 0 {
		return nil, fmt.Errorf("no providers available")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	counter := s.counters[modelType]
	selected := providers[counter%len(providers)]
	s.counters[modelType] = (counter + 1) % len(providers)

	return selected, nil
}

// RandomStrategy implements random provider selection
type RandomStrategy struct {
	rand *rand.Rand
}

// NewRandomStrategy creates a new random strategy
func NewRandomStrategy() *RandomStrategy {
	return &RandomStrategy{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// SelectProvider selects a provider randomly
func (s *RandomStrategy) SelectProvider(providers []common.Provider, modelType common.ModelType) (common.Provider, error) {
	if len(providers) == 0 {
		return nil, fmt.Errorf("no providers available")
	}

	index := s.rand.Intn(len(providers))
	return providers[index], nil
}

// CostOptimizedStrategy selects providers based on cost optimization
type CostOptimizedStrategy struct{}

// NewCostOptimizedStrategy creates a new cost-optimized strategy
func NewCostOptimizedStrategy() *CostOptimizedStrategy {
	return &CostOptimizedStrategy{}
}

// SelectProvider selects the most cost-effective provider
func (s *CostOptimizedStrategy) SelectProvider(providers []common.Provider, modelType common.ModelType) (common.Provider, error) {
	if len(providers) == 0 {
		return nil, fmt.Errorf("no providers available")
	}

	var bestProvider common.Provider
	var lowestCost float64 = -1

	for _, provider := range providers {
		models := provider.GetModels()
		for _, model := range models {
			if model.Type == modelType {
				// Calculate average cost (input + output)
				avgCost := (model.InputCost + model.OutputCost) / 2
				if lowestCost == -1 || avgCost < lowestCost {
					lowestCost = avgCost
					bestProvider = provider
				}
				break
			}
		}
	}

	if bestProvider == nil {
		return nil, fmt.Errorf("no suitable provider found for model type %s", modelType)
	}

	return bestProvider, nil
}

// PerformanceOptimizedStrategy selects providers based on performance metrics
type PerformanceOptimizedStrategy struct{}

// NewPerformanceOptimizedStrategy creates a new performance-optimized strategy
func NewPerformanceOptimizedStrategy() *PerformanceOptimizedStrategy {
	return &PerformanceOptimizedStrategy{}
}

// SelectProvider selects the best performing provider
func (s *PerformanceOptimizedStrategy) SelectProvider(providers []common.Provider, modelType common.ModelType) (common.Provider, error) {
	if len(providers) == 0 {
		return nil, fmt.Errorf("no providers available")
	}

	var bestProvider common.Provider
	var bestScore float64 = -1

	for _, provider := range providers {
		// Check if provider supports the model type
		supportsType := false
		models := provider.GetModels()
		for _, model := range models {
			if model.Type == modelType {
				supportsType = true
				break
			}
		}

		if !supportsType {
			continue
		}

		// Calculate performance score based on usage statistics
		usage := provider.GetUsage()
		if usage.TotalRequests == 0 {
			// New provider, give it a chance
			if bestProvider == nil {
				bestProvider = provider
			}
			continue
		}

		// Calculate success rate
		successRate := float64(usage.SuccessfulReqs) / float64(usage.TotalRequests)

		// Calculate latency score (lower is better, so invert)
		latencyScore := 1.0 / (usage.AverageLatency.Seconds() + 0.001) // Add small value to avoid division by zero

		// Combined score (weighted)
		score := (successRate * 0.7) + (latencyScore * 0.3)

		if score > bestScore {
			bestScore = score
			bestProvider = provider
		}
	}

	if bestProvider == nil {
		return nil, fmt.Errorf("no suitable provider found for model type %s", modelType)
	}

	return bestProvider, nil
}

// WeightedStrategy implements weighted provider selection
type WeightedStrategy struct {
	weights map[string]float64
	rand    *rand.Rand
}

// NewWeightedStrategy creates a new weighted strategy
func NewWeightedStrategy(weights map[string]float64) *WeightedStrategy {
	return &WeightedStrategy{
		weights: weights,
		rand:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// SelectProvider selects a provider based on weights
func (s *WeightedStrategy) SelectProvider(providers []common.Provider, modelType common.ModelType) (common.Provider, error) {
	if len(providers) == 0 {
		return nil, fmt.Errorf("no providers available")
	}

	// Filter providers that support the model type
	var supportedProviders []common.Provider
	var weights []float64
	var totalWeight float64

	for _, provider := range providers {
		models := provider.GetModels()
		supportsType := false
		for _, model := range models {
			if model.Type == modelType {
				supportsType = true
				break
			}
		}

		if supportsType {
			supportedProviders = append(supportedProviders, provider)
			weight := s.weights[provider.GetName()]
			if weight <= 0 {
				weight = 1.0 // Default weight
			}
			weights = append(weights, weight)
			totalWeight += weight
		}
	}

	if len(supportedProviders) == 0 {
		return nil, fmt.Errorf("no providers support model type %s", modelType)
	}

	// Select based on weighted random
	target := s.rand.Float64() * totalWeight
	var current float64

	for i, weight := range weights {
		current += weight
		if current >= target {
			return supportedProviders[i], nil
		}
	}

	// Fallback to last provider
	return supportedProviders[len(supportedProviders)-1], nil
}

// FailoverStrategy implements failover provider selection
type FailoverStrategy struct {
	primaryProvider   string
	fallbackProviders []string
}

// NewFailoverStrategy creates a new failover strategy
func NewFailoverStrategy(primary string, fallbacks []string) *FailoverStrategy {
	return &FailoverStrategy{
		primaryProvider:   primary,
		fallbackProviders: fallbacks,
	}
}

// SelectProvider selects a provider using failover logic
func (s *FailoverStrategy) SelectProvider(providers []common.Provider, modelType common.ModelType) (common.Provider, error) {
	if len(providers) == 0 {
		return nil, fmt.Errorf("no providers available")
	}

	// Create provider map for quick lookup
	providerMap := make(map[string]common.Provider)
	for _, provider := range providers {
		models := provider.GetModels()
		supportsType := false
		for _, model := range models {
			if model.Type == modelType {
				supportsType = true
				break
			}
		}
		if supportsType {
			providerMap[provider.GetName()] = provider
		}
	}

	// Try primary provider first
	if primary, exists := providerMap[s.primaryProvider]; exists {
		// Check if primary is healthy (simplified check)
		usage := primary.GetUsage()
		if usage.TotalRequests == 0 || float64(usage.SuccessfulReqs)/float64(usage.TotalRequests) > 0.8 {
			return primary, nil
		}
	}

	// Try fallback providers
	for _, fallbackName := range s.fallbackProviders {
		if fallback, exists := providerMap[fallbackName]; exists {
			return fallback, nil
		}
	}

	// If no configured providers are available, use any available provider
	for _, provider := range providerMap {
		return provider, nil
	}

	return nil, fmt.Errorf("no suitable provider found for model type %s", modelType)
}

// AdaptiveStrategy implements adaptive provider selection based on real-time metrics
type AdaptiveStrategy struct {
	recentWindow time.Duration
	mutex        sync.RWMutex
}

// NewAdaptiveStrategy creates a new adaptive strategy
func NewAdaptiveStrategy(recentWindow time.Duration) *AdaptiveStrategy {
	return &AdaptiveStrategy{
		recentWindow: recentWindow,
	}
}

// SelectProvider selects a provider based on recent performance
func (s *AdaptiveStrategy) SelectProvider(providers []common.Provider, modelType common.ModelType) (common.Provider, error) {
	if len(providers) == 0 {
		return nil, fmt.Errorf("no providers available")
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var bestProvider common.Provider
	var bestScore float64 = -1

	for _, provider := range providers {
		// Check if provider supports the model type
		supportsType := false
		models := provider.GetModels()
		for _, model := range models {
			if model.Type == modelType {
				supportsType = true
				break
			}
		}

		if !supportsType {
			continue
		}

		// Calculate adaptive score based on recent performance
		usage := provider.GetUsage()

		// If no recent requests, give it a chance
		if usage.TotalRequests == 0 || time.Since(usage.LastRequestTime) > s.recentWindow {
			if bestProvider == nil {
				bestProvider = provider
			}
			continue
		}

		// Calculate score based on recent success rate and latency
		successRate := float64(usage.SuccessfulReqs) / float64(usage.TotalRequests)
		latencyPenalty := usage.AverageLatency.Seconds() / 10.0 // Normalize latency

		// Adaptive score considers recency
		recencyFactor := 1.0
		if time.Since(usage.LastRequestTime) < s.recentWindow {
			recencyFactor = 1.2 // Boost for recent activity
		}

		score := (successRate - latencyPenalty) * recencyFactor

		if score > bestScore {
			bestScore = score
			bestProvider = provider
		}
	}

	if bestProvider == nil {
		return nil, fmt.Errorf("no suitable provider found for model type %s", modelType)
	}

	return bestProvider, nil
}

// CompositeStrategy combines multiple strategies
type CompositeStrategy struct {
	strategies []common.ProviderStrategy
	weights    []float64
	rand       *rand.Rand
}

// NewCompositeStrategy creates a new composite strategy
func NewCompositeStrategy(strategies []common.ProviderStrategy, weights []float64) *CompositeStrategy {
	if len(strategies) != len(weights) {
		panic("strategies and weights must have the same length")
	}

	return &CompositeStrategy{
		strategies: strategies,
		weights:    weights,
		rand:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// SelectProvider selects a provider using composite strategy
func (s *CompositeStrategy) SelectProvider(providers []common.Provider, modelType common.ModelType) (common.Provider, error) {
	if len(providers) == 0 {
		return nil, fmt.Errorf("no providers available")
	}

	if len(s.strategies) == 0 {
		return providers[0], nil
	}

	// Select strategy based on weights
	totalWeight := 0.0
	for _, weight := range s.weights {
		totalWeight += weight
	}

	target := s.rand.Float64() * totalWeight
	current := 0.0

	for i, weight := range s.weights {
		current += weight
		if current >= target {
			return s.strategies[i].SelectProvider(providers, modelType)
		}
	}

	// Fallback to first strategy
	return s.strategies[0].SelectProvider(providers, modelType)
}
