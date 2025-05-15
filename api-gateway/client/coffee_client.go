package client

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// CoffeeClient представляє gRPC клієнт для взаємодії з сервісом кави
type CoffeeClient struct {
	conn   *grpc.ClientConn
	target string
}

// NewCoffeeClient створює новий CoffeeClient
func NewCoffeeClient(target string) *CoffeeClient {
	return &CoffeeClient{
		target: target,
	}
}

// Connect встановлює з'єднання з gRPC сервером
func (c *CoffeeClient) Connect() error {
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c.conn, err = grpc.DialContext(
		ctx,
		c.target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
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
