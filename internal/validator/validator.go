package validator

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// ValidationError представляет ошибку валидации
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors представляет множественные ошибки валидации
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}
	var messages []string
	for _, err := range e {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

func (e ValidationErrors) Strings() []string {
	var messages []string
	for _, err := range e {
		messages = append(messages, err.Error())
	}
	return messages
}

// ValidateMonthYear валидирует формат даты окончания дд.мм.гггг и возвращает дату
func ValidateMonthYear(monthYear string) (time.Time, error) {
	if monthYear == "" {
		return time.Time{}, ValidationError{
			Field:   "month_year",
			Message: "обязательное поле",
		}
	}

	// Пробуем парсить формат дд.мм.гггг
	t, err := time.Parse("02.01.2006", monthYear)
	if err == nil {
		return t, nil
	}

	// Обратная совместимость: пробуем парсить старый формат YYYY-MM
	t, err = time.Parse("2006-01", monthYear)
	if err == nil {
		// Конвертируем в последний день месяца
		lastDay := time.Date(t.Year(), t.Month()+1, 0, 0, 0, 0, 0, time.UTC)
		return lastDay, nil
	}

	return time.Time{}, ValidationError{
		Field:   "month_year",
		Message: fmt.Sprintf("неверный формат даты, ожидается дд.мм.гггг (например, 31.12.2024), получено: %s", monthYear),
	}
}

// ValidateCashbackPercent валидирует процент кэшбэка
func ValidateCashbackPercent(percent float64) error {
	if math.IsNaN(percent) || math.IsInf(percent, 0) {
		return ValidationError{
			Field:   "cashback_percent",
			Message: "недопустимое числовое значение",
		}
	}

	if percent < 0.00 || percent > 100.00 {
		return ValidationError{
			Field:   "cashback_percent",
			Message: fmt.Sprintf("должен быть в диапазоне 0.00 - 100.00, получено: %.2f", percent),
		}
	}

	return nil
}

// ValidateMaxAmount валидирует максимальную сумму
func ValidateMaxAmount(amount float64) error {
	if math.IsNaN(amount) || math.IsInf(amount, 0) {
		return ValidationError{
			Field:   "max_amount",
			Message: "недопустимое числовое значение",
		}
	}

	if amount < 0.00 {
		return ValidationError{
			Field:   "max_amount",
			Message: fmt.Sprintf("должен быть >= 0.00, получено: %.2f", amount),
		}
	}

	return nil
}

// ValidateTextField валидирует текстовые поля
func ValidateTextField(fieldName, value string, required bool) error {
	if required && strings.TrimSpace(value) == "" {
		return ValidationError{
			Field:   fieldName,
			Message: "обязательное поле",
		}
	}

	if len(value) > 500 {
		return ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("максимальная длина 500 символов, получено: %d", len(value)),
		}
	}

	return nil
}

// RoundToTwoDecimals округляет число до двух знаков после запятой
func RoundToTwoDecimals(value float64) float64 {
	return math.Round(value*100) / 100
}

// ValidateSuggestRequest валидирует запрос на предложения
func ValidateSuggestRequest(groupName, category, bankName, userDisplayName, monthYear string, cashbackPercent, maxAmount float64) ValidationErrors {
	var errors ValidationErrors

	// Валидация текстовых полей
	if err := ValidateTextField("group_name", groupName, true); err != nil {
		errors = append(errors, err.(ValidationError))
	}
	if err := ValidateTextField("category", category, true); err != nil {
		errors = append(errors, err.(ValidationError))
	}
	if err := ValidateTextField("bank_name", bankName, true); err != nil {
		errors = append(errors, err.(ValidationError))
	}
	if err := ValidateTextField("user_display_name", userDisplayName, true); err != nil {
		errors = append(errors, err.(ValidationError))
	}

	// Валидация month_year
	if _, err := ValidateMonthYear(monthYear); err != nil {
		errors = append(errors, err.(ValidationError))
	}

	// Валидация cashback_percent
	if err := ValidateCashbackPercent(cashbackPercent); err != nil {
		errors = append(errors, err.(ValidationError))
	}

	// Валидация max_amount
	if err := ValidateMaxAmount(maxAmount); err != nil {
		errors = append(errors, err.(ValidationError))
	}

	return errors
}

// ValidateCreateRequest валидирует запрос на создание правила
func ValidateCreateRequest(groupName, category, bankName, userID, userDisplayName, monthYear string, cashbackPercent, maxAmount float64) ValidationErrors {
	var errors ValidationErrors

	// Валидация текстовых полей
	if err := ValidateTextField("group_name", groupName, true); err != nil {
		errors = append(errors, err.(ValidationError))
	}
	if err := ValidateTextField("category", category, true); err != nil {
		errors = append(errors, err.(ValidationError))
	}
	if err := ValidateTextField("bank_name", bankName, true); err != nil {
		errors = append(errors, err.(ValidationError))
	}
	if err := ValidateTextField("user_id", userID, true); err != nil {
		errors = append(errors, err.(ValidationError))
	}
	if err := ValidateTextField("user_display_name", userDisplayName, true); err != nil {
		errors = append(errors, err.(ValidationError))
	}

	// Валидация month_year
	if _, err := ValidateMonthYear(monthYear); err != nil {
		errors = append(errors, err.(ValidationError))
	}

	// Валидация cashback_percent
	if err := ValidateCashbackPercent(cashbackPercent); err != nil {
		errors = append(errors, err.(ValidationError))
	}

	// Валидация max_amount
	if err := ValidateMaxAmount(maxAmount); err != nil {
		errors = append(errors, err.(ValidationError))
	}

	return errors
}

