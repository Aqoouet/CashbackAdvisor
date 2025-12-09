package bot

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ParsedData содержит распарсенные данные от пользователя
type ParsedData struct {
	GroupName       string
	Category        string
	BankName        string
	MonthYear       string
	CashbackPercent float64
	MaxAmount       float64
}

// ParseMessage пытается извлечь данные из сообщения пользователя
func ParseMessage(text string) (*ParsedData, error) {
	data := &ParsedData{}
	errors := []string{}

	// Паттерны для извлечения данных
	
	// Месяц и год (декабрь, 2024-12, 12/2024, дек 2024, и т.д.)
	monthPattern := regexp.MustCompile(`(?i)(январ[ья]|феврал[ья]|март[а]?|апрел[ья]|ма[йя]|июн[ья]|июл[ья]|август[а]?|сентябр[ья]|октябр[ья]|ноябр[ья]|декабр[ья]|(\d{4})-(\d{2})|(\d{2})/(\d{4})|(\d{2})\.(\d{4}))`)
	if match := monthPattern.FindString(text); match != "" {
		monthYear, err := parseMonth(match)
		if err == nil {
			data.MonthYear = monthYear
		} else {
			errors = append(errors, "не удалось распознать месяц")
		}
	}

	// Процент кэшбэка (5%, 10 процентов, и т.д.)
	percentPattern := regexp.MustCompile(`(\d+\.?\d*)\s*(%|процент|кэшбэк)`)
	if match := percentPattern.FindStringSubmatch(text); len(match) > 1 {
		if percent, err := strconv.ParseFloat(match[1], 64); err == nil {
			data.CashbackPercent = percent
		}
	}

	// Максимальная сумма (3000р, 3000 рублей, 3000 руб, и т.д.)
	amountPattern := regexp.MustCompile(`(\d+\.?\d*)\s*(р|руб|рубл|₽|рублей)`)
	if match := amountPattern.FindStringSubmatch(text); len(match) > 1 {
		if amount, err := strconv.ParseFloat(match[1], 64); err == nil {
			data.MaxAmount = amount
		}
	}

	// Извлекаем известные банки (можно расширить список)
	banks := []string{
		"Тинькофф", "Тинькоф", "Тинков", "tinkoff",
		"Сбер", "Сбербанк", "sber",
		"Альфа", "Альфа-Банк", "alfa",
		"ВТБ", "vtb",
		"Райффайзен", "raiffeisen",
		"Газпромбанк", "gazprom",
		"Открытие", "otkrytie",
	}
	
	textLower := strings.ToLower(text)
	for _, bank := range banks {
		if strings.Contains(textLower, strings.ToLower(bank)) {
			data.BankName = bank
			break
		}
	}

	// Извлекаем известные категории
	categories := []string{
		"Такси", "такси",
		"Рестораны", "ресторан", "кафе",
		"Супермаркеты", "супермаркет", "продукты",
		"Аптеки", "аптека",
		"АЗС", "бензин", "заправка",
		"Кино", "кинотеатр",
		"Транспорт", "транспорт",
		"Развлечения", "развлечения",
	}
	
	for _, cat := range categories {
		if strings.Contains(textLower, strings.ToLower(cat)) {
			data.Category = cat
			break
		}
	}

	// Если не нашли категорию, пробуем извлечь первое существительное
	if data.Category == "" {
		words := strings.Fields(text)
		for _, word := range words {
			// Простая эвристика: берем слова длиннее 3 символов
			if len(word) > 3 && !isNumber(word) {
				data.Category = word
				break
			}
		}
	}

	return data, nil
}

// parseMonth преобразует различные форматы месяца в YYYY-MM
func parseMonth(monthStr string) (string, error) {
	monthStr = strings.ToLower(strings.TrimSpace(monthStr))
	
	// Если уже в формате YYYY-MM
	if matched, _ := regexp.MatchString(`^\d{4}-\d{2}$`, monthStr); matched {
		return monthStr, nil
	}

	// Если в формате MM/YYYY
	if matched, _ := regexp.MatchString(`^\d{2}/\d{4}$`, monthStr); matched {
		parts := strings.Split(monthStr, "/")
		return fmt.Sprintf("%s-%s", parts[1], parts[0]), nil
	}

	// Если в формате MM.YYYY
	if matched, _ := regexp.MatchString(`^\d{2}\.\d{4}$`, monthStr); matched {
		parts := strings.Split(monthStr, ".")
		return fmt.Sprintf("%s-%s", parts[1], parts[0]), nil
	}

	// Названия месяцев
	months := map[string]string{
		"январ": "01", "янв": "01",
		"феврал": "02", "фев": "02",
		"март": "03", "мар": "03",
		"апрел": "04", "апр": "04",
		"май": "05", "ма": "05",
		"июн": "06", "ию": "06",
		"июл": "07",
		"август": "08", "авг": "08",
		"сентябр": "09", "сен": "09",
		"октябр": "10", "окт": "10",
		"ноябр": "11", "ноя": "11",
		"декабр": "12", "дек": "12",
	}

	// Определяем месяц по названию
	for key, month := range months {
		if strings.Contains(monthStr, key) {
			// Берем текущий год
			year := time.Now().Year()
			return fmt.Sprintf("%d-%s", year, month), nil
		}
	}

	return "", fmt.Errorf("не удалось распознать месяц: %s", monthStr)
}

// isNumber проверяет, является ли строка числом
func isNumber(s string) bool {
	_, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return err == nil
}

// ValidateParsedData проверяет, что все необходимые данные заполнены
func ValidateParsedData(data *ParsedData) []string {
	var missing []string

	if data.BankName == "" {
		missing = append(missing, "название банка")
	}
	if data.Category == "" {
		missing = append(missing, "категория")
	}
	if data.MonthYear == "" {
		missing = append(missing, "месяц и год")
	}
	if data.CashbackPercent == 0 {
		missing = append(missing, "процент кэшбэка")
	}
	if data.MaxAmount == 0 {
		missing = append(missing, "максимальная сумма")
	}

	return missing
}

// FormatParsedData форматирует данные для отображения пользователю
func FormatParsedData(data *ParsedData) string {
	return fmt.Sprintf(
		"📋 Распознанные данные:\n\n"+
			"🏦 Банк: %s\n"+
			"📁 Категория: %s\n"+
			"📅 Месяц: %s\n"+
			"💰 Кэшбэк: %.1f%%\n"+
			"💵 Макс. сумма: %.0f₽",
		data.BankName,
		data.Category,
		data.MonthYear,
		data.CashbackPercent,
		data.MaxAmount,
	)
}

