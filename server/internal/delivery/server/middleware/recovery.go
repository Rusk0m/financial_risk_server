package middleware

import (
	"log"
	"net/http"
	"runtime/debug"
)

// RecoveryMiddleware перехватывает паники и предотвращает падение сервера
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Логируем панику с трейсбэком
				log.Printf("💥 Паника при обработке запроса [%s] %s: %v\n%s",
					r.Method, r.URL.Path, err, debug.Stack())

				// Отправляем клиенту ошибку 500
				http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
			}
		}()

		// Выполняем следующий обработчик
		next.ServeHTTP(w, r)
	})
}
