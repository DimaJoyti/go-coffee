package redisadvanced

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// TimeSeriesClient provides time series operations using RedisTimeSeries
type TimeSeriesClient struct {
	redis  *redis.Client
	logger *zap.Logger
	config *TimeSeriesConfig
}

// TimeSeriesConfig contains configuration for time series operations
type TimeSeriesConfig struct {
	DefaultRetention time.Duration
	DefaultLabels    map[string]string
	ChunkSize        int64
	DuplicatePolicy  string
}

// TimeSeriesPoint represents a single data point
type TimeSeriesPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// TimeSeriesData represents time series data with metadata
type TimeSeriesData struct {
	Key        string                 `json:"key"`
	Labels     map[string]string      `json:"labels"`
	Points     []TimeSeriesPoint      `json:"points"`
	Aggregation string                `json:"aggregation,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// AggregationQuery represents an aggregation query
type AggregationQuery struct {
	Aggregator string        `json:"aggregator"` // AVG, SUM, MIN, MAX, COUNT, etc.
	TimeBucket time.Duration `json:"time_bucket"`
	StartTime  time.Time     `json:"start_time"`
	EndTime    time.Time     `json:"end_time"`
	Filters    []string      `json:"filters,omitempty"`
}

// MetricDefinition defines a metric with its properties
type MetricDefinition struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Labels      map[string]string `json:"labels"`
	Retention   time.Duration     `json:"retention"`
	ChunkSize   int64             `json:"chunk_size,omitempty"`
}

// NewTimeSeriesClient creates a new time series client
func NewTimeSeriesClient(redisClient *redis.Client, logger *zap.Logger, config *TimeSeriesConfig) *TimeSeriesClient {
	if config == nil {
		config = &TimeSeriesConfig{
			DefaultRetention: 30 * 24 * time.Hour, // 30 days
			DefaultLabels:    map[string]string{"source": "go-coffee"},
			ChunkSize:        4096,
			DuplicatePolicy:  "LAST",
		}
	}

	return &TimeSeriesClient{
		redis:  redisClient,
		logger: logger,
		config: config,
	}
}

// CreateTimeSeries creates a new time series
func (tsc *TimeSeriesClient) CreateTimeSeries(ctx context.Context, key string, labels map[string]string, retention time.Duration) error {
	tsc.logger.Info("Creating time series", zap.String("key", key))

	// Merge with default labels
	allLabels := make(map[string]string)
	for k, v := range tsc.config.DefaultLabels {
		allLabels[k] = v
	}
	for k, v := range labels {
		allLabels[k] = v
	}

	// Build TS.CREATE command
	args := []interface{}{key}

	// Add retention
	if retention > 0 {
		args = append(args, "RETENTION", int64(retention.Milliseconds()))
	} else if tsc.config.DefaultRetention > 0 {
		args = append(args, "RETENTION", int64(tsc.config.DefaultRetention.Milliseconds()))
	}

	// Add chunk size
	if tsc.config.ChunkSize > 0 {
		args = append(args, "CHUNK_SIZE", tsc.config.ChunkSize)
	}

	// Add duplicate policy
	if tsc.config.DuplicatePolicy != "" {
		args = append(args, "DUPLICATE_POLICY", tsc.config.DuplicatePolicy)
	}

	// Add labels
	if len(allLabels) > 0 {
		args = append(args, "LABELS")
		for k, v := range allLabels {
			args = append(args, k, v)
		}
	}

	cmd := redis.NewCmd(ctx, append([]interface{}{"TS.CREATE"}, args...)...)
	if err := tsc.redis.Process(ctx, cmd); err != nil {
		return fmt.Errorf("failed to create time series: %w", err)
	}

	tsc.logger.Info("Time series created successfully", zap.String("key", key))
	return nil
}

// AddPoint adds a single data point to a time series
func (tsc *TimeSeriesClient) AddPoint(ctx context.Context, key string, timestamp time.Time, value float64) error {
	tsc.logger.Debug("Adding point to time series", 
		zap.String("key", key), 
		zap.Time("timestamp", timestamp),
		zap.Float64("value", value),
	)

	timestampMs := timestamp.UnixMilli()
	if timestamp.IsZero() {
		timestampMs = time.Now().UnixMilli()
	}

	cmd := redis.NewCmd(ctx, "TS.ADD", key, timestampMs, value)
	if err := tsc.redis.Process(ctx, cmd); err != nil {
		return fmt.Errorf("failed to add point: %w", err)
	}

	return nil
}

// AddPoints adds multiple data points to a time series
func (tsc *TimeSeriesClient) AddPoints(ctx context.Context, key string, points []TimeSeriesPoint) error {
	tsc.logger.Info("Adding multiple points to time series", 
		zap.String("key", key), 
		zap.Int("points_count", len(points)),
	)

	// Use pipeline for better performance
	pipe := tsc.redis.Pipeline()

	for _, point := range points {
		timestampMs := point.Timestamp.UnixMilli()
		if point.Timestamp.IsZero() {
			timestampMs = time.Now().UnixMilli()
		}
		pipe.Process(ctx, redis.NewCmd(ctx, "TS.ADD", key, timestampMs, point.Value))
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to add points: %w", err)
	}

	tsc.logger.Info("Points added successfully", zap.String("key", key))
	return nil
}

// GetRange retrieves data points within a time range
func (tsc *TimeSeriesClient) GetRange(ctx context.Context, key string, startTime, endTime time.Time) ([]TimeSeriesPoint, error) {
	tsc.logger.Info("Getting time series range", 
		zap.String("key", key),
		zap.Time("start_time", startTime),
		zap.Time("end_time", endTime),
	)

	startMs := startTime.UnixMilli()
	endMs := endTime.UnixMilli()

	cmd := redis.NewCmd(ctx, "TS.RANGE", key, startMs, endMs)
	if err := tsc.redis.Process(ctx, cmd); err != nil {
		return nil, fmt.Errorf("failed to get range: %w", err)
	}

	points, err := tsc.parseTimeSeriesPoints(cmd.Val())
	if err != nil {
		return nil, fmt.Errorf("failed to parse points: %w", err)
	}

	tsc.logger.Info("Range retrieved successfully", 
		zap.String("key", key), 
		zap.Int("points_count", len(points)),
	)

	return points, nil
}

// GetAggregation retrieves aggregated data
func (tsc *TimeSeriesClient) GetAggregation(ctx context.Context, key string, query *AggregationQuery) ([]TimeSeriesPoint, error) {
	tsc.logger.Info("Getting time series aggregation", 
		zap.String("key", key),
		zap.String("aggregator", query.Aggregator),
		zap.Duration("time_bucket", query.TimeBucket),
	)

	startMs := query.StartTime.UnixMilli()
	endMs := query.EndTime.UnixMilli()
	bucketMs := int64(query.TimeBucket.Milliseconds())

	cmd := redis.NewCmd(ctx, "TS.RANGE", key, startMs, endMs, 
		"AGGREGATION", query.Aggregator, bucketMs)
	
	if err := tsc.redis.Process(ctx, cmd); err != nil {
		return nil, fmt.Errorf("failed to get aggregation: %w", err)
	}

	points, err := tsc.parseTimeSeriesPoints(cmd.Val())
	if err != nil {
		return nil, fmt.Errorf("failed to parse aggregated points: %w", err)
	}

	tsc.logger.Info("Aggregation retrieved successfully", 
		zap.String("key", key), 
		zap.Int("points_count", len(points)),
	)

	return points, nil
}

// MultiGet retrieves data from multiple time series
func (tsc *TimeSeriesClient) MultiGet(ctx context.Context, filters []string, startTime, endTime time.Time) (map[string][]TimeSeriesPoint, error) {
	tsc.logger.Info("Getting multiple time series", zap.Strings("filters", filters))

	startMs := startTime.UnixMilli()
	endMs := endTime.UnixMilli()

	// Build TS.MRANGE command
	args := []interface{}{startMs, endMs, "FILTER"}
	for _, filter := range filters {
		args = append(args, filter)
	}

	cmd := redis.NewCmd(ctx, append([]interface{}{"TS.MRANGE"}, args...)...)
	if err := tsc.redis.Process(ctx, cmd); err != nil {
		return nil, fmt.Errorf("failed to get multiple series: %w", err)
	}

	results, err := tsc.parseMultiTimeSeriesResult(cmd.Val())
	if err != nil {
		return nil, fmt.Errorf("failed to parse multi-series result: %w", err)
	}

	tsc.logger.Info("Multiple series retrieved successfully", zap.Int("series_count", len(results)))
	return results, nil
}

// CreateRule creates a compaction rule
func (tsc *TimeSeriesClient) CreateRule(ctx context.Context, sourceKey, destKey string, aggregator string, bucketDuration time.Duration) error {
	tsc.logger.Info("Creating compaction rule", 
		zap.String("source", sourceKey),
		zap.String("dest", destKey),
		zap.String("aggregator", aggregator),
	)

	bucketMs := int64(bucketDuration.Milliseconds())

	cmd := redis.NewCmd(ctx, "TS.CREATERULE", sourceKey, destKey, 
		"AGGREGATION", aggregator, bucketMs)
	
	if err := tsc.redis.Process(ctx, cmd); err != nil {
		return fmt.Errorf("failed to create rule: %w", err)
	}

	tsc.logger.Info("Compaction rule created successfully")
	return nil
}

// GetInfo retrieves information about a time series
func (tsc *TimeSeriesClient) GetInfo(ctx context.Context, key string) (map[string]interface{}, error) {
	cmd := redis.NewCmd(ctx, "TS.INFO", key)
	if err := tsc.redis.Process(ctx, cmd); err != nil {
		return nil, fmt.Errorf("failed to get time series info: %w", err)
	}

	info, err := tsc.parseTimeSeriesInfo(cmd.Val())
	if err != nil {
		return nil, fmt.Errorf("failed to parse time series info: %w", err)
	}

	return info, nil
}

// RecordMetric records a metric with automatic time series creation
func (tsc *TimeSeriesClient) RecordMetric(ctx context.Context, metricName string, value float64, labels map[string]string) error {
	key := fmt.Sprintf("metric:%s", metricName)
	
	// Add labels to key for uniqueness
	if len(labels) > 0 {
		for k, v := range labels {
			key += fmt.Sprintf(":%s=%s", k, v)
		}
	}

	// Try to add point (will create series if it doesn't exist)
	timestampMs := time.Now().UnixMilli()
	
	// Use TS.ADD with ON_DUPLICATE policy
	cmd := redis.NewCmd(ctx, "TS.ADD", key, timestampMs, value, 
		"ON_DUPLICATE", tsc.config.DuplicatePolicy)
	
	if err := tsc.redis.Process(ctx, cmd); err != nil {
		// If series doesn't exist, create it first
		if err := tsc.CreateTimeSeries(ctx, key, labels, tsc.config.DefaultRetention); err != nil {
			return fmt.Errorf("failed to create time series: %w", err)
		}
		
		// Retry adding the point
		if err := tsc.redis.Process(ctx, cmd); err != nil {
			return fmt.Errorf("failed to add metric point: %w", err)
		}
	}

	return nil
}

// Helper methods

func (tsc *TimeSeriesClient) parseTimeSeriesPoints(rawData interface{}) ([]TimeSeriesPoint, error) {
	points := []TimeSeriesPoint{}
	
	if dataSlice, ok := rawData.([]interface{}); ok {
		for _, item := range dataSlice {
			if pointSlice, ok := item.([]interface{}); ok && len(pointSlice) == 2 {
				timestampMs, _ := strconv.ParseInt(fmt.Sprintf("%v", pointSlice[0]), 10, 64)
				value, _ := strconv.ParseFloat(fmt.Sprintf("%v", pointSlice[1]), 64)
				
				point := TimeSeriesPoint{
					Timestamp: time.UnixMilli(timestampMs),
					Value:     value,
				}
				points = append(points, point)
			}
		}
	}
	
	return points, nil
}

func (tsc *TimeSeriesClient) parseMultiTimeSeriesResult(rawData interface{}) (map[string][]TimeSeriesPoint, error) {
	results := make(map[string][]TimeSeriesPoint)
	
	if dataSlice, ok := rawData.([]interface{}); ok {
		for _, item := range dataSlice {
			if seriesSlice, ok := item.([]interface{}); ok && len(seriesSlice) >= 2 {
				key := fmt.Sprintf("%v", seriesSlice[0])
				
				if len(seriesSlice) > 2 {
					points, err := tsc.parseTimeSeriesPoints(seriesSlice[2])
					if err == nil {
						results[key] = points
					}
				}
			}
		}
	}
	
	return results, nil
}

func (tsc *TimeSeriesClient) parseTimeSeriesInfo(rawData interface{}) (map[string]interface{}, error) {
	info := make(map[string]interface{})
	
	if dataSlice, ok := rawData.([]interface{}); ok {
		for i := 0; i < len(dataSlice); i += 2 {
			if i+1 < len(dataSlice) {
				key := fmt.Sprintf("%v", dataSlice[i])
				value := dataSlice[i+1]
				info[key] = value
			}
		}
	}
	
	return info, nil
}
