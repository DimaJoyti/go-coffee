package grpc

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/timestamppb"

	"kafka_producer/config"
	"kafka_producer/kafka"
	"kafka_producer/store"
)

// CoffeeServiceServer представляє gRPC сервер для сервісу кави
type CoffeeServiceServer struct {
	UnimplementedCoffeeServiceServer
	kafkaProducer kafka.Producer
	config        *config.Config
	orderStore    store.OrderStore
}

// NewCoffeeServiceServer створює новий CoffeeServiceServer
func NewCoffeeServiceServer(kafkaProducer kafka.Producer, config *config.Config, orderStore store.OrderStore) *CoffeeServiceServer {
	return &CoffeeServiceServer{
		kafkaProducer: kafkaProducer,
		config:        config,
		orderStore:    orderStore,
	}
}

// StartGRPCServer запускає gRPC сервер
func StartGRPCServer(server *CoffeeServiceServer, port string) error {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}

	s := grpc.NewServer()
	RegisterCoffeeServiceServer(s, server)
	reflection.Register(s)

	log.Printf("Starting gRPC server on %s", port)
	return s.Serve(lis)
}

// PlaceOrder обробляє запит на створення замовлення
func (s *CoffeeServiceServer) PlaceOrder(ctx context.Context, req *PlaceOrderRequest) (*PlaceOrderResponse, error) {
	// Створення замовлення
	now := time.Now()
	order := &store.Order{
		ID:           uuid.New().String(),
		CustomerName: req.CustomerName,
		CoffeeType:   req.CoffeeType,
		Status:       "pending",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Додавання замовлення до сховища
	if err := s.orderStore.Add(order); err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, "Failed to add order to store: %v", err)
	}

	// Конвертація замовлення в JSON
	orderJSON, err := json.Marshal(order)
	if err != nil {
		log.Println(err)
		// Видалення замовлення зі сховища, якщо не вдалося конвертувати в JSON
		s.orderStore.Delete(order.ID)
		return nil, status.Errorf(codes.Internal, "Failed to marshal order: %v", err)
	}

	// Відправка замовлення в Kafka
	err = s.kafkaProducer.PushToQueue(s.config.Kafka.Topic, orderJSON)
	if err != nil {
		log.Println(err)
		// Видалення замовлення зі сховища, якщо не вдалося відправити в Kafka
		s.orderStore.Delete(order.ID)
		return nil, status.Errorf(codes.Internal, "Failed to send order to Kafka: %v", err)
	}

	// Створення відповіді
	return &PlaceOrderResponse{
		Success: true,
		Message: "Order for " + req.CustomerName + " placed successfully!",
		Order: &Order{
			Id:           order.ID,
			CustomerName: order.CustomerName,
			CoffeeType:   order.CoffeeType,
			Status:       order.Status,
			CreatedAt:    timestamppb.New(order.CreatedAt),
			UpdatedAt:    timestamppb.New(order.UpdatedAt),
		},
	}, nil
}

// GetOrder обробляє запит на отримання інформації про замовлення
func (s *CoffeeServiceServer) GetOrder(ctx context.Context, req *GetOrderRequest) (*GetOrderResponse, error) {
	// Отримання замовлення зі сховища
	order, err := s.orderStore.Get(req.OrderId)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.NotFound, "Order not found: %v", err)
	}

	// Створення відповіді
	return &GetOrderResponse{
		Success: true,
		Message: "Order retrieved successfully",
		Order: &Order{
			Id:           order.ID,
			CustomerName: order.CustomerName,
			CoffeeType:   order.CoffeeType,
			Status:       order.Status,
			CreatedAt:    timestamppb.New(order.CreatedAt),
			UpdatedAt:    timestamppb.New(order.UpdatedAt),
		},
	}, nil
}

// ListOrders обробляє запит на отримання списку замовлень
func (s *CoffeeServiceServer) ListOrders(ctx context.Context, req *ListOrdersRequest) (*ListOrdersResponse, error) {
	// Отримання списку замовлень зі сховища
	orders, err := s.orderStore.List()
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, "Failed to list orders: %v", err)
	}

	// Конвертація замовлень у формат gRPC
	var protoOrders []*Order
	for _, order := range orders {
		protoOrders = append(protoOrders, &Order{
			Id:           order.ID,
			CustomerName: order.CustomerName,
			CoffeeType:   order.CoffeeType,
			Status:       order.Status,
			CreatedAt:    timestamppb.New(order.CreatedAt),
			UpdatedAt:    timestamppb.New(order.UpdatedAt),
		})
	}

	// Створення відповіді
	return &ListOrdersResponse{
		Success:    true,
		Message:    "Orders retrieved successfully",
		Orders:     protoOrders,
		TotalCount: int32(len(protoOrders)),
	}, nil
}

// CancelOrder обробляє запит на скасування замовлення
func (s *CoffeeServiceServer) CancelOrder(ctx context.Context, req *CancelOrderRequest) (*CancelOrderResponse, error) {
	// Отримання замовлення зі сховища
	order, err := s.orderStore.Get(req.OrderId)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.NotFound, "Order not found: %v", err)
	}

	// Оновлення статусу замовлення
	order.Status = "cancelled"
	order.UpdatedAt = time.Now()

	// Оновлення замовлення в сховищі
	if err := s.orderStore.Update(order); err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, "Failed to update order: %v", err)
	}

	// Конвертація замовлення в JSON
	orderJSON, err := json.Marshal(order)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, "Failed to marshal order: %v", err)
	}

	// Відправка оновленого замовлення в Kafka
	err = s.kafkaProducer.PushToQueue(s.config.Kafka.Topic, orderJSON)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, "Failed to send order to Kafka: %v", err)
	}

	// Створення відповіді
	return &CancelOrderResponse{
		Success: true,
		Message: "Order cancelled successfully",
		Order: &Order{
			Id:           order.ID,
			CustomerName: order.CustomerName,
			CoffeeType:   order.CoffeeType,
			Status:       order.Status,
			CreatedAt:    timestamppb.New(order.CreatedAt),
			UpdatedAt:    timestamppb.New(order.UpdatedAt),
		},
	}, nil
}
