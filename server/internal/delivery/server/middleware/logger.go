// internal/middleware/logger.go
package middleware

import (
	"log"
	"net/http"
	"time"
)

// responseWriter оборачивает http.ResponseWriter для перехвата статуса
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// LoggerMiddleware логирует все входящие запросы
func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Оборачиваем ResponseWriter
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		// Выполняем запрос
		next.ServeHTTP(wrapped, r)
		
		// Логируем результат
		duration := time.Since(start)
		log.Printf("📥 [%s] %s %s → %d | %v",
			r.Method,
			r.RemoteAddr,
			r.URL.Path,
			wrapped.statusCode,
			duration,
		)
		
		// Логируем тело ошибки для 4xx/5xx
		if wrapped.statusCode >= 400 {
			log.Printf("⚠️  Error response for %s %s", r.Method, r.URL.Path)
		}
	})
}