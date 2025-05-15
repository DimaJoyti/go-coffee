package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"api_gateway/client"
	"api_gateway/config"
)

// HTTPServer представляє HTTP сервер для API Gateway
type HTTPServer struct {
	*http.Server
	config         *config.Config
	producerClient *client.CoffeeClient
}

// OrderRequest представляє запит на створення замовлення
type OrderRequest struct {
	CustomerName string `json:"customer_name"`
	CoffeeType   string `json:"coffee_type"`
}

// Order представляє замовлення
type Order struct {
	ID           string    `json:"id"`
	CustomerName string    `json:"customer_name"`
	CoffeeType   string    `json:"coffee_type"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Response представляє відповідь API
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"msg"`
	Order   *Order `json:"order,omitempty"`
}

// NewHTTPServer створює новий HTTP сервер для API Gateway
func NewHTTPServer(config *config.Config, producerClient *client.CoffeeClient) *HTTPServer {
	server := &HTTPServer{
		config:         config,
		producerClient: producerClient,
	}

	// Створення маршрутизатора
	mux := http.NewServeMux()

	// Реєстрація маршрутів
	mux.HandleFunc("/order", server.handleOrder)
	mux.HandleFunc("/order/", server.handleOrderWithID)
	mux.HandleFunc("/orders", server.handleListOrders)
	mux.HandleFunc("/health", server.handleHealth)

	// Створення HTTP сервера
	server.Server = &http.Server{
		Addr:         ":" + string(config.Server.Port),
		Handler:      server.logMiddleware(server.corsMiddleware(mux)),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return server
}

// handleOrder обробляє запити на створення замовлення
func (s *HTTPServer) handleOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var orderReq OrderRequest
	if err := json.NewDecoder(r.Body).Decode(&orderReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Валідація запиту
	if orderReq.CustomerName == "" || orderReq.CoffeeType == "" {
		http.Error(w, "Customer name and coffee type are required", http.StatusBadRequest)
		return
	}

	// Створення замовлення
	order := &Order{
		ID:           uuid.New().String(),
		CustomerName: orderReq.CustomerName,
		CoffeeType:   orderReq.CoffeeType,
		Status:       "pending",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// TODO: Відправка замовлення через gRPC до Producer сервісу
	// Це буде реалізовано після створення gRPC клієнта

	// Відповідь
	response := Response{
		Success: true,
		Message: "Order for " + orderReq.CustomerName + " placed successfully!",
		Order:   order,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println(err)
		http.Error(w, "Error placing order", http.StatusInternalServerError)
		return
	}
}

// handleOrderWithID обробляє запити на отримання або скасування замовлення за ID
func (s *HTTPServer) handleOrderWithID(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "/order/" {
		http.Error(w, "Order ID is required", http.StatusBadRequest)
		return
	}

	// Отримання ID замовлення з URL
	orderID := strings.TrimPrefix(path, "/order/")

	// Перевірка, чи це запит на скасування замовлення
	if strings.HasSuffix(orderID, "/cancel") {
		orderID = strings.TrimSuffix(orderID, "/cancel")
		if r.Method == http.MethodPost {
			s.handleCancelOrder(w, r, orderID)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Обробка запиту на отримання інформації про замовлення
	if r.Method == http.MethodGet {
		s.handleGetOrder(w, r, orderID)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// handleGetOrder обробляє запити на отримання інформації про замовлення
func (s *HTTPServer) handleGetOrder(w http.ResponseWriter, r *http.Request, orderID string) {
	// TODO: Отримання інформації про замовлення через gRPC

	// Заглушка для демонстрації
	order := &Order{
		ID:           orderID,
		CustomerName: "John Doe",
		CoffeeType:   "Latte",
		Status:       "pending",
		CreatedAt:    time.Now().Add(-5 * time.Minute),
		UpdatedAt:    time.Now().Add(-5 * time.Minute),
	}

	response := Response{
		Success: true,
		Message: "Order retrieved successfully",
		Order:   order,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println(err)
		http.Error(w, "Error retrieving order", http.StatusInternalServerError)
		return
	}
}

// handleCancelOrder обробляє запити на скасування замовлення
func (s *HTTPServer) handleCancelOrder(w http.ResponseWriter, r *http.Request, orderID string) {
	// TODO: Скасування замовлення через gRPC

	// Заглушка для демонстрації
	order := &Order{
		ID:           orderID,
		CustomerName: "John Doe",
		CoffeeType:   "Latte",
		Status:       "cancelled",
		CreatedAt:    time.Now().Add(-5 * time.Minute),
		UpdatedAt:    time.Now(),
	}

	response := Response{
		Success: true,
		Message: "Order cancelled successfully",
		Order:   order,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println(err)
		http.Error(w, "Error cancelling order", http.StatusInternalServerError)
		return
	}
}

// handleListOrders обробляє запити на отримання списку замовлень
func (s *HTTPServer) handleListOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Отримання списку замовлень через gRPC

	// Заглушка для демонстрації
	orders := []*Order{
		{
			ID:           uuid.New().String(),
			CustomerName: "John Doe",
			CoffeeType:   "Latte",
			Status:       "pending",
			CreatedAt:    time.Now().Add(-5 * time.Minute),
			UpdatedAt:    time.Now().Add(-5 * time.Minute),
		},
		{
			ID:           uuid.New().String(),
			CustomerName: "Jane Smith",
			CoffeeType:   "Espresso",
			Status:       "completed",
			CreatedAt:    time.Now().Add(-10 * time.Minute),
			UpdatedAt:    time.Now().Add(-5 * time.Minute),
		},
	}

	response := struct {
		Success    bool     `json:"success"`
		Message    string   `json:"msg"`
		Orders     []*Order `json:"orders"`
		TotalCount int      `json:"total_count"`
	}{
		Success:    true,
		Message:    "Orders retrieved successfully",
		Orders:     orders,
		TotalCount: len(orders),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println(err)
		http.Error(w, "Error retrieving orders", http.StatusInternalServerError)
		return
	}
}

// handleHealth обробляє запити на перевірку стану сервера
func (s *HTTPServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}{
		Status:  "ok",
		Message: "API Gateway is running",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println(err)
		http.Error(w, "Error checking health", http.StatusInternalServerError)
		return
	}
}

// logMiddleware логує інформацію про запити
func (s *HTTPServer) logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := uuid.New().String()

		// Додавання requestID до контексту
		ctx := context.WithValue(r.Context(), "requestID", requestID)
		r = r.WithContext(ctx)

		// Логування запиту
		log.Printf("[%s] %s %s %s", requestID, r.Method, r.URL.Path, r.RemoteAddr)

		// Виклик наступного обробника
		next.ServeHTTP(w, r)

		// Логування часу виконання
		log.Printf("[%s] Completed in %v", requestID, time.Since(start))
	})
}

// corsMiddleware додає CORS заголовки до відповідей
func (s *HTTPServer) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Додавання CORS заголовків
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Обробка OPTIONS запитів
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Виклик наступного обробника
		next.ServeHTTP(w, r)
	})
}
