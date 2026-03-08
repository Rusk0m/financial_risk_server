package middleware

import (
	"net/http"
)

// CORSMiddleware настраивает заголовки CORS для кросс-доменных запросов
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Разрешаем запросы с любого источника (для разработки)
		// В продакшне нужно ограничить конкретными доменами
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "3600")

		// Обработка preflight-запросов (запросы OPTIONS)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Передаём управление следующему обработчику
		next.ServeHTTP(w, r)
	})
}
