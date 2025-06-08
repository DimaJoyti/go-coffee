package payment

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/DimaJoyti/go-coffee/pkg/models"
)

// SetupRoutes configures the HTTP routes for the payment service
func SetupRoutes(mux *http.ServeMux, service *Service) {
	// Wallet operations
	mux.HandleFunc("/api/v1/payment/wallet/create", methodHandler("POST", createWalletHandler(service)))
	mux.HandleFunc("/api/v1/payment/wallet/import", methodHandler("POST", importWalletHandler(service)))
	mux.HandleFunc("/api/v1/payment/wallet/validate", methodHandler("POST", validateAddressHandler(service)))

	// Multisig operations
	mux.HandleFunc("/api/v1/payment/multisig/create", methodHandler("POST", createMultisigHandler(service)))

	// Message signing
	mux.HandleFunc("/api/v1/payment/message/sign", methodHandler("POST", signMessageHandler(service)))
	mux.HandleFunc("/api/v1/payment/message/verify", methodHandler("POST", verifyMessageHandler(service)))

	// Service info
	mux.HandleFunc("/api/v1/payment/features", methodHandler("GET", getFeaturesHandler(service)))
	mux.HandleFunc("/api/v1/payment/version", methodHandler("GET", getVersionHandler(service)))
}

// methodHandler wraps handlers to only accept specific HTTP methods
func methodHandler(method string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			writeErrorResponse(w, http.StatusMethodNotAllowed, fmt.Sprintf("Method %s not allowed", r.Method))
			return
		}
		
		// Add CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		handler(w, r)
	}
}

// writeJSONResponse writes a JSON response
func writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// writeErrorResponse writes an error response
func writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	writeJSONResponse(w, statusCode, map[string]interface{}{
		"success": false,
		"error":   message,
	})
}

// writeSuccessResponse writes a success response
func writeSuccessResponse(w http.ResponseWriter, data interface{}) {
	writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

// createWalletHandler handles wallet creation requests
func createWalletHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Testnet bool `json:"testnet"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "Invalid request format")
			return
		}

		wallet, err := service.CreateWallet(r.Context(), req.Testnet)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeSuccessResponse(w, wallet)
	}
}

// importWalletHandler handles wallet import requests
func importWalletHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			WIF string `json:"wif"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "Invalid request format")
			return
		}

		if req.WIF == "" {
			writeErrorResponse(w, http.StatusBadRequest, "WIF is required")
			return
		}

		wallet, err := service.ImportWallet(r.Context(), req.WIF)
		if err != nil {
			writeErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		writeSuccessResponse(w, wallet)
	}
}

// validateAddressHandler handles address validation requests
func validateAddressHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Address string `json:"address"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "Invalid request format")
			return
		}

		if req.Address == "" {
			writeErrorResponse(w, http.StatusBadRequest, "Address is required")
			return
		}

		validation, err := service.ValidateAddress(r.Context(), req.Address)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeSuccessResponse(w, validation)
	}
}

// createMultisigHandler handles multisig address creation requests
func createMultisigHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.MultisigRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "Invalid request format")
			return
		}

		if len(req.PublicKeys) < 2 {
			writeErrorResponse(w, http.StatusBadRequest, "At least 2 public keys required")
			return
		}

		if req.Threshold < 1 {
			writeErrorResponse(w, http.StatusBadRequest, "Threshold must be at least 1")
			return
		}

		multisigAddr, err := service.CreateMultisigAddress(r.Context(), &req)
		if err != nil {
			writeErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		writeSuccessResponse(w, multisigAddr)
	}
}

// signMessageHandler handles message signing requests
func signMessageHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.SignMessageRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "Invalid request format")
			return
		}

		if req.Message == "" {
			writeErrorResponse(w, http.StatusBadRequest, "Message is required")
			return
		}

		if req.PrivateKey == "" {
			writeErrorResponse(w, http.StatusBadRequest, "Private key is required")
			return
		}

		response, err := service.SignMessage(r.Context(), &req)
		if err != nil {
			writeErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		writeSuccessResponse(w, response)
	}
}

// verifyMessageHandler handles message verification requests
func verifyMessageHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.VerifyMessageRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "Invalid request format")
			return
		}

		if req.Message == "" || req.Signature == "" || req.Address == "" {
			writeErrorResponse(w, http.StatusBadRequest, "Message, signature, and address are required")
			return
		}

		response, err := service.VerifyMessage(r.Context(), &req)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeSuccessResponse(w, response)
	}
}

// getFeaturesHandler returns supported Bitcoin features
func getFeaturesHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		features := service.GetSupportedFeatures(r.Context())

		writeSuccessResponse(w, map[string]interface{}{
			"features": features,
			"count":    len(features),
		})
	}
}

// getVersionHandler returns the service version
func getVersionHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		version := service.GetVersion(r.Context())

		writeSuccessResponse(w, map[string]interface{}{
			"version": version,
			"service": "payment-service",
		})
	}
}
