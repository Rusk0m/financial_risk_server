package handlers

import (
	"context"
	"errors"
	"financial-risk-server/internal/service"
	"financial-risk-server/pkg/response"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// EnterpriseHandler обрабатывает запросы, связанные с предприятиями
type EnterpriseHandler struct {
	enterpriseService *service.EnterpriseService
}

// NewEnterpriseHandler создаёт новый хэндлер предприятий
func NewEnterpriseHandler(enterpriseService *service.EnterpriseService) *EnterpriseHandler {
	return &EnterpriseHandler{
		enterpriseService: enterpriseService,
	}
}


// handlers/enterprise_handler.go
func (h *EnterpriseHandler) GetEnterpriseByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	log.Printf("📥 [Handler] Запрос: GET /enterprises/%s", idStr)
	
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		log.Printf("⚠️  [Handler] Ошибка парсинга ID: %v", err)
		response.ValidationError(w, "id", "Invalid enterprise ID format")
		return
	}

	enterprise, err := h.enterpriseService.GetEnterpriseByID(r.Context(), id)
	if err != nil {
		// 🔍 Логируем реальную ошибку перед возвратом 404
		log.Printf("❌ [Handler] Ошибка сервиса: %v", err)
		
		// Возвращаем 404 только если действительно "не найдено"
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			response.NotFound(w)
			return
		}
		// Для других ошибок — 500 с деталями
		response.InternalServerError(w, err)
		return
	}

	log.Printf("✅ [Handler] Успешный ответ: предприятие найдено")
	response.JSON(w, http.StatusOK, enterprise)
}
// isNotFound проверяет, является ли ошибкой "не найдено"
func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "not found") || 
		   strings.Contains(errStr, "no rows") ||
		   errors.Is(err, context.Canceled) == false && 
		   (errors.Is(err, context.DeadlineExceeded) == false)
}