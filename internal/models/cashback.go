package models

import (
	"time"
)

// CashbackRule представляет правило кэшбэка
type CashbackRule struct {
	ID              int64     `json:"id"`
	GroupName       string    `json:"group_name"`
	Category        string    `json:"category"`
	BankName        string    `json:"bank_name"`
	UserID          string    `json:"user_id"`
	UserDisplayName string    `json:"user_display_name"`
	MonthYear       time.Time `json:"month_year"`
	CashbackPercent float64   `json:"cashback_percent"`
	MaxAmount       float64   `json:"max_amount"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// CreateCashbackRequest представляет запрос на создание правила
type CreateCashbackRequest struct {
	GroupName       string  `json:"group_name"`
	Category        string  `json:"category"`
	BankName        string  `json:"bank_name"`
	UserID          string  `json:"user_id"`
	UserDisplayName string  `json:"user_display_name"`
	MonthYear       string  `json:"month_year"`
	CashbackPercent float64 `json:"cashback_percent"`
	MaxAmount       float64 `json:"max_amount"`
	Force           bool    `json:"force,omitempty"`
}

// UpdateCashbackRequest представляет запрос на обновление правила
type UpdateCashbackRequest struct {
	GroupName       string  `json:"group_name"`
	Category        string  `json:"category"`
	BankName        string  `json:"bank_name"`
	MonthYear       string  `json:"month_year"`
	CashbackPercent float64 `json:"cashback_percent"`
	MaxAmount       float64 `json:"max_amount"`
}

// SuggestRequest представляет запрос на анализ данных
type SuggestRequest struct {
	GroupName       string  `json:"group_name"`
	Category        string  `json:"category"`
	BankName        string  `json:"bank_name"`
	UserDisplayName string  `json:"user_display_name"`
	MonthYear       string  `json:"month_year"`
	CashbackPercent float64 `json:"cashback_percent"`
	MaxAmount       float64 `json:"max_amount"`
}

// FuzzySuggestion представляет предложение для исправления
type FuzzySuggestion struct {
	Value      string  `json:"value"`
	Similarity float64 `json:"similarity"`
}

// SuggestResponse представляет ответ с предложениями
type SuggestResponse struct {
	Valid        bool              `json:"valid"`
	Errors       []string          `json:"errors,omitempty"`
	Suggestions  Suggestions       `json:"suggestions"`
	CanProceed   bool              `json:"can_proceed"`
}

// Suggestions содержит предложения по всем полям
type Suggestions struct {
	GroupName       []FuzzySuggestion `json:"group_name,omitempty"`
	Category        []FuzzySuggestion `json:"category,omitempty"`
	BankName        []FuzzySuggestion `json:"bank_name,omitempty"`
	UserDisplayName []FuzzySuggestion `json:"user_display_name,omitempty"`
}

// BestCashbackRequest представляет запрос на получение лучшего кэшбэка
type BestCashbackRequest struct {
	GroupName string `json:"group_name"`
	Category  string `json:"category"`
	MonthYear string `json:"month_year"`
}

// ListCashbackRequest представляет запрос на получение списка правил
type ListCashbackRequest struct {
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
	UserID string `json:"user_id,omitempty"`
}

// ListCashbackResponse представляет ответ со списком правил
type ListCashbackResponse struct {
	Rules []CashbackRule `json:"rules"`
	Total int            `json:"total"`
	Limit int            `json:"limit"`
	Offset int           `json:"offset"`
}

// ErrorResponse представляет ответ с ошибкой
type ErrorResponse struct {
	Error   string   `json:"error"`
	Details []string `json:"details,omitempty"`
}

