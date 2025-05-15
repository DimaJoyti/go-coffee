package handler

import (
	"log"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/google/uuid"
)

// Middleware представляє HTTP middleware
type Middleware func(http.Handler) http.Handler

// Chain об'єднує middleware в ланцюжок
func Chain(middlewares ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

// LoggingMiddleware логує HTTP запити
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Отримання request ID з контексту або створення нового
		requestID := r.Context().Value("requestID")
		if requestID == nil {
			requestID = uuid.New().String()
		}

		// Логування запиту
		log.Printf("[%s] %s %s %s", requestID, r.Method, r.URL.Path, r.RemoteAddr)

		// Створення response writer, який відстежує статус код
		rw := &responseWriter{w, http.StatusOK}

		// Виклик наступного обробника
		next.ServeHTTP(rw, r)

		// Логування відповіді
		log.Printf("[%s] %s %s %s - %d %s", requestID, r.Method, r.URL.Path, r.RemoteAddr, rw.statusCode, time.Since(start))
	})
}

// RequestIDMiddleware додає request ID до контексту
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Отримання request ID з заголовка або створення нового
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Додавання request ID до заголовків відповіді
		w.Header().Set("X-Request-ID", requestID)

		// Виклик наступного обробника
		next.ServeHTTP(w, r)
	})
}

// RecoverMiddleware відновлює після паніки
func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Логування паніки
				log.Printf("Panic: %v\n%s", err, debug.Stack())

				// Відправка помилки клієнту
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"success":false,"message":"Internal server error"}`))
			}
		}()

		// Виклик наступного обробника
		next.ServeHTTP(w, r)
	})
}

// CORSMiddleware додає CORS заголовки
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Додавання CORS заголовків
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")

		// Обробка OPTIONS запиту
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Виклик наступного обробника
		next.ServeHTTP(w, r)
	})
}

// responseWriter обгортка для http.ResponseWriter, яка відстежує статус код
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader перевизначає метод WriteHeader для відстеження статус коду
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
