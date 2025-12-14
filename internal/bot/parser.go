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
// Поддерживает два формата:
// 1. Через запятую: "Банк, Категория, Процент, Сумма, Месяц"
// 2. Свободный текст (старый формат)
func ParseMessage(text string) (*ParsedData, error) {
	// Проверяем, есть ли запятые - значит используется новый формат
	if strings.Contains(text, ",") {
		return parseCommaSeparated(text)
	}
	
	// Старый формат - парсим свободный текст
	return parseFreeText(text)
}

// parseCommaSeparated парсит данные в формате: "Банк, Категория, Процент, Сумма[, Месяц]"
// Месяц опционален - если не указан, используется текущий
func parseCommaSeparated(text string) (*ParsedData, error) {
	parts := strings.Split(text, ",")
	if len(parts) < 4 {
		return nil, fmt.Errorf("неверный формат. Используйте: Банк, Категория, Процент, Сумма[, Месяц]")
	}
	
	data := &ParsedData{
		GroupName: "Общие",
	}
	
	// 1. Банк (автоматическая нормализация)
	data.BankName = normalizeString(parts[0])
	
	// 2. Категория (автоматическая нормализация)
	data.Category = normalizeString(parts[1])
	
	// 3. Процент
	percentStr := strings.TrimSpace(parts[2])
	percentStr = strings.ReplaceAll(percentStr, "%", "")
	percentStr = strings.TrimSpace(percentStr)
	if percent, err := strconv.ParseFloat(percentStr, 64); err == nil {
		data.CashbackPercent = percent
	} else {
		return nil, fmt.Errorf("неверный формат процента: %s", parts[2])
	}
	
	// 4. Сумма
	amountStr := strings.TrimSpace(parts[3])
	amountStr = strings.ReplaceAll(amountStr, "р", "")
	amountStr = strings.ReplaceAll(amountStr, "₽", "")
	amountStr = strings.ReplaceAll(amountStr, " ", "")
	if amount, err := strconv.ParseFloat(amountStr, 64); err == nil {
		data.MaxAmount = amount
	} else {
		return nil, fmt.Errorf("неверный формат суммы: %s", parts[3])
	}
	
	// 5. Месяц (опционален)
	if len(parts) >= 5 && strings.TrimSpace(parts[4]) != "" {
		monthStr := strings.TrimSpace(parts[4])
		if monthYear, err := parseMonth(monthStr); err == nil {
			data.MonthYear = monthYear
		} else {
			return nil, fmt.Errorf("неверный формат месяца: %s", parts[4])
		}
	} else {
		// Используем текущий месяц по умолчанию
		now := time.Now()
		data.MonthYear = fmt.Sprintf("%d-%02d", now.Year(), now.Month())
	}
	
	return data, nil
}

// parseFreeText парсит данные из свободного текста (старый формат)
func parseFreeText(text string) (*ParsedData, error) {
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
			data.BankName = normalizeString(bank)
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
			data.Category = normalizeString(cat)
			break
		}
	}

	// Если не нашли категорию, пробуем извлечь из текста
	if data.Category == "" {
		words := strings.Fields(text)
		var categoryWords []string
		
		for _, word := range words {
			// Пропускаем банк, числа, процент, рубли, месяцы
			wordLower := strings.ToLower(word)
			
			// Проверяем, не является ли слово названием банка
			isBankName := false
			if data.BankName != "" {
				isBankName = strings.Contains(wordLower, strings.ToLower(data.BankName)) ||
							 strings.Contains(strings.ToLower(data.BankName), wordLower)
			}
			
			if len(word) > 2 && !isNumber(word) && 
			   !isBankName &&
			   !strings.Contains(wordLower, "%") &&
			   !strings.Contains(wordLower, "руб") &&
			   !strings.HasSuffix(wordLower, "р") &&
			   !strings.Contains(wordLower, "январ") &&
			   !strings.Contains(wordLower, "феврал") &&
			   !strings.Contains(wordLower, "март") &&
			   !strings.Contains(wordLower, "апрел") &&
			   !strings.Contains(wordLower, "ма") &&
			   !strings.Contains(wordLower, "июн") &&
			   !strings.Contains(wordLower, "июл") &&
			   !strings.Contains(wordLower, "август") &&
			   !strings.Contains(wordLower, "сентябр") &&
			   !strings.Contains(wordLower, "октябр") &&
			   !strings.Contains(wordLower, "ноябр") &&
			   !strings.Contains(wordLower, "декабр") {
				categoryWords = append(categoryWords, word)
				// Берем до 3 слов для категории
				if len(categoryWords) >= 3 {
					break
				}
			}
		}
		
		if len(categoryWords) > 0 {
			data.Category = normalizeString(strings.Join(categoryWords, " "))
		}
	}

	return data, nil
}

// normalizeString нормализует строку: убирает лишние пробелы по краям и между словами
func normalizeString(s string) string {
	// Убираем пробелы по краям
	s = strings.TrimSpace(s)
	
	// Убираем множественные пробелы между словами
	words := strings.Fields(s)
	return strings.Join(words, " ")
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

