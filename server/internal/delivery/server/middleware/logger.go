package middleware

import (
	"log"
	"net/http"
	"time"
)

// LoggerMiddleware логирует все входящие запросы
func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Запоминаем время начала запроса
		start := time.Now()

		// Логируем начало запроса
		log.Printf("📥 [%s] %s %s", r.Method, r.RemoteAddr, r.URL.Path)

		// Выполняем следующий обработчик в цепочке
		next.ServeHTTP(w, r)

		// Логируем завершение запроса с временем выполнения
		duration := time.Since(start)
		log.Printf("📤 [%s] %s %s - %v", r.Method, r.RemoteAddr, r.URL.Path, duration)
	})
}
