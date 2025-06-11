package worker

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMessage represents a mock Kafka message
type MockMessage struct {
	topic     string
	partition int32
	offset    int64
	key       []byte
	value     []byte
}

func (m *MockMessage) Topic() string {
	if m == nil {
		return ""
	}
	return m.topic
}
func (m *MockMessage) Partition() int32 {
	if m == nil {
		return 0
	}
	return m.partition
}
func (m *MockMessage) Offset() int64 {
	if m == nil {
		return 0
	}
	return m.offset
}
func (m *MockMessage) Key() []byte {
	if m == nil {
		return nil
	}
	return m.key
}
func (m *MockMessage) Value() []byte {
	if m == nil {
		return nil
	}
	return m.value
}

// MockConsumer is a mock implementation of the Kafka consumer
type MockConsumer struct {
	mock.Mock
	messages chan *MockMessage
	closed   bool
}

func NewMockConsumer() *MockConsumer {
	return &MockConsumer{
		messages: make(chan *MockMessage, 10),
	}
}

func (m *MockConsumer) Subscribe(topics []string) error {
	args := m.Called(topics)
	return args.Error(0)
}

func (m *MockConsumer) Poll(timeout time.Duration) (interface{}, error) {
	select {
	case msg := <-m.messages:
		return msg, nil
	case <-time.After(timeout):
		return nil, nil
	}
}

func (m *MockConsumer) Close() error {
	if !m.closed {
		m.closed = true
		close(m.messages)
	}
	args := m.Called()
	return args.Error(0)
}

func (m *MockConsumer) AddMessage(topic string, partition int32, offset int64, key, value []byte) {
	if !m.closed {
		m.messages <- &MockMessage{
			topic:     topic,
			partition: partition,
			offset:    offset,
			key:       key,
			value:     value,
		}
	}
}

// MockProcessor is a mock message processor
type MockProcessor struct {
	mock.Mock
	processedMessages [][]byte
}

func NewMockProcessor() *MockProcessor {
	return &MockProcessor{
		processedMessages: make([][]byte, 0),
	}
}

func (m *MockProcessor) ProcessMessage(ctx context.Context, message []byte) error {
	m.processedMessages = append(m.processedMessages, message)
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockProcessor) GetProcessedMessages() [][]byte {
	return m.processedMessages
}

func TestWorker_Start(t *testing.T) {
	// Arrange
	mockConsumer := NewMockConsumer()
	mockProcessor := NewMockProcessor()

	mockConsumer.On("Subscribe", []string{"test-topic"}).Return(nil)
	mockConsumer.On("Close").Return(nil)
	mockProcessor.On("ProcessMessage", mock.Anything, mock.Anything).Return(nil)

	worker := &Worker{
		consumer:  mockConsumer,
		processor: mockProcessor,
		topics:    []string{"test-topic"},
		stopChan:  make(chan struct{}),
	}

	// Add a test message
	testMessage := []byte(`{"id": "test-order", "customer": "John Doe", "coffee": "Espresso"}`)
	mockConsumer.AddMessage("test-topic", 0, 1, []byte("key"), testMessage)

	// Act
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := worker.Start(ctx)
	assert.NoError(t, err)

	// Give the worker time to process the message
	time.Sleep(30 * time.Millisecond)

	// Stop the worker
	worker.Stop()

	// Give time for cleanup
	time.Sleep(10 * time.Millisecond)

	// Assert
	mockConsumer.AssertExpectations(t)
	mockProcessor.AssertExpectations(t)
}

func TestWorker_Stop(t *testing.T) {
	// Arrange
	mockConsumer := NewMockConsumer()
	mockProcessor := NewMockProcessor()

	mockConsumer.On("Close").Return(nil)

	worker := &Worker{
		consumer:  mockConsumer,
		processor: mockProcessor,
		stopChan:  make(chan struct{}),
	}

	// Act
	worker.Stop()

	// Assert
	select {
	case <-worker.stopChan:
		// Channel should be closed
	default:
		t.Error("Stop channel should be closed")
	}

	mockConsumer.AssertExpectations(t)
}

func TestWorker_ProcessMessage_Success(t *testing.T) {
	// Arrange
	mockConsumer := NewMockConsumer()
	mockProcessor := NewMockProcessor()

	testMessage := []byte(`{"id": "test-order", "customer": "Jane Doe", "coffee": "Latte"}`)
	mockProcessor.On("ProcessMessage", mock.Anything, testMessage).Return(nil)

	worker := &Worker{
		consumer:  mockConsumer,
		processor: mockProcessor,
	}

	// Act
	ctx := context.Background()
	err := worker.processMessage(ctx, testMessage)

	// Assert
	assert.NoError(t, err)
	assert.Contains(t, mockProcessor.GetProcessedMessages(), testMessage)
	mockProcessor.AssertExpectations(t)
}

func TestWorker_ProcessMessage_Error(t *testing.T) {
	// Arrange
	mockConsumer := NewMockConsumer()
	mockProcessor := NewMockProcessor()

	testMessage := []byte(`invalid json`)
	expectedError := assert.AnError
	mockProcessor.On("ProcessMessage", mock.Anything, testMessage).Return(expectedError)

	worker := &Worker{
		consumer:  mockConsumer,
		processor: mockProcessor,
	}

	// Act
	ctx := context.Background()
	err := worker.processMessage(ctx, testMessage)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockProcessor.AssertExpectations(t)
}

func TestWorker_HealthCheck(t *testing.T) {
	// Arrange
	mockConsumer := NewMockConsumer()
	mockProcessor := NewMockProcessor()

	worker := &Worker{
		consumer:  mockConsumer,
		processor: mockProcessor,
		running:   true,
	}

	// Act
	isHealthy := worker.IsHealthy()

	// Assert
	assert.True(t, isHealthy)
}

func TestWorker_HealthCheck_NotRunning(t *testing.T) {
	// Arrange
	mockConsumer := NewMockConsumer()
	mockProcessor := NewMockProcessor()

	worker := &Worker{
		consumer:  mockConsumer,
		processor: mockProcessor,
		running:   false,
	}

	// Act
	isHealthy := worker.IsHealthy()

	// Assert
	assert.False(t, isHealthy)
}

// Helper function to create a test worker
func createTestWorker() *Worker {
	return &Worker{
		consumer:  NewMockConsumer(),
		processor: NewMockProcessor(),
		topics:    []string{"test-topic"},
		stopChan:  make(chan struct{}),
		running:   false,
	}
}

func TestWorker_Integration(t *testing.T) {
	// This is a more comprehensive integration test
	worker := createTestWorker()

	// Test that worker can be created and configured
	assert.NotNil(t, worker)
	assert.NotNil(t, worker.consumer)
	assert.NotNil(t, worker.processor)
	assert.Equal(t, []string{"test-topic"}, worker.topics)
	assert.False(t, worker.running)
}
