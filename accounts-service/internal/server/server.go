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
	"github.com/yourusername/coffee-order-system/accounts-service/internal/repository"
)

// HTTPServer represents the HTTP server
type HTTPServer struct {
	server *http.Server
	router *mux.Router
}

// NewHTTPServer creates a new HTTP server
func NewHTTPServer(
	cfg *config.Config,
	accountRepo repository.AccountRepository,
	vendorRepo repository.VendorRepository,
	// TODO: Add other repositories
) *HTTPServer {
	router := mux.NewRouter()

	// Create GraphQL resolver
	resolver := resolvers.NewResolver(accountRepo, vendorRepo)

	// Create GraphQL server
	gqlServer := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{
		Resolvers: resolver,
	}))

	// Set up routes
	router.HandleFunc("/health", healthHandler).Methods(http.MethodGet)
	router.Handle("/graphql", gqlServer)
	router.Handle("/playground", playground.Handler("GraphQL Playground", "/graphql"))

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &HTTPServer{
		server: server,
		router: router,
	}
}

// Start starts the HTTP server
func (s *HTTPServer) Start() error {
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the HTTP server
func (s *HTTPServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// healthHandler handles health check requests
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
