package validator

import (
	"testing"
	"time"
)

func TestValidateMonthYear(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{"Valid date", "2024-12", false},
		{"Valid date January", "2024-01", false},
		{"Invalid month 13", "2024-13", true},
		{"Invalid month 00", "2024-00", true},
		{"Invalid format", "Dec 2024", true},
		{"Invalid format slash", "2024/12", true},
		{"Empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateMonthYear(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateMonthYear() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidateMonthYearReturnsFirstDay(t *testing.T) {
	result, err := ValidateMonthYear("2024-12")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC)
	if !result.Equal(expected) {
		t.Errorf("Expected first day of month, got %v", result)
	}
}

func TestValidateCashbackPercent(t *testing.T) {
	tests := []struct {
		name      string
		input     float64
		wantError bool
	}{
		{"Valid 0", 0.0, false},
		{"Valid 50", 50.0, false},
		{"Valid 100", 100.0, false},
		{"Valid 5.5", 5.5, false},
		{"Invalid negative", -5.0, true},
		{"Invalid over 100", 150.0, true},
		{"Invalid 100.01", 100.01, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCashbackPercent(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateCashbackPercent() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidateMaxAmount(t *testing.T) {
	tests := []struct {
		name      string
		input     float64
		wantError bool
	}{
		{"Valid 0", 0.0, false},
		{"Valid 1000", 1000.0, false},
		{"Valid 3000.50", 3000.50, false},
		{"Invalid negative", -100.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMaxAmount(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateMaxAmount() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidateTextField(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		value     string
		required  bool
		wantError bool
	}{
		{"Valid non-empty", "test_field", "Valid Value", true, false},
		{"Empty required", "test_field", "", true, true},
		{"Empty not required", "test_field", "", false, false},
		{"Whitespace required", "test_field", "   ", true, true},
		{"Too long", "test_field", string(make([]byte, 501)), true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTextField(tt.fieldName, tt.value, tt.required)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateTextField() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestRoundToTwoDecimals(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"Already rounded", 5.50, 5.50},
		{"Round up", 5.556, 5.56},
		{"Round down", 5.554, 5.55},
		{"Integer", 5.0, 5.0},
		{"Three decimals", 123.456, 123.46},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RoundToTwoDecimals(tt.input)
			if result != tt.expected {
				t.Errorf("RoundToTwoDecimals(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidateSuggestRequest(t *testing.T) {
	tests := []struct {
		name            string
		groupName       string
		category        string
		bankName        string
		userDisplayName string
		monthYear       string
		cashbackPercent float64
		maxAmount       float64
		wantErrors      bool
	}{
		{
			"All valid",
			"Транспорт", "Такси", "Тинькофф", "Иван",
			"2024-12", 5.5, 3000.0,
			false,
		},
		{
			"Invalid month_year",
			"Транспорт", "Такси", "Тинькофф", "Иван",
			"2024-13", 5.5, 3000.0,
			true,
		},
		{
			"Invalid cashback_percent",
			"Транспорт", "Такси", "Тинькофф", "Иван",
			"2024-12", 150.0, 3000.0,
			true,
		},
		{
			"Invalid max_amount",
			"Транспорт", "Такси", "Тинькофф", "Иван",
			"2024-12", 5.5, -100.0,
			true,
		},
		{
			"Empty group_name",
			"", "Такси", "Тинькофф", "Иван",
			"2024-12", 5.5, 3000.0,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateSuggestRequest(
				tt.groupName, tt.category, tt.bankName, tt.userDisplayName,
				tt.monthYear, tt.cashbackPercent, tt.maxAmount,
			)
			if (len(errors) > 0) != tt.wantErrors {
				t.Errorf("ValidateSuggestRequest() errors = %v, wantErrors %v", errors, tt.wantErrors)
			}
		})
	}
}

func TestValidationErrorsStrings(t *testing.T) {
	errors := ValidationErrors{
		{Field: "field1", Message: "error1"},
		{Field: "field2", Message: "error2"},
	}

	strings := errors.Strings()
	if len(strings) != 2 {
		t.Errorf("Expected 2 error strings, got %d", len(strings))
	}

	if strings[0] != "field1: error1" {
		t.Errorf("Unexpected error string: %s", strings[0])
	}
}

