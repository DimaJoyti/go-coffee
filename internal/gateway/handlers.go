package gateway

import (
	"io"
	"net/http"
	"strings"
)

// SetupRoutes configures the HTTP routes for the API gateway
func SetupRoutes(mux *http.ServeMux, service *Service) {
	// Apply middleware chain
	chain := func(handler http.HandlerFunc) http.HandlerFunc {
		return service.CORSMiddleware(
			service.LoggingMiddleware(
				service.RateLimitMiddleware(handler),
			),
		)
	}

	// Gateway management endpoints
	mux.HandleFunc("/api/v1/gateway/status", chain(getGatewayStatusHandler(service)))
	mux.HandleFunc("/api/v1/gateway/services", chain(getServicesStatusHandler(service)))

	// Proxy all other API requests
	mux.HandleFunc("/api/", chain(proxyHandler(service)))

	// Serve API documentation
	mux.HandleFunc("/docs", chain(docsHandler(service)))
	mux.HandleFunc("/docs/", chain(docsHandler(service)))
}

// proxyHandler handles proxying requests to appropriate services
func proxyHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Route the request to determine target service
		serviceName, servicePath, err := service.RouteRequest(r.URL.Path)
		if err != nil {
			writeErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// Read request body
		var body []byte
		if r.Body != nil {
			body, err = io.ReadAll(r.Body)
			if err != nil {
				writeErrorResponse(w, http.StatusBadRequest, "Failed to read request body")
				return
			}
		}

		// Prepare headers
		headers := make(map[string]string)
		for key, values := range r.Header {
			if len(values) > 0 {
				headers[key] = values[0]
			}
		}

		// Proxy the request
		resp, err := service.ProxyRequest(r.Context(), serviceName, servicePath, r.Method, body, headers)
		if err != nil {
			writeErrorResponse(w, http.StatusBadGateway, err.Error())
			return
		}
		defer resp.Body.Close()

		// Copy response headers
		for key, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}

		// Set status code
		w.WriteHeader(resp.StatusCode)

		// Copy response body
		if _, err := io.Copy(w, resp.Body); err != nil {
			service.logger.Error("Failed to copy response body: %v", err)
		}
	}
}

// getGatewayStatusHandler returns the gateway status
func getGatewayStatusHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		status := map[string]interface{}{
			"gateway":  "go-coffee-api-gateway",
			"version":  "1.0.0",
			"status":   "healthy",
			"uptime":   "running", // In production, calculate actual uptime
			"services": len(service.services),
		}

		writeJSONResponse(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"data":    status,
		})
	}
}

// getServicesStatusHandler returns the status of all services
func getServicesStatusHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		servicesStatus := service.GetServiceStatus(r.Context())

		writeJSONResponse(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"data":    servicesStatus,
		})
	}
}

// docsHandler serves API documentation
func docsHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		// Simple API documentation
		docs := map[string]interface{}{
			"title":       "Go Coffee API Gateway",
			"version":     "1.0.0",
			"description": "API Gateway for Go Coffee microservices",
			"services": map[string]interface{}{
				"auth": map[string]interface{}{
					"base_url":    "/api/v1/auth",
					"description": "Authentication and user management",
					"endpoints": []string{
						"POST /api/v1/auth/register",
						"POST /api/v1/auth/login",
						"POST /api/v1/auth/refresh",
						"GET /api/v1/auth/profile",
					},
				},
				"payment": map[string]interface{}{
					"base_url":    "/api/v1/payment",
					"description": "Bitcoin payment processing",
					"endpoints": []string{
						"POST /api/v1/payment/wallet/create",
						"POST /api/v1/payment/wallet/import",
						"POST /api/v1/payment/wallet/validate",
						"POST /api/v1/payment/message/sign",
						"GET /api/v1/payment/features",
					},
				},
				"order": map[string]interface{}{
					"base_url":    "/api/v1/order",
					"description": "Coffee order management",
					"endpoints": []string{
						"GET /api/v1/order/menu",
						"POST /api/v1/order/create",
						"GET /api/v1/order/{id}",
						"PUT /api/v1/order/{id}/status",
					},
				},
				"kitchen": map[string]interface{}{
					"base_url":    "/api/v1/kitchen",
					"description": "Kitchen operations and order processing",
					"endpoints": []string{
						"GET /api/v1/kitchen/queue",
						"PUT /api/v1/kitchen/order/{id}/start",
						"PUT /api/v1/kitchen/order/{id}/complete",
						"GET /api/v1/kitchen/inventory",
					},
				},
			},
			"gateway": map[string]interface{}{
				"endpoints": []string{
					"GET /api/v1/gateway/status",
					"GET /api/v1/gateway/services",
					"GET /docs",
				},
			},
		}

		// Check if request wants HTML documentation
		if strings.Contains(r.Header.Get("Accept"), "text/html") {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)

			html := generateHTMLDocs(docs)
			w.Write([]byte(html))
			return
		}

		// Return JSON documentation
		writeJSONResponse(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"data":    docs,
		})
	}
}

