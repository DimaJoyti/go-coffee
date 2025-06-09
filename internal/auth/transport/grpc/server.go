package grpc

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/DimaJoyti/go-coffee/internal/auth/application"
	"github.com/DimaJoyti/go-coffee/internal/auth/config"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// Server represents the gRPC server for auth service
type Server struct {
	config      *config.Config
	authService application.AuthService
	logger      *logger.Logger
	grpcServer  *grpc.Server
	listener    net.Listener
}

// NewServer creates a new gRPC server instance
func NewServer(
	cfg *config.Config,
	authService application.AuthService,
	logger *logger.Logger,
) *Server {
	return &Server{
		config:      cfg,
		authService: authService,
		logger:      logger,
	}
}

// Start starts the gRPC server
func (s *Server) Start() error {
	// Create listener
	addr := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.GRPCPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	s.listener = listener

	// Create gRPC server with interceptors
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(s.unaryInterceptor),
		grpc.StreamInterceptor(s.streamInterceptor),
	}

	// Add TLS if enabled
	if s.config.TLS.Enabled {
		// TODO: Add TLS credentials
		s.logger.Info("TLS is enabled but not implemented yet")
	}

	s.grpcServer = grpc.NewServer(opts...)

	// Register auth service
	// TODO: Register the generated auth service when proto files are generated
	// auth.RegisterAuthServiceServer(s.grpcServer, NewAuthHandler(s.authService, s.logger))

	// Enable reflection for development
	if s.config.Environment == "development" {
		reflection.Register(s.grpcServer)
		s.logger.Info("gRPC reflection enabled for development")
	}

	// Start serving
	s.logger.WithField("address", addr).Info("🌐 gRPC server starting")
	
	go func() {
		if err := s.grpcServer.Serve(listener); err != nil {
			s.logger.WithError(err).Error("gRPC server failed")
		}
	}()

	return nil
}

// Stop gracefully stops the gRPC server
func (s *Server) Stop() error {
	if s.grpcServer != nil {
		s.logger.Info("🛑 Stopping gRPC server...")
		s.grpcServer.GracefulStop()
		s.logger.Info("✅ gRPC server stopped")
	}
	return nil
}

// GetListener returns the server listener
func (s *Server) GetListener() net.Listener {
	return s.listener
}

// unaryInterceptor provides logging and error handling for unary RPCs
func (s *Server) unaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	// Log request
	s.logger.WithFields(map[string]interface{}{
		"method": info.FullMethod,
		"type":   "unary",
	}).Debug("gRPC request received")

	// Call handler
	resp, err := handler(ctx, req)

	// Log response
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"method": info.FullMethod,
			"error":  err.Error(),
		}).Error("gRPC request failed")
	} else {
		s.logger.WithFields(map[string]interface{}{
			"method": info.FullMethod,
		}).Debug("gRPC request completed")
	}

	return resp, err
}

// streamInterceptor provides logging and error handling for streaming RPCs
func (s *Server) streamInterceptor(
	srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	// Log stream start
	s.logger.WithFields(map[string]interface{}{
		"method": info.FullMethod,
		"type":   "stream",
	}).Debug("gRPC stream started")

	// Call handler
	err := handler(srv, ss)

	// Log stream end
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"method": info.FullMethod,
			"error":  err.Error(),
		}).Error("gRPC stream failed")
	} else {
		s.logger.WithFields(map[string]interface{}{
			"method": info.FullMethod,
		}).Debug("gRPC stream completed")
	}

	return err
}
