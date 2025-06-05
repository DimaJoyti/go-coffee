package client

import (
	"context"
	"crypto/tls"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
)

// CoffeeClient представляє gRPC клієнт для взаємодії з сервісом кави
type CoffeeClient struct {
	conn    *grpc.ClientConn
	target  string
	useTLS  bool
	tlsConf *tls.Config
}

// NewCoffeeClient створює новий CoffeeClient
func NewCoffeeClient(target string) *CoffeeClient {
	return &CoffeeClient{
		target:  target,
		useTLS:  false,
		tlsConf: nil,
	}
}

// Connect встановлює з'єднання з gRPC сервером
func (c *CoffeeClient) Connect(ctx context.Context) error {
	var err error

	opts := []grpc.DialOption{
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:            3 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff:           backoff.DefaultConfig,
			MinConnectTimeout: 5 * time.Second,
		}),
		grpc.WithDefaultServiceConfig(`{
			"loadBalancingPolicy": "round_robin",
			"healthCheckConfig": {
				"serviceName": ""
			},
			"retryPolicy": {
				"MaxAttempts": 3,
				"InitialBackoff": "0.1s",
				"MaxBackoff": "1s",
				"BackoffMultiplier": 2.0,
				"RetryableStatusCodes": [ "UNAVAILABLE" ]
			}
		}`),
	}

	if c.useTLS && c.tlsConf != nil {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(c.tlsConf)))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	c.conn, err = grpc.DialContext(
		ctx,
		c.target,
		opts...,
	)
	if err != nil {
		log.Printf("Failed to connect to gRPC server at %s: %v", c.target, err)
		return err
	}

	log.Printf("Connected to gRPC server at %s", c.target)
	return nil
}

// Close закриває з'єднання з gRPC сервером
func (c *CoffeeClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// GetConnection повертає gRPC з'єднання
func (c *CoffeeClient) GetConnection() *grpc.ClientConn {
	return c.conn
}

// CheckHealth verifies the connection health
func (c *CoffeeClient) CheckHealth(ctx context.Context) error {
	healthClient := grpc_health_v1.NewHealthClient(c.conn)
	_, err := healthClient.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	return err
}