// generateHTMLDocs generates HTML documentation
func generateHTMLDocs(docs map[string]interface{}) string {
	return `
<!DOCTYPE html>
<html>
<head>
    <title>Go Coffee API Documentation</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        h1 { color: #8B4513; }
        h2 { color: #D2691E; }
        .service { margin: 20px 0; padding: 15px; border: 1px solid #ddd; border-radius: 5px; }
        .endpoint { background: #f5f5f5; padding: 5px; margin: 5px 0; border-radius: 3px; }
        .method { font-weight: bold; color: #2E8B57; }
    </style>
</head>
<body>
    <h1>☕ Go Coffee API Gateway</h1>
    <p>Welcome to the Go Coffee API documentation. This gateway coordinates all microservices.</p>
    
    <h2>🔗 Available Services</h2>
    
    <div class="service">
        <h3>🔐 Authentication Service</h3>
        <p>Handles user authentication and authorization</p>
        <div class="endpoint"><span class="method">POST</span> /api/v1/auth/register</div>
        <div class="endpoint"><span class="method">POST</span> /api/v1/auth/login</div>
        <div class="endpoint"><span class="method">POST</span> /api/v1/auth/refresh</div>
        <div class="endpoint"><span class="method">GET</span> /api/v1/auth/profile</div>
    </div>
    
    <div class="service">
        <h3>₿ Payment Service</h3>
        <p>Bitcoin payment processing and wallet management</p>
        <div class="endpoint"><span class="method">POST</span> /api/v1/payment/wallet/create</div>
        <div class="endpoint"><span class="method">POST</span> /api/v1/payment/wallet/import</div>
        <div class="endpoint"><span class="method">POST</span> /api/v1/payment/wallet/validate</div>
        <div class="endpoint"><span class="method">POST</span> /api/v1/payment/message/sign</div>
        <div class="endpoint"><span class="method">GET</span> /api/v1/payment/features</div>
    </div>
    
    <div class="service">
        <h3>📋 Order Service</h3>
        <p>Coffee order management and tracking</p>
        <div class="endpoint"><span class="method">GET</span> /api/v1/order/menu</div>
        <div class="endpoint"><span class="method">POST</span> /api/v1/order/create</div>
        <div class="endpoint"><span class="method">GET</span> /api/v1/order/{id}</div>
        <div class="endpoint"><span class="method">PUT</span> /api/v1/order/{id}/status</div>
    </div>
    
    <div class="service">
        <h3>👨‍🍳 Kitchen Service</h3>
        <p>Kitchen operations and order processing</p>
        <div class="endpoint"><span class="method">GET</span> /api/v1/kitchen/queue</div>
        <div class="endpoint"><span class="method">PUT</span> /api/v1/kitchen/order/{id}/start</div>
        <div class="endpoint"><span class="method">PUT</span> /api/v1/kitchen/order/{id}/complete</div>
        <div class="endpoint"><span class="method">GET</span> /api/v1/kitchen/inventory</div>
    </div>
    
    <h2>🔧 Gateway Endpoints</h2>
    <div class="service">
        <div class="endpoint"><span class="method">GET</span> /api/v1/gateway/status</div>
        <div class="endpoint"><span class="method">GET</span> /api/v1/gateway/services</div>
        <div class="endpoint"><span class="method">GET</span> /docs</div>
    </div>
    
    <h2>🚀 Getting Started</h2>
    <p>All requests should be made to the API Gateway at <code>http://localhost:8080</code></p>
    <p>Authentication is required for most endpoints. Include the JWT token in the Authorization header:</p>
    <pre>Authorization: Bearer &lt;your-jwt-token&gt;</pre>
</body>
</html>
`
}
