package kafka

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStreamProcessor is a mock implementation of a stream processor
type MockStreamProcessor struct {
	mock.Mock
	processedRecords []StreamRecord
}

type StreamRecord struct {
	Key       string
	Value     []byte
	Topic     string
	Partition int32
	Offset    int64
	Timestamp time.Time
}

func NewMockStreamProcessor() *MockStreamProcessor {
	return &MockStreamProcessor{
		processedRecords: make([]StreamRecord, 0),
	}
}

func (m *MockStreamProcessor) ProcessRecord(ctx context.Context, record StreamRecord) error {
	m.processedRecords = append(m.processedRecords, record)
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *MockStreamProcessor) GetProcessedRecords() []StreamRecord {
	return m.processedRecords
}

// MockKafkaStreams is a mock implementation of Kafka Streams
type MockKafkaStreams struct {
	mock.Mock
	running bool
	records chan StreamRecord
}

func NewMockKafkaStreams() *MockKafkaStreams {
	return &MockKafkaStreams{
		records: make(chan StreamRecord, 10),
	}
}

func (m *MockKafkaStreams) Start(ctx context.Context) error {
	m.running = true
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockKafkaStreams) Stop() error {
	m.running = false
	close(m.records)
	args := m.Called()
	return args.Error(0)
}

func (m *MockKafkaStreams) IsRunning() bool {
	return m.running
}

func (m *MockKafkaStreams) AddRecord(record StreamRecord) {
	if m.running {
		select {
		case m.records <- record:
		default:
			// Channel is full, skip
		}
	}
}

func (m *MockKafkaStreams) GetRecords() <-chan StreamRecord {
	return m.records
}

// StreamsApplication represents a Kafka Streams application
type StreamsApplication struct {
	processor MockStreamProcessor
	streams   *MockKafkaStreams
	running   bool
	stopChan  chan struct{}
}

func NewStreamsApplication(processor MockStreamProcessor) *StreamsApplication {
	return &StreamsApplication{
		processor: processor,
		streams:   NewMockKafkaStreams(),
		stopChan:  make(chan struct{}),
	}
}

func (s *StreamsApplication) Start(ctx context.Context) error {
	s.running = true
	
	// Start the streams
	if err := s.streams.Start(ctx); err != nil {
		return err
	}

	// Process records in a goroutine
	go s.processRecords(ctx)
	
	return nil
}

func (s *StreamsApplication) Stop() error {
	s.running = false
	close(s.stopChan)
	return s.streams.Stop()
}

func (s *StreamsApplication) processRecords(ctx context.Context) {
	for {
		select {
		case <-s.stopChan:
			return
		case <-ctx.Done():
			return
		case record := <-s.streams.GetRecords():
			if err := s.processor.ProcessRecord(ctx, record); err != nil {
				// Log error in real implementation
				continue
			}
		}
	}
}

func (s *StreamsApplication) IsHealthy() bool {
	return s.running && s.streams.IsRunning()
}

func TestStreamsApplication_Start(t *testing.T) {
	// Arrange
	mockProcessor := NewMockStreamProcessor()
	mockProcessor.On("ProcessRecord", mock.Anything, mock.Anything).Return(nil)
	
	app := NewStreamsApplication(*mockProcessor)
	app.streams.On("Start", mock.Anything).Return(nil)
	app.streams.On("Stop").Return(nil)

	// Act
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := app.Start(ctx)
	
	// Add a test record
	testRecord := StreamRecord{
		Key:       "test-key",
		Value:     []byte(`{"order_id": "123", "status": "processing"}`),
		Topic:     "orders",
		Partition: 0,
		Offset:    1,
		Timestamp: time.Now(),
	}
	app.streams.AddRecord(testRecord)

	// Wait a bit for processing
	time.Sleep(50 * time.Millisecond)
	
	// Stop the application
	stopErr := app.Stop()

	// Assert
	assert.NoError(t, err)
	assert.NoError(t, stopErr)
	assert.True(t, len(mockProcessor.GetProcessedRecords()) > 0)
	app.streams.AssertExpectations(t)
}

func TestStreamsApplication_Stop(t *testing.T) {
	// Arrange
	mockProcessor := NewMockStreamProcessor()
	app := NewStreamsApplication(*mockProcessor)
	app.streams.On("Stop").Return(nil)

	// Act
	err := app.Stop()

	// Assert
	assert.NoError(t, err)
	assert.False(t, app.running)
	app.streams.AssertExpectations(t)
}

func TestStreamsApplication_HealthCheck(t *testing.T) {
	// Arrange
	mockProcessor := NewMockStreamProcessor()
	app := NewStreamsApplication(*mockProcessor)
	
	// Test when not running
	assert.False(t, app.IsHealthy())
	
	// Test when running
	app.running = true
	app.streams.running = true
	assert.True(t, app.IsHealthy())
	
	// Test when streams not running
	app.streams.running = false
	assert.False(t, app.IsHealthy())
}

func TestStreamRecord_Processing(t *testing.T) {
	// Arrange
	mockProcessor := NewMockStreamProcessor()
	testRecord := StreamRecord{
		Key:       "order-123",
		Value:     []byte(`{"id": "123", "customer": "John", "total": 15.50}`),
		Topic:     "coffee-orders",
		Partition: 0,
		Offset:    42,
		Timestamp: time.Now(),
	}
	
	mockProcessor.On("ProcessRecord", mock.Anything, testRecord).Return(nil)

	// Act
	ctx := context.Background()
	err := mockProcessor.ProcessRecord(ctx, testRecord)

	// Assert
	assert.NoError(t, err)
	processedRecords := mockProcessor.GetProcessedRecords()
	assert.Len(t, processedRecords, 1)
	assert.Equal(t, testRecord, processedRecords[0])
	mockProcessor.AssertExpectations(t)
}

func TestStreamRecord_ErrorHandling(t *testing.T) {
	// Arrange
	mockProcessor := NewMockStreamProcessor()
	testRecord := StreamRecord{
		Key:   "invalid-record",
		Value: []byte(`invalid json`),
		Topic: "orders",
	}
	
	expectedError := assert.AnError
	mockProcessor.On("ProcessRecord", mock.Anything, testRecord).Return(expectedError)

	// Act
	ctx := context.Background()
	err := mockProcessor.ProcessRecord(ctx, testRecord)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockProcessor.AssertExpectations(t)
}

func TestStreamsApplication_Integration(t *testing.T) {
	// This is a more comprehensive integration test
	mockProcessor := NewMockStreamProcessor()
	mockProcessor.On("ProcessRecord", mock.Anything, mock.Anything).Return(nil)
	
	app := NewStreamsApplication(*mockProcessor)
	app.streams.On("Start", mock.Anything).Return(nil)
	app.streams.On("Stop").Return(nil)

	// Test full lifecycle
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// Start application
	err := app.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, app.IsHealthy())

	// Add multiple records
	records := []StreamRecord{
		{Key: "order-1", Value: []byte(`{"id": "1", "status": "new"}`), Topic: "orders"},
		{Key: "order-2", Value: []byte(`{"id": "2", "status": "processing"}`), Topic: "orders"},
		{Key: "order-3", Value: []byte(`{"id": "3", "status": "completed"}`), Topic: "orders"},
	}

	for _, record := range records {
		app.streams.AddRecord(record)
	}

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Stop application
	stopErr := app.Stop()
	assert.NoError(t, stopErr)
	assert.False(t, app.IsHealthy())

	// Verify records were processed
	processedRecords := mockProcessor.GetProcessedRecords()
	assert.True(t, len(processedRecords) >= len(records))
	
	app.streams.AssertExpectations(t)
}
