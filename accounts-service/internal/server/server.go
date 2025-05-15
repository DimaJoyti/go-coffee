package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/mux"
	"github.com/yourusername/coffee-order-system/accounts-service/internal/config"
	"github.com/yourusername/coffee-order-system/accounts-service/internal/graphql/generated"
	"github.com/yourusername/coffee-order-system/accounts-service/internal/graphql/resolvers"
	"github.com/yourusername/coffee-order-system/accounts-service/internal/logging"
	"github.com/yourusername/coffee-order-system/accounts-service/internal/metrics"
)

// HTTPServer represents the HTTP server
type HTTPServer struct {
	server  *http.Server
	router  *mux.Router
	logger  *logging.Logger
	metrics *metrics.Metrics
}

// NewHTTPServer creates a new HTTP server
func NewHTTPServer(
	cfg *config.Config,
	resolver *Resolver,
) *HTTPServer {
	// Create logger
	logger, err := logging.NewLogger(logging.Config{
		Level:       cfg.Logging.Level,
		Development: cfg.Logging.Development,
		Encoding:    cfg.Logging.Encoding,
	})
	if err != nil {
		logger = logging.NewDefaultLogger()
		logger.Sugar().Errorf("Failed to create logger: %v", err)
	}

	// Create metrics
	metricsInstance := metrics.NewMetrics()

	// Create router
	router := mux.NewRouter()

	// Create GraphQL resolver
	gqlResolver := resolvers.NewResolver(
		resolver.AccountService,
		resolver.VendorService,
		resolver.ProductService,
		resolver.OrderService,
	)

	// Create GraphQL server
	gqlServer := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{
		Resolvers: gqlResolver,
	}))

	// Add metrics middleware
	router.Use(metricsInstance.HTTPMiddleware)

	// Set up routes
	router.HandleFunc("/health", healthHandler).Methods(http.MethodGet)
	router.Handle("/graphql", gqlServer)
	router.Handle("/playground", playground.Handler("GraphQL Playground", "/graphql"))

	// Add metrics endpoint
	if cfg.Metrics.Enabled {
		router.Handle("/metrics", metricsInstance.Handler())
	}

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &HTTPServer{
		server:  server,
		router:  router,
		logger:  logger,
		metrics: metricsInstance,
	}
}

// Start starts the HTTP server
func (s *HTTPServer) Start() error {
	s.logger.Sugar().Infof("Starting HTTP server on %s", s.server.Addr)
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the HTTP server
func (s *HTTPServer) Shutdown(ctx context.Context) error {
	s.logger.Sugar().Info("Shutting down HTTP server")
	return s.server.Shutdown(ctx)
}

// healthHandler handles health check requests
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
