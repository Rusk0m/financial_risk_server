package response

import (
	"encoding/json"
	"net/http"
)

// Response представляет стандартный формат ответа API
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *Errors     `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// Error представляет информацию об ошибке
type Errors struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// Meta содержит метаинформацию (пагинация, время выполнения и т.д.)
type Meta struct {
	Page       int `json:"page,omitempty"`
	PerPage    int `json:"per_page,omitempty"`
	Total      int `json:"total,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

// JSON отправляет ответ в формате JSON с указанным статус-кодом
func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	resp := Response{
		Success: statusCode >= 200 && statusCode < 300,
		Data:    data,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Ошибка сериализации ответа", http.StatusInternalServerError)
	}
}

// Error отправляет ответ с ошибкой
func Error(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	resp := Response{
		Success: false,
		Error: &Errors{
			Code:    http.StatusText(statusCode),
			Message: message,
		},
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Ошибка сериализации ответа", http.StatusInternalServerError)
	}
}

// ErrorWithCode отправляет ответ с ошибкой и кастомным кодом
func ErrorWithCode(w http.ResponseWriter, statusCode int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	resp := Response{
		Success: false,
		Error: &Errors{
			Code:    code,
			Message: message,
		},
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Ошибка сериализации ответа", http.StatusInternalServerError)
	}
}

// ValidationError отправляет ответ с ошибками валидации
func ValidationError(w http.ResponseWriter, errors map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	resp := Response{
		Success: false,
		Error: &Errors{
			Code:    "VALIDATION_ERROR",
			Message: "Ошибки валидации полей",
			Details: errors,
		},
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Ошибка сериализации ответа", http.StatusInternalServerError)
	}
}
