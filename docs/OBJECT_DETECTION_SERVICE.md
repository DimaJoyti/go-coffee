# Object Detection Service

A real-time object detection system built in Go that can identify and track different objects in video streams using machine learning models.

## Features

### Core Capabilities
- **Real-time Object Detection**: Process video streams in real-time using YOLO, SSD, or other ML models
- **Object Tracking**: Track objects across frames with unique IDs and motion prediction
- **Multiple Video Sources**: Support for webcam, file, RTMP, and HTTP video streams
- **WebSocket Streaming**: Real-time detection results via WebSocket connections
- **REST API**: Complete API for stream management, detection control, and data retrieval
- **Model Management**: Upload, activate, and manage different detection models
- **Alert System**: Configurable alerts for specific object detections
- **Performance Monitoring**: Comprehensive metrics and observability

### Architecture
- **Clean Architecture**: Domain-driven design with clear separation of concerns
- **Microservice Ready**: Designed to integrate with existing microservice architectures
- **Scalable**: Horizontal scaling support with stateless design
- **Observable**: OpenTelemetry integration for tracing, metrics, and logging

## Quick Start

### Prerequisites
- Go 1.24+
- PostgreSQL 15+
- Redis 7+
- OpenCV 4.8+ (for video processing)
- Docker & Docker Compose (for development)

### Development Setup

1. **Clone and navigate to the project**:
   ```bash
   cd go-coffee
   ```

2. **Start development environment**:
   ```bash
   make -f Makefile.object-detection dev-up
   ```

3. **Build and run the service**:
   ```bash
   make -f Makefile.object-detection build
   make -f Makefile.object-detection run-dev
   ```

4. **Check service health**:
   ```bash
   curl http://localhost:8080/health
   ```

### Using Docker

1. **Build Docker image**:
   ```bash
   make -f Makefile.object-detection docker-build
   ```

2. **Run with Docker Compose**:
   ```bash
   docker-compose -f docker/docker-compose.object-detection.yml up
   ```

## API Documentation

### Health & Monitoring
- `GET /health` - Health check
- `GET /ready` - Readiness check
- `GET /metrics` - Prometheus metrics

### Stream Management
- `POST /api/v1/streams` - Create new video stream
- `GET /api/v1/streams` - List all streams
- `GET /api/v1/streams/{id}` - Get stream details
- `PUT /api/v1/streams/{id}` - Update stream
- `DELETE /api/v1/streams/{id}` - Delete stream
- `POST /api/v1/streams/{id}/start` - Start stream processing
- `POST /api/v1/streams/{id}/stop` - Stop stream processing

### Detection Control
- `POST /api/v1/detection/start` - Start detection on stream
- `POST /api/v1/detection/stop` - Stop detection
- `GET /api/v1/detection/results` - Get detection results
- `GET /api/v1/detection/stats` - Get processing statistics

### Model Management
- `POST /api/v1/models` - Upload new model
- `GET /api/v1/models` - List available models
- `GET /api/v1/models/{id}` - Get model details
- `PUT /api/v1/models/{id}/activate` - Activate model
- `DELETE /api/v1/models/{id}` - Delete model

### Alert Management
- `GET /api/v1/alerts` - Get alerts with filtering
- `PUT /api/v1/alerts/{id}/acknowledge` - Acknowledge alert

### Object Tracking
- `GET /api/v1/tracking/active/{stream_id}` - Get active tracking
- `GET /api/v1/tracking/history/{tracking_id}` - Get tracking history

### WebSocket Endpoints
- `/ws/detections` - Real-time detection results
- `/ws/alerts` - Real-time alert notifications

## Configuration

The service uses YAML configuration with environment variable overrides:

```yaml
# configs/object-detection.yaml
environment: development

server:
  port: 8080
  read_timeout: 30
  write_timeout: 30

database:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  database: object_detection

redis:
  host: localhost
  port: 6379

detection:
  model_path: "./data/models/yolov5s.onnx"
  model_type: "yolo"
  confidence_threshold: 0.5
  nms_threshold: 0.4
  input_size: 640
  enable_gpu: false

tracking:
  enabled: true
  max_age: 30
  min_hits: 3
  iou_threshold: 0.3

websocket:
  enabled: true
  path: "/ws"
  max_connections: 100
```

### Environment Variables
- `PORT` - Server port (default: 8080)
- `ENVIRONMENT` - Environment (development/production)
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` - Database config
- `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD` - Redis config

## Development

### Project Structure
```
internal/object-detection/
├── domain/              # Domain models and interfaces
│   ├── models.go       # Core domain models
│   ├── repository.go   # Repository interfaces
│   └── service.go      # Service interfaces
├── application/        # Business logic (to be implemented)
├── infrastructure/     # External dependencies (to be implemented)
├── transport/          # HTTP handlers and WebSocket
├── config/            # Configuration management
├── container/         # Dependency injection
└── monitoring/        # Metrics and observability
```

### Running Tests
```bash
# Unit tests
make -f Makefile.object-detection test-unit

