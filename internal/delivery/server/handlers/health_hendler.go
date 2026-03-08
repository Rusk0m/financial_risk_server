package handlers

import (
	"financial-risk-server/pkg/response"
	"net/http"
)

// HealthHandler обрабатывает запросы проверки работоспособности
type HealthHandler struct{}

// NewHealthHandler создаёт новый обработчик проверки работоспособности
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Check проверяет работоспособность сервера
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"message": "Сервер работает",
		"version": "1.0.0",
	})
}
