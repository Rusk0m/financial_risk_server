package validator

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// ValidationError представляет ошибку валидации одного поля
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors представляет коллекцию ошибок валидации
type ValidationErrors []ValidationError

// Error реализует интерфейс error
func (ve ValidationErrors) Error() string {
	messages := make([]string, len(ve))
	for i, err := range ve {
		messages[i] = err.Field + ": " + err.Message
	}
	return strings.Join(messages, "; ")
}

// ToMap преобразует ошибки валидации в карту для ответа API
func (ve ValidationErrors) ToMap() map[string]string {
	result := make(map[string]string)
	for _, err := range ve {
		result[err.Field] = err.Message
	}
	return result
}

// Validator предоставляет методы для валидации данных
type Validator struct{}

// NewValidator создаёт новый валидатор
func NewValidator() *Validator {
	return &Validator{}
}

// Required проверяет, что строка не пустая
func (v *Validator) Required(value string, field string) *ValidationErrors {
	var errors ValidationErrors
	if strings.TrimSpace(value) == "" {
		errors = append(errors, ValidationError{
			Field:   field,
			Message: "поле не может быть пустым",
		})
	}
	return &errors
}

// MinLength проверяет минимальную длину строки
func (v *Validator) MinLength(value string, minLength int, field string) *ValidationErrors {
	var errors ValidationErrors
	if utf8.RuneCountInString(value) < minLength {
		errors = append(errors, ValidationError{
			Field:   field,
			Message: "минимальная длина " + string(rune('0'+minLength)),
		})
	}
	return &errors
}

// MaxLength проверяет максимальную длину строки
func (v *Validator) MaxLength(value string, maxLength int, field string) *ValidationErrors {
	var errors ValidationErrors
	if utf8.RuneCountInString(value) > maxLength {
		errors = append(errors, ValidationError{
			Field:   field,
			Message: "максимальная длина " + string(rune('0'+maxLength)),
		})
	}
	return &errors
}

// MinValue проверяет минимальное значение числа
func (v *Validator) MinValue(value float64, minValue float64, field string) *ValidationErrors {
	var errors ValidationErrors
	if value < minValue {
		errors = append(errors, ValidationError{
			Field:   field,
			Message: "значение должно быть не меньше " + string(rune('0'+int(minValue))),
		})
	}
	return &errors
}

// MaxValue проверяет максимальное значение числа
func (v *Validator) MaxValue(value float64, maxValue float64, field string) *ValidationErrors {
	var errors ValidationErrors
	if value > maxValue {
		errors = append(errors, ValidationError{
			Field:   field,
			Message: "значение должно быть не больше " + string(rune('0'+int(maxValue))),
		})
	}
	return &errors
}

// CurrencyCode проверяет корректность кода валюты (ISO 4217)
func (v *Validator) CurrencyCode(value string, field string) *ValidationErrors {
	var errors ValidationErrors

	// Код валюты должен состоять из 3 букв
	if len(value) != 3 {
		errors = append(errors, ValidationError{
			Field:   field,
			Message: "код валюты должен состоять из 3 символов",
		})
	}

	// Проверяем, что все символы — буквы
	currencyRegex := regexp.MustCompile(`^[A-Z]{3}$`)
	if !currencyRegex.MatchString(value) {
		errors = append(errors, ValidationError{
			Field:   field,
			Message: "код валюты должен состоять из 3 заглавных латинских букв",
		})
	}

	// Проверяем известные валюты (упрощённо)
	knownCurrencies := map[string]bool{
		"USD": true, "EUR": true, "BYN": true, "RUB": true, "CNY": true,
	}
	if !knownCurrencies[value] {
		errors = append(errors, ValidationError{
			Field:   field,
			Message: "неизвестная валюта (поддерживаются: USD, EUR, BYN, RUB, CNY)",
		})
	}

	return &errors
}

// MergeErrors объединяет несколько списков ошибок
func (v *Validator) MergeErrors(errors ...*ValidationErrors) *ValidationErrors {
	var result ValidationErrors
	for _, errList := range errors {
		if errList != nil && len(*errList) > 0 {
			result = append(result, *errList...)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return &result
}
