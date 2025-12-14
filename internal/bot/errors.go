package bot

import (
	"errors"
	"fmt"
)

// Стандартные ошибки бота.
var (
	ErrUserNotInGroup   = errors.New("пользователь не в группе")
	ErrGroupNotExists   = errors.New("группа не существует")
	ErrGroupAlreadyExists = errors.New("группа уже существует")
	ErrRuleNotFound     = errors.New("правило не найдено")
	ErrNotRuleOwner     = errors.New("вы не владелец этого правила")
	ErrInvalidInput     = errors.New("некорректные входные данные")
	ErrAPIUnavailable   = errors.New("API недоступен")
)

// APIError представляет ошибку от API.
type APIError struct {
	StatusCode int
	Message    string
	Details    string
}

func (e *APIError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("API ошибка (%d): %s - %s", e.StatusCode, e.Message, e.Details)
	}
	return fmt.Sprintf("API ошибка (%d): %s", e.StatusCode, e.Message)
}

// NewAPIError создаёт новую ошибку API.
func NewAPIError(statusCode int, message, details string) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Message:    message,
		Details:    details,
	}
}

// ParseError представляет ошибку парсинга.
type ParseError struct {
	Field   string
	Message string
}

func (e *ParseError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("ошибка парсинга поля '%s': %s", e.Field, e.Message)
	}
	return fmt.Sprintf("ошибка парсинга: %s", e.Message)
}

// NewParseError создаёт новую ошибку парсинга.
func NewParseError(field, message string) *ParseError {
	return &ParseError{
		Field:   field,
		Message: message,
	}
}

// ValidationError представляет ошибку валидации.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// IsAPIError проверяет, является ли ошибка ошибкой API.
func IsAPIError(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr)
}

// IsParseError проверяет, является ли ошибка ошибкой парсинга.
func IsParseError(err error) bool {
	var parseErr *ParseError
	return errors.As(err, &parseErr)
}

