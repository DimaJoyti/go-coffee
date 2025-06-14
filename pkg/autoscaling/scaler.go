package autoscaling

import (
	"context"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// AutoScaler provides intelligent auto-scaling capabilities
type AutoScaler struct {
	logger         *zap.Logger
	config         *AutoScalerConfig
	metrics        *ScalingMetrics
	predictors     map[string]*LoadPredictor
	scalingActions chan *ScalingAction
	running        int32 // atomic boolean
	mu             sync.RWMutex
}

// AutoScalerConfig contains auto-scaling configuration
type AutoScalerConfig struct {
	Enabled                 bool                     `json:"enabled"`
	MinReplicas             int32                    `json:"min_replicas"`
	MaxReplicas             int32                    `json:"max_replicas"`
	TargetCPUUtilization    float64                  `json:"target_cpu_utilization"`
	TargetMemoryUtilization float64                  `json:"target_memory_utilization"`
	ScaleUpCooldown         time.Duration            `json:"scale_up_cooldown"`
	ScaleDownCooldown       time.Duration            `json:"scale_down_cooldown"`
	MetricsWindow           time.Duration            `json:"metrics_window"`
	EvaluationInterval      time.Duration            `json:"evaluation_interval"`
	PredictiveScaling       bool                     `json:"predictive_scaling"`
	CustomMetrics           map[string]*CustomMetric `json:"custom_metrics"`
	ScalingPolicies         []*ScalingPolicy         `json:"scaling_policies"`
}

// CustomMetric defines a custom scaling metric
type CustomMetric struct {
	Name           string  `json:"name"`
	TargetValue    float64 `json:"target_value"`
	Weight         float64 `json:"weight"`
	ScaleDirection string  `json:"scale_direction"` // "up", "down", "both"
}

// ScalingPolicy defines scaling behavior
type ScalingPolicy struct {
	Type                string        `json:"type"`                 // "Pods", "Percent"
	Value               int32         `json:"value"`                // Number of pods or percentage
	PeriodSeconds       int32         `json:"period_seconds"`       // Time window for scaling
	StabilizationWindow time.Duration `json:"stabilization_window"` // Stabilization period
}

// ScalingMetrics tracks scaling performance
type ScalingMetrics struct {
	CurrentReplicas    int32                      `json:"current_replicas"`
	DesiredReplicas    int32                      `json:"desired_replicas"`
	LastScaleTime      time.Time                  `json:"last_scale_time"`
	ScaleUpCount       int64                      `json:"scale_up_count"`
	ScaleDownCount     int64                      `json:"scale_down_count"`
	CPUUtilization     float64                    `json:"cpu_utilization"`
	MemoryUtilization  float64                    `json:"memory_utilization"`
	CustomMetricValues map[string]float64         `json:"custom_metric_values"`
	PredictedLoad      map[string]*LoadPrediction `json:"predicted_load"`
	ScalingHistory     []*ScalingEvent            `json:"scaling_history"`
	mu                 sync.RWMutex
}

// ScalingAction represents a scaling action to be performed
type ScalingAction struct {
	Type           string             `json:"type"` // "scale_up", "scale_down"
	TargetReplicas int32              `json:"target_replicas"`
	Reason         string             `json:"reason"`
	Metrics        map[string]float64 `json:"metrics"`
	Timestamp      time.Time          `json:"timestamp"`
	Predictive     bool               `json:"predictive"`
}

// ScalingEvent records a scaling event
type ScalingEvent struct {
	Timestamp    time.Time          `json:"timestamp"`
	FromReplicas int32              `json:"from_replicas"`
	ToReplicas   int32              `json:"to_replicas"`
	Reason       string             `json:"reason"`
	Metrics      map[string]float64 `json:"metrics"`
	Duration     time.Duration      `json:"duration"`
	Success      bool               `json:"success"`
}

// LoadPredictor predicts future load based on historical data
type LoadPredictor struct {
	name           string
	historicalData []DataPoint
	config         *PredictorConfig
	mu             sync.RWMutex
}

// PredictorConfig contains predictor configuration
type PredictorConfig struct {
	Algorithm         string        `json:"algorithm"`          // "linear", "exponential", "seasonal"
	WindowSize        int           `json:"window_size"`        // Number of data points to consider
	PredictionHorizon time.Duration `json:"prediction_horizon"` // How far ahead to predict
	SeasonalPeriod    time.Duration `json:"seasonal_period"`    // For seasonal algorithms
	Sensitivity       float64       `json:"sensitivity"`        // Prediction sensitivity
}

// DataPoint represents a single data point for prediction
type DataPoint struct {
	Timestamp time.Time              `json:"timestamp"`
	Value     float64                `json:"value"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// LoadPrediction represents a load prediction
type LoadPrediction struct {
	Timestamp      time.Time     `json:"timestamp"`
	PredictedValue float64       `json:"predicted_value"`
	Confidence     float64       `json:"confidence"`
	Algorithm      string        `json:"algorithm"`
	Horizon        time.Duration `json:"horizon"`
}

// NewAutoScaler creates a new auto-scaler
func NewAutoScaler(config *AutoScalerConfig, logger *zap.Logger) *AutoScaler {
	scaler := &AutoScaler{
		logger: logger,
		config: config,
		metrics: &ScalingMetrics{
			CurrentReplicas:    config.MinReplicas,
			DesiredReplicas:    config.MinReplicas,
			CustomMetricValues: make(map[string]float64),
			PredictedLoad:      make(map[string]*LoadPrediction),
			ScalingHistory:     make([]*ScalingEvent, 0),
		},
		predictors:     make(map[string]*LoadPredictor),
		scalingActions: make(chan *ScalingAction, 100),
	}

	// Initialize predictors if predictive scaling is enabled
	if config.PredictiveScaling {
		scaler.initializePredictors()
	}

	return scaler
}

// Start starts the auto-scaler
func (as *AutoScaler) Start(ctx context.Context) error {
	if !as.config.Enabled {
		as.logger.Info("Auto-scaling is disabled")
		return nil
	}

	if !atomic.CompareAndSwapInt32(&as.running, 0, 1) {
		return fmt.Errorf("auto-scaler is already running")
	}

	as.logger.Info("Starting auto-scaler",
		zap.Int32("min_replicas", as.config.MinReplicas),
		zap.Int32("max_replicas", as.config.MaxReplicas),
		zap.Float64("target_cpu", as.config.TargetCPUUtilization),
		zap.Bool("predictive_scaling", as.config.PredictiveScaling))

	// Start scaling evaluator
	go as.scalingEvaluator(ctx)

	// Start scaling executor
	go as.scalingExecutor(ctx)

	// Start metrics collector
	go as.metricsCollector(ctx)

	return nil
}

// Stop stops the auto-scaler
func (as *AutoScaler) Stop() error {
	if !atomic.CompareAndSwapInt32(&as.running, 1, 0) {
		return nil
	}

	as.logger.Info("Stopping auto-scaler")
	close(as.scalingActions)

	return nil
}

// UpdateMetrics updates current metrics for scaling decisions
func (as *AutoScaler) UpdateMetrics(cpuUtilization, memoryUtilization float64, customMetrics map[string]float64) {
	as.metrics.mu.Lock()
	defer as.metrics.mu.Unlock()

	as.metrics.CPUUtilization = cpuUtilization
	as.metrics.MemoryUtilization = memoryUtilization

	for name, value := range customMetrics {
		as.metrics.CustomMetricValues[name] = value
	}

	// Update predictors
	if as.config.PredictiveScaling {
		as.updatePredictors(cpuUtilization, memoryUtilization, customMetrics)
	}
}

// scalingEvaluator evaluates scaling decisions
func (as *AutoScaler) scalingEvaluator(ctx context.Context) {
	ticker := time.NewTicker(as.config.EvaluationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if atomic.LoadInt32(&as.running) == 0 {
				return
			}
			as.evaluateScaling()
		}
	}
}

// evaluateScaling evaluates whether scaling is needed
func (as *AutoScaler) evaluateScaling() {
	as.metrics.mu.RLock()
	currentReplicas := as.metrics.CurrentReplicas
	cpuUtilization := as.metrics.CPUUtilization
	memoryUtilization := as.metrics.MemoryUtilization
	customMetrics := make(map[string]float64)
	for k, v := range as.metrics.CustomMetricValues {
		customMetrics[k] = v
	}
	as.metrics.mu.RUnlock()

	// Calculate desired replicas based on current metrics
	desiredReplicas := as.calculateDesiredReplicas(cpuUtilization, memoryUtilization, customMetrics)

	// Apply predictive scaling if enabled
	if as.config.PredictiveScaling {
		predictiveReplicas := as.calculatePredictiveReplicas()
		if predictiveReplicas > desiredReplicas {
			desiredReplicas = predictiveReplicas
		}
	}

	// Ensure replicas are within bounds
	desiredReplicas = as.clampReplicas(desiredReplicas)

	// Check if scaling is needed
	if desiredReplicas != currentReplicas {
		if as.shouldScale(currentReplicas, desiredReplicas) {
			action := &ScalingAction{
				TargetReplicas: desiredReplicas,
				Timestamp:      time.Now(),
				Metrics: map[string]float64{
					"cpu_utilization":    cpuUtilization,
					"memory_utilization": memoryUtilization,
				},
			}

			if desiredReplicas > currentReplicas {
				action.Type = "scale_up"
				action.Reason = as.getScaleUpReason(cpuUtilization, memoryUtilization, customMetrics)
			} else {
				action.Type = "scale_down"
				action.Reason = as.getScaleDownReason(cpuUtilization, memoryUtilization, customMetrics)
			}

			// Add custom metrics to action
			for name, value := range customMetrics {
				action.Metrics[name] = value
			}

			select {
			case as.scalingActions <- action:
				as.logger.Info("Scaling action queued",
					zap.String("type", action.Type),
					zap.Int32("from", currentReplicas),
					zap.Int32("to", desiredReplicas),
					zap.String("reason", action.Reason))
			default:
				as.logger.Warn("Scaling action queue is full, dropping action")
			}
		}
	}
}

// calculateDesiredReplicas calculates desired replicas based on metrics
func (as *AutoScaler) calculateDesiredReplicas(cpuUtilization, memoryUtilization float64, customMetrics map[string]float64) int32 {
	currentReplicas := as.metrics.CurrentReplicas

	// Calculate CPU-based scaling
	var cpuReplicas int32
	if cpuUtilization > 0 {
		cpuReplicas = int32(math.Ceil(float64(currentReplicas) * cpuUtilization / as.config.TargetCPUUtilization))
	} else {
		cpuReplicas = currentReplicas
	}

	// Calculate memory-based scaling
	var memoryReplicas int32
	if memoryUtilization > 0 {
		memoryReplicas = int32(math.Ceil(float64(currentReplicas) * memoryUtilization / as.config.TargetMemoryUtilization))
	} else {
		memoryReplicas = currentReplicas
	}

	// Take the maximum of CPU and memory requirements
	desiredReplicas := cpuReplicas
	if memoryReplicas > desiredReplicas {
		desiredReplicas = memoryReplicas
	}

	// Consider custom metrics
	for name, metric := range as.config.CustomMetrics {
		if value, exists := customMetrics[name]; exists {
			var customReplicas int32
			if value > 0 && metric.TargetValue > 0 {
				customReplicas = int32(math.Ceil(float64(currentReplicas) * value / metric.TargetValue))

				// Apply weight
				weightedReplicas := int32(float64(customReplicas) * metric.Weight)

				if metric.ScaleDirection == "up" || metric.ScaleDirection == "both" {
					if weightedReplicas > desiredReplicas {
						desiredReplicas = weightedReplicas
					}
				}
			}
		}
	}

	return desiredReplicas
}

// calculatePredictiveReplicas calculates replicas based on predictions
func (as *AutoScaler) calculatePredictiveReplicas() int32 {
	as.metrics.mu.RLock()
	defer as.metrics.mu.RUnlock()

	maxPredictedReplicas := as.metrics.CurrentReplicas

	for _, prediction := range as.metrics.PredictedLoad {
		if prediction.Confidence > 0.7 { // Only consider high-confidence predictions
			predictedReplicas := int32(math.Ceil(float64(as.metrics.CurrentReplicas) * prediction.PredictedValue))
			if predictedReplicas > maxPredictedReplicas {
				maxPredictedReplicas = predictedReplicas
			}
		}
	}

	return maxPredictedReplicas
}

// clampReplicas ensures replicas are within configured bounds
func (as *AutoScaler) clampReplicas(replicas int32) int32 {
	if replicas < as.config.MinReplicas {
		return as.config.MinReplicas
	}
	if replicas > as.config.MaxReplicas {
		return as.config.MaxReplicas
	}
	return replicas
}

// shouldScale determines if scaling should occur based on cooldown periods
func (as *AutoScaler) shouldScale(currentReplicas, desiredReplicas int32) bool {
	as.metrics.mu.RLock()
	lastScaleTime := as.metrics.LastScaleTime
	as.metrics.mu.RUnlock()

	now := time.Now()

	if desiredReplicas > currentReplicas {
		// Scale up
		return now.Sub(lastScaleTime) >= as.config.ScaleUpCooldown
	} else {
		// Scale down
		return now.Sub(lastScaleTime) >= as.config.ScaleDownCooldown
	}
}

// scalingExecutor executes scaling actions
func (as *AutoScaler) scalingExecutor(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case action, ok := <-as.scalingActions:
			if !ok {
				return
			}
			as.executeScalingAction(action)
		}
	}
}

// executeScalingAction executes a scaling action
func (as *AutoScaler) executeScalingAction(action *ScalingAction) {
	startTime := time.Now()

	as.metrics.mu.Lock()
	fromReplicas := as.metrics.CurrentReplicas
	as.metrics.DesiredReplicas = action.TargetReplicas
	as.metrics.LastScaleTime = action.Timestamp
	as.metrics.mu.Unlock()

	// Simulate scaling operation (replace with actual scaling logic)
	success := as.performScaling(action.TargetReplicas)

	duration := time.Since(startTime)

	// Record scaling event
	event := &ScalingEvent{
		Timestamp:    action.Timestamp,
		FromReplicas: fromReplicas,
		ToReplicas:   action.TargetReplicas,
		Reason:       action.Reason,
		Metrics:      action.Metrics,
		Duration:     duration,
		Success:      success,
	}

	as.metrics.mu.Lock()
	if success {
		as.metrics.CurrentReplicas = action.TargetReplicas
		if action.Type == "scale_up" {
			as.metrics.ScaleUpCount++
		} else {
			as.metrics.ScaleDownCount++
		}
	}

	// Add to history (keep last 100 events)
	as.metrics.ScalingHistory = append(as.metrics.ScalingHistory, event)
	if len(as.metrics.ScalingHistory) > 100 {
		as.metrics.ScalingHistory = as.metrics.ScalingHistory[1:]
	}
	as.metrics.mu.Unlock()

	if success {
		as.logger.Info("Scaling action completed",
			zap.String("type", action.Type),
			zap.Int32("from", fromReplicas),
			zap.Int32("to", action.TargetReplicas),
			zap.Duration("duration", duration))
	} else {
		as.logger.Error("Scaling action failed",
			zap.String("type", action.Type),
			zap.Int32("target", action.TargetReplicas),
			zap.Duration("duration", duration))
	}
}

// performScaling performs the actual scaling operation
func (as *AutoScaler) performScaling(targetReplicas int32) bool {
	// This is where you would integrate with Kubernetes HPA, VPA, or custom scaling logic
	// For now, we'll simulate the scaling operation

	as.logger.Info("Performing scaling operation", zap.Int32("target_replicas", targetReplicas))

	// Simulate scaling delay
	time.Sleep(time.Second * 2)

	// Simulate 95% success rate
	return time.Now().UnixNano()%100 < 95
}

// Helper methods for scaling reasons
func (as *AutoScaler) getScaleUpReason(cpu, memory float64, custom map[string]float64) string {
	reasons := []string{}

	if cpu > as.config.TargetCPUUtilization {
		reasons = append(reasons, fmt.Sprintf("CPU utilization %.1f%% > target %.1f%%", cpu*100, as.config.TargetCPUUtilization*100))
	}

	if memory > as.config.TargetMemoryUtilization {
		reasons = append(reasons, fmt.Sprintf("Memory utilization %.1f%% > target %.1f%%", memory*100, as.config.TargetMemoryUtilization*100))
	}

	for name, metric := range as.config.CustomMetrics {
		if value, exists := custom[name]; exists && value > metric.TargetValue {
			reasons = append(reasons, fmt.Sprintf("%s %.2f > target %.2f", name, value, metric.TargetValue))
		}
	}

	if len(reasons) == 0 {
		return "Predictive scaling"
	}

	return fmt.Sprintf("Scale up: %v", reasons)
}

func (as *AutoScaler) getScaleDownReason(cpu, memory float64, custom map[string]float64) string {
	return fmt.Sprintf("Scale down: CPU %.1f%%, Memory %.1f%% below targets", cpu*100, memory*100)
}

// metricsCollector collects and logs scaling metrics
func (as *AutoScaler) metricsCollector(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if atomic.LoadInt32(&as.running) == 0 {
				return
			}
			as.logMetrics()
		}
	}
}

// logMetrics logs current scaling metrics
func (as *AutoScaler) logMetrics() {
	metrics := as.GetMetrics()

	as.logger.Info("Auto-scaling metrics",
		zap.Int32("current_replicas", metrics.CurrentReplicas),
		zap.Int32("desired_replicas", metrics.DesiredReplicas),
		zap.Float64("cpu_utilization", metrics.CPUUtilization*100),
		zap.Float64("memory_utilization", metrics.MemoryUtilization*100),
		zap.Int64("scale_up_count", metrics.ScaleUpCount),
		zap.Int64("scale_down_count", metrics.ScaleDownCount))
}

// GetMetrics returns current scaling metrics
func (as *AutoScaler) GetMetrics() *ScalingMetrics {
	as.metrics.mu.RLock()
	defer as.metrics.mu.RUnlock()

	// Create a copy to avoid race conditions
	metrics := &ScalingMetrics{
		CurrentReplicas:    as.metrics.CurrentReplicas,
		DesiredReplicas:    as.metrics.DesiredReplicas,
		LastScaleTime:      as.metrics.LastScaleTime,
		ScaleUpCount:       as.metrics.ScaleUpCount,
		ScaleDownCount:     as.metrics.ScaleDownCount,
		CPUUtilization:     as.metrics.CPUUtilization,
		MemoryUtilization:  as.metrics.MemoryUtilization,
		CustomMetricValues: make(map[string]float64),
		PredictedLoad:      make(map[string]*LoadPrediction),
	}

	// Copy custom metrics
	for k, v := range as.metrics.CustomMetricValues {
		metrics.CustomMetricValues[k] = v
	}

	// Copy predictions
	for k, v := range as.metrics.PredictedLoad {
		metrics.PredictedLoad[k] = v
	}

	// Copy recent scaling history
	historyLen := len(as.metrics.ScalingHistory)
	if historyLen > 10 {
		metrics.ScalingHistory = make([]*ScalingEvent, 10)
		copy(metrics.ScalingHistory, as.metrics.ScalingHistory[historyLen-10:])
	} else {
		metrics.ScalingHistory = make([]*ScalingEvent, historyLen)
		copy(metrics.ScalingHistory, as.metrics.ScalingHistory)
	}

	return metrics
}

// initializePredictors initializes load predictors
func (as *AutoScaler) initializePredictors() {
	// Initialize CPU predictor
	as.predictors["cpu"] = &LoadPredictor{
		name:           "cpu",
		historicalData: make([]DataPoint, 0),
		config: &PredictorConfig{
			Algorithm:         "linear",
			WindowSize:        100,
			PredictionHorizon: 10 * time.Minute,
			Sensitivity:       0.1,
		},
	}

	// Initialize memory predictor
	as.predictors["memory"] = &LoadPredictor{
		name:           "memory",
		historicalData: make([]DataPoint, 0),
		config: &PredictorConfig{
			Algorithm:         "exponential",
			WindowSize:        50,
			PredictionHorizon: 5 * time.Minute,
			Sensitivity:       0.15,
		},
	}
}

// updatePredictors updates load predictors with new data
func (as *AutoScaler) updatePredictors(cpu, memory float64, custom map[string]float64) {
	now := time.Now()

	// Update CPU predictor
	if cpuPredictor, exists := as.predictors["cpu"]; exists {
		cpuPredictor.addDataPoint(DataPoint{
			Timestamp: now,
			Value:     cpu,
		})

		if prediction := cpuPredictor.predict(); prediction != nil {
			as.metrics.PredictedLoad["cpu"] = prediction
		}
	}

	// Update memory predictor
	if memoryPredictor, exists := as.predictors["memory"]; exists {
		memoryPredictor.addDataPoint(DataPoint{
			Timestamp: now,
			Value:     memory,
		})

		if prediction := memoryPredictor.predict(); prediction != nil {
			as.metrics.PredictedLoad["memory"] = prediction
		}
	}
}

// LoadPredictor methods

// addDataPoint adds a new data point to the predictor
func (lp *LoadPredictor) addDataPoint(point DataPoint) {
	lp.mu.Lock()
	defer lp.mu.Unlock()

	lp.historicalData = append(lp.historicalData, point)

	// Keep only the configured window size
	if len(lp.historicalData) > lp.config.WindowSize {
		lp.historicalData = lp.historicalData[len(lp.historicalData)-lp.config.WindowSize:]
	}
}

// predict generates a load prediction
func (lp *LoadPredictor) predict() *LoadPrediction {
	lp.mu.RLock()
	defer lp.mu.RUnlock()

	if len(lp.historicalData) < 10 {
		return nil // Need at least 10 data points
	}

	switch lp.config.Algorithm {
	case "linear":
		return lp.linearPredict()
	case "exponential":
		return lp.exponentialPredict()
	default:
		return lp.linearPredict()
	}
}

// linearPredict performs linear regression prediction
func (lp *LoadPredictor) linearPredict() *LoadPrediction {
	n := len(lp.historicalData)
	if n < 2 {
		return nil
	}

	// Simple linear regression
	var sumX, sumY, sumXY, sumX2 float64

	for i, point := range lp.historicalData {
		x := float64(i)
		y := point.Value
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	nf := float64(n)
	slope := (nf*sumXY - sumX*sumY) / (nf*sumX2 - sumX*sumX)
	intercept := (sumY - slope*sumX) / nf

	// Predict future value
	futureX := float64(n) + float64(lp.config.PredictionHorizon.Seconds()/60) // Convert to minutes
	predictedValue := slope*futureX + intercept

	// Calculate confidence based on variance
	var variance float64
	for i, point := range lp.historicalData {
		predicted := slope*float64(i) + intercept
		variance += math.Pow(point.Value-predicted, 2)
	}
	variance /= nf

	confidence := math.Max(0, 1.0-variance/math.Max(predictedValue, 0.1))

	return &LoadPrediction{
		Timestamp:      time.Now().Add(lp.config.PredictionHorizon),
		PredictedValue: math.Max(0, predictedValue),
		Confidence:     confidence,
		Algorithm:      "linear",
		Horizon:        lp.config.PredictionHorizon,
	}
}

// exponentialPredict performs exponential smoothing prediction
func (lp *LoadPredictor) exponentialPredict() *LoadPrediction {
	if len(lp.historicalData) < 2 {
		return nil
	}

	alpha := 0.3 // Smoothing factor
	smoothed := lp.historicalData[0].Value

	for i := 1; i < len(lp.historicalData); i++ {
		smoothed = alpha*lp.historicalData[i].Value + (1-alpha)*smoothed
	}

	// Simple trend calculation
	recent := lp.historicalData[len(lp.historicalData)-1].Value
	trend := recent - smoothed

	// Predict future value
	predictedValue := smoothed + trend

	// Calculate confidence based on recent stability
	var recentVariance float64
	recentWindow := math.Min(10, float64(len(lp.historicalData)))
	start := len(lp.historicalData) - int(recentWindow)

	for i := start; i < len(lp.historicalData); i++ {
		recentVariance += math.Pow(lp.historicalData[i].Value-smoothed, 2)
	}
	recentVariance /= recentWindow

	confidence := math.Max(0, 1.0-recentVariance/math.Max(predictedValue, 0.1))

	return &LoadPrediction{
		Timestamp:      time.Now().Add(lp.config.PredictionHorizon),
		PredictedValue: math.Max(0, predictedValue),
		Confidence:     confidence,
		Algorithm:      "exponential",
		Horizon:        lp.config.PredictionHorizon,
	}
}
