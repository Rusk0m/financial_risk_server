package handlers

import (
	"financial-risk-server/internal/service"
	"financial-risk-server/pkg/response"
	"net/http"
	"strconv"
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

// GetEnterpriseByID возвращает предприятие по ID
func (h *EnterpriseHandler) GetEnterpriseByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.ValidationError(w, "id", "Invalid enterprise ID")
		return
	}

	enterprise, err := h.enterpriseService.GetEnterpriseByID(r.Context(), id)
	if err != nil {
		response.NotFound(w)
		return
	}

	response.JSON(w, http.StatusOK, enterprise)
}