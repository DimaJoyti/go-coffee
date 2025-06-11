module github.com/DimaJoyti/go-coffee

go 1.24.0

require (
	github.com/briandowns/spinner v1.23.1
	github.com/fatih/color v1.18.0
	github.com/gin-gonic/gin v1.10.1
	github.com/go-redis/redis/v8 v8.11.5
	github.com/golang-jwt/jwt/v5 v5.2.2
	github.com/google/uuid v1.6.0
	github.com/gorilla/mux v1.8.1
	github.com/gorilla/websocket v1.5.4-0.20250319132907-e064f32e3674

	// Infrastructure as Code
	github.com/hashicorp/terraform-exec v0.21.0
	github.com/hashicorp/terraform-json v0.22.1
	github.com/jmoiron/sqlx v1.4.0
	github.com/lib/pq v1.10.9
	github.com/manifoldco/promptui v0.9.0
	github.com/olekukonko/tablewriter v0.0.5
	github.com/prometheus/client_golang v1.22.0
	github.com/redis/go-redis/v9 v9.10.0
	github.com/rs/cors v1.7.0
	github.com/shopspring/decimal v1.4.0

	// CLI and Cloud-Native Dependencies
	github.com/spf13/cobra v1.8.1
	github.com/spf13/viper v1.19.0
	github.com/stretchr/testify v1.10.0

	// OpenTelemetry
	go.opentelemetry.io/otel v1.32.0
	go.opentelemetry.io/otel/exporters/jaeger v1.17.0
	go.opentelemetry.io/otel/exporters/prometheus v0.54.0
	go.opentelemetry.io/otel/metric v1.32.0
	go.opentelemetry.io/otel/sdk v1.32.0
	go.opentelemetry.io/otel/sdk/metric v1.32.0
	go.opentelemetry.io/otel/trace v1.32.0
	go.uber.org/zap v1.27.0
	golang.org/x/crypto v0.38.0

	// Additional dependencies
	google.golang.org/grpc v1.67.3
	google.golang.org/protobuf v1.36.5
	gopkg.in/yaml.v3 v3.0.1

	// Kubernetes and Cloud Dependencies
	k8s.io/api v0.31.3
	k8s.io/apimachinery v0.31.3
	k8s.io/client-go v0.31.3
	sigs.k8s.io/controller-runtime v0.19.3
	sigs.k8s.io/yaml v1.4.0
)
