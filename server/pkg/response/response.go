package response

import (
	"encoding/json"
	"net/http"
)

// Response представляет стандартный формат ответа API
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *Errors      `json:"error,omitempty"`
}

// Error представляет структуру ошибки в ответе API
type Errors struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// JSON отправляет ответ в формате JSON
func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	resp := Response{
		Success: statusCode < 400,
		Data:    data,
	}
	
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// Error отправляет ошибку в формате JSON
func Error(w http.ResponseWriter, statusCode int, code, message string, details ...interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	var detail interface{}
	if len(details) > 0 {
		detail = details[0]
	}
	
	resp := Response{
		Success: false,
		Error: &Errors{
			Code:    code,
			Message: message,
			Details: detail,
		},
	}
	
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
	}
}

// ErrorWithCode отправляет ошибку с предопределенным кодом
func ErrorWithCode(w http.ResponseWriter, statusCode int, code, message string) {
	Error(w, statusCode, code, message, nil)
}

// ValidationError отправляет ошибку валидации
func ValidationError(w http.ResponseWriter, field, message string) {
	details := map[string]string{
		"field": field,
		"error": message,
	}
	Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", details)
}

// NotFound отправляет ошибку 404
func NotFound(w http.ResponseWriter) {
	ErrorWithCode(w, http.StatusNotFound, "NOT_FOUND", "Resource not found")
}

// InternalServerError отправляет ошибку 500
func InternalServerError(w http.ResponseWriter, err error) {
	Error(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal server error", err.Error())
}