# Integration tests
make -f Makefile.object-detection test-integration

# Coverage report
make -f Makefile.object-detection test-coverage
```

### Code Quality
```bash
# Format code
make -f Makefile.object-detection fmt

# Run linter
make -f Makefile.object-detection lint

# Run vet
make -f Makefile.object-detection vet
```

## Monitoring & Observability

### Metrics
The service exposes Prometheus metrics at `/metrics`:
- HTTP request metrics (duration, count, status codes)
- Stream processing metrics (FPS, frame count, errors)
- Detection metrics (objects detected, processing time)
- System metrics (memory, CPU, goroutines)

### Logging
Structured logging with configurable levels:
- JSON format for production
- Human-readable for development
- Request correlation IDs
- Performance tracking

### Health Checks
- `/health` - Basic service health
- `/ready` - Dependency readiness (DB, Redis, models)

## Deployment

### Docker
```bash
# Build image
docker build -f docker/Dockerfile.object-detection -t object-detection-service .

# Run container
docker run -p 8080:8080 object-detection-service
```

### Kubernetes
Helm charts and Kubernetes manifests available in `/k8s` directory.

## Performance Considerations

### Optimization Tips
1. **GPU Acceleration**: Enable GPU support for faster inference
2. **Frame Dropping**: Configure frame rate limits for real-time performance
3. **Batch Processing**: Use batch inference for multiple streams
4. **Memory Management**: Monitor memory usage with large video files
5. **Connection Pooling**: Optimize database and Redis connections

### Scaling
- Horizontal scaling with load balancers
- Stream distribution across instances
- Shared model storage
- Redis clustering for high availability

## Troubleshooting

### Common Issues
1. **OpenCV Installation**: Ensure OpenCV is properly installed with video codecs
2. **Model Loading**: Check model file paths and formats
3. **Memory Usage**: Monitor memory consumption with large models
4. **Performance**: Adjust confidence thresholds and input sizes

### Debug Mode
```bash
ENVIRONMENT=development LOG_LEVEL=debug ./object-detection-service
```

## Contributing

1. Follow the existing code structure and patterns
2. Write tests for new functionality
3. Update documentation for API changes
4. Use conventional commit messages
5. Ensure all tests pass before submitting PRs

## License

This project is part of the go-coffee platform and follows the same licensing terms.

## Next Steps

Phase 1 (✅ Complete):
- [x] Core infrastructure setup
- [x] Domain models and interfaces
- [x] Configuration management
- [x] Basic HTTP server and API endpoints
- [x] Dependency injection container
- [x] Unit tests and documentation

Phase 2 (✅ Complete):
- [x] Video processing with GoCV integration
- [x] Video input handlers (webcam, file, RTMP, HTTP)
- [x] Frame processing pipeline with goroutines
- [x] Basic image preprocessing (resize, normalize, crop, color conversion)
- [x] Stream manager for lifecycle management
- [x] Comprehensive video processing tests

Phase 3 (✅ Complete):
- [x] Object detection engine with ONNX Runtime integration
- [x] YOLO model loading and inference (YOLOv5/v8 support)
- [x] Detection result processing with NMS and confidence filtering
- [x] Model management service with upload/activation
- [x] Inference engine with async processing
- [x] Integration with video processing pipeline
- [x] Comprehensive detection engine tests

Phase 4 (✅ Complete):
- [x] Object tracking algorithms with multi-stream support
- [x] Kalman filters for motion prediction and state estimation
- [x] Unique ID assignment and lifecycle management
- [x] Hungarian algorithm for optimal detection-track association
- [x] Trajectory recording and analysis
- [x] Track lifecycle management (creation, confirmation, deletion)
- [x] Comprehensive tracking service integration
- [x] Tracking performance tests and validation

Phase 5 (✅ Complete):
- [x] Real-time WebSocket streaming with hub infrastructure
- [x] Live video feed streaming with detection overlays
- [x] Detection and tracking result broadcasting
- [x] Client connection management and authentication
- [x] Adaptive frame rate control and quality adjustment
- [x] Multi-protocol streaming support (WebSocket, SSE)
- [x] Stream quality adaptation based on network conditions
- [x] Comprehensive streaming tests and validation

Phase 6 (📋 Planned):
- [ ] Advanced features (detection zones, alerts, recording)
- [ ] Analytics and reporting
- [ ] Performance optimization

Phase 7 (📋 Planned):
- [ ] GPU acceleration
- [ ] Horizontal scaling
- [ ] Production deployment
