package bot

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rymax1e/open-cashback-advisor/internal/models"
)

// handleNewCashback обрабатывает добавление нового кэшбэка (одна или несколько строк).
func (b *Bot) handleNewCashback(message *tgbotapi.Message, userID int64) {
	// Проверяем, многострочное ли сообщение
	lines := strings.Split(message.Text, "\n")
	
	// Фильтруем пустые строки
	var validLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && strings.Contains(line, ",") {
			validLines = append(validLines, line)
		}
	}
	
	// Если несколько строк, обрабатываем каждую
	if len(validLines) > 1 {
		b.handleMultilineCashback(message, validLines)
		return
	}
	
	// Одна строка - стандартная обработка
	data, err := ParseMessage(message.Text)
	if err != nil {
		b.sendText(message.Chat.ID, fmt.Sprintf("❌ Ошибка парсинга: %s", err))
		return
	}

	log.Printf("🔍 Распознано: Bank='%s', Category='%s', Percent=%.1f%%, Amount=%.0f, Month='%s'",
		data.BankName, data.Category, data.CashbackPercent, data.MaxAmount, data.MonthYear)

	// Проверяем опечатки в названии банка
	if correctedBank, found := FindSimilarBank(data.BankName); found && correctedBank != data.BankName {
		log.Printf("💡 Исправление банка: '%s' → '%s'", data.BankName, correctedBank)
		b.suggestBankCorrection(message, data, correctedBank)
		return
	}

	// Проверяем полноту данных
	missing := ValidateParsedData(data)
	if len(missing) > 0 {
		text := "⚠️ Не хватает данных:\n" + strings.Join(missing, ", ") + "\n\n" +
			"Формат: Банк, Категория, Процент, Сумма[, Дата окончания]\n" +
			"Пример: \"Тинькофф, Такси, 5%, 3000\""
		b.sendText(message.Chat.ID, text)
		return
	}

	b.continueWithValidation(message, data)
}

// handleMultilineCashback обрабатывает добавление нескольких кэшбэков за раз.
func (b *Bot) handleMultilineCashback(message *tgbotapi.Message, lines []string) {
	b.sendText(message.Chat.ID, fmt.Sprintf("📝 Обрабатываю %d строк...\n", len(lines)))
	
	var results []string
	successCount := 0
	errorCount := 0
	
	for i, line := range lines {
		// Парсим строку
		data, err := ParseMessage(line)
		if err != nil {
			results = append(results, fmt.Sprintf("❌ Строка %d: %s", i+1, err))
			errorCount++
			continue
		}
		
		// Проверяем полноту данных
		missing := ValidateParsedData(data)
		if len(missing) > 0 {
			results = append(results, fmt.Sprintf("❌ Строка %d: не хватает %s", i+1, strings.Join(missing, ", ")))
			errorCount++
			continue
		}
		
		// Проверяем опечатки в банке (автоматическая коррекция)
		if correctedBank, found := FindSimilarBank(data.BankName); found && correctedBank != data.BankName {
			log.Printf("💡 Автокоррекция банка: '%s' → '%s'", data.BankName, correctedBank)
			data.BankName = correctedBank
		}
		
		// Сохраняем без валидации через API (упрощенный режим)
		userIDStr := strconv.FormatInt(message.From.ID, 10)
		groupName := b.getUserGroup(message.From.ID)
		
		req := &models.CreateCashbackRequest{
			GroupName:       groupName,
			Category:        data.Category,
			BankName:        data.BankName,
			UserID:          userIDStr,
			UserDisplayName: getUserDisplayName(message.From),
			MonthYear:       data.MonthYear,
			CashbackPercent: data.CashbackPercent,
			MaxAmount:       data.MaxAmount,
			Force:           true,
		}
		
		rule, err := b.client.CreateCashback(req)
		if err != nil {
			results = append(results, fmt.Sprintf("❌ Строка %d: %s", i+1, err))
			errorCount++
		} else {
			results = append(results, fmt.Sprintf("✅ Строка %d: %s - %s (ID: %d)", 
				i+1, rule.BankName, rule.Category, rule.ID))
			successCount++
		}
	}
	
	// Формируем итоговое сообщение
	summary := fmt.Sprintf("📊 Результаты:\n✅ Успешно: %d\n❌ Ошибки: %d\n\n", successCount, errorCount)
	b.sendText(message.Chat.ID, summary+strings.Join(results, "\n"))
	
	b.clearState(message.From.ID)
}

// suggestBankCorrection предлагает исправление названия банка.
func (b *Bot) suggestBankCorrection(message *tgbotapi.Message, data *ParsedData, correctedBank string) {
	text := fmt.Sprintf(
		"💡 Возможная опечатка в названии банка:\n\n"+
			"Вы написали: \"%s\"\n"+
			"Предлагаю исправить на: \"%s\"\n\n"+
			"❓ Исправить?",
		data.BankName, correctedBank,
	)

	correctedData := *data
	correctedData.BankName = correctedBank

	b.setState(message.From.ID, StateAwaitingBankCorrection, &correctedData, nil, 0)
	b.sendWithButtons(message.Chat.ID, text, ButtonsConfirmSimple)
}

// continueWithValidation продолжает валидацию через API.
func (b *Bot) continueWithValidation(message *tgbotapi.Message, data *ParsedData) {
	userID := message.From.ID
	groupName := b.getUserGroup(message.From.ID)

	b.sendText(message.Chat.ID, FormatParsedData(data))
	b.sendText(message.Chat.ID, "🔍 Проверяю данные...")

	suggestReq := &models.SuggestRequest{
		GroupName:       groupName,
		Category:        data.Category,
		BankName:        data.BankName,
		UserDisplayName: getUserDisplayName(message.From),
		MonthYear:       data.MonthYear,
		CashbackPercent: data.CashbackPercent,
		MaxAmount:       data.MaxAmount,
	}

	suggestion, err := b.client.Suggest(suggestReq)
	if err != nil {
		b.sendText(message.Chat.ID, fmt.Sprintf("❌ Ошибка проверки: %s", err))
		b.clearState(userID)
		return
	}

	logSuggestions(suggestion, data)

	if !suggestion.Valid {
		text := "❌ Ошибки валидации:\n" + strings.Join(suggestion.Errors, "\n")
		b.sendText(message.Chat.ID, text)
		b.clearState(userID)
		return
	}

	// Проверяем реальные отличия в предложениях
	realSuggestions := b.findRealSuggestions(data, suggestion)

	if len(realSuggestions) > 0 {
		text := "💡 Возможно, вы имели в виду:\n\n"
		text += strings.Join(realSuggestions, "\n")
		text += "\n\n❓ Исправить и сохранить?"

		b.setState(userID, StateAwaitingConfirmation, data, suggestion, 0)
		b.sendWithButtons(message.Chat.ID, text, ButtonsConfirm)
	} else {
		b.saveCashback(message.Chat.ID, message.From, data, false)
		b.clearState(userID)
	}
}

// findRealSuggestions находит реальные отличия между введёнными данными и предложениями.
func (b *Bot) findRealSuggestions(data *ParsedData, suggestion *models.SuggestResponse) []string {
	var realSuggestions []string

	if len(suggestion.Suggestions.BankName) > 0 {
		suggestedBank := strings.TrimSpace(suggestion.Suggestions.BankName[0].Value)
		originalBank := strings.TrimSpace(data.BankName)

		if originalBank != suggestedBank {
			realSuggestions = append(realSuggestions,
				fmt.Sprintf("🏦 Банк: %s → %s", originalBank, suggestedBank))
		}
	}

	if len(suggestion.Suggestions.Category) > 0 {
		suggestedCategory := strings.TrimSpace(suggestion.Suggestions.Category[0].Value)
		originalCategory := strings.TrimSpace(data.Category)

		if originalCategory != suggestedCategory {
			realSuggestions = append(realSuggestions,
				fmt.Sprintf("📁 Категория: %s → %s", originalCategory, suggestedCategory))
		}
	}

	return realSuggestions
}

// saveCashback сохраняет кэшбэк через API.
func (b *Bot) saveCashback(chatID int64, user *tgbotapi.User, data *ParsedData, force bool) {
	userIDStr := strconv.FormatInt(user.ID, 10)
	groupName := b.getUserGroup(user.ID)

	req := &models.CreateCashbackRequest{
		GroupName:       groupName,
		Category:        data.Category,
		BankName:        data.BankName,
		UserID:          userIDStr,
		UserDisplayName: getUserDisplayName(user),
		MonthYear:       data.MonthYear,
		CashbackPercent: data.CashbackPercent,
		MaxAmount:       data.MaxAmount,
		Force:           force,
	}

	log.Printf("💾 Сохранение в API: Bank='%s', Category='%s', Force=%v",
		req.BankName, req.Category, force)

	rule, err := b.client.CreateCashback(req)
	if err != nil {
		b.sendText(chatID, fmt.Sprintf("❌ Ошибка сохранения: %s", err))
		return
	}

	b.sendText(chatID, formatSavedCashback(rule))
}

// handleBestQueryByCategory обрабатывает поиск лучшего кэшбэка по категории.
func (b *Bot) handleBestQueryByCategory(message *tgbotapi.Message) {
	b.handleBestQueryWithCorrection(message, normalizeString(message.Text), false)
}

// handleBestQueryWithCorrection выполняет поиск с возможностью исправления.
func (b *Bot) handleBestQueryWithCorrection(message *tgbotapi.Message, category string, skipSuggestion bool) {
	if category == "" {
		b.sendText(message.Chat.ID, "❌ Укажите категорию. Например: \"Такси\"")
		return
	}

	userIDStr := strconv.FormatInt(message.From.ID, 10)
	groupName, err := b.client.GetUserGroup(userIDStr)
	if err != nil {
		b.sendText(message.Chat.ID, "❌ Вы должны быть в группе. Используйте /creategroup или /joingroup")
		return
	}

	now := time.Now()
	monthYear := fmt.Sprintf("%d-%02d", now.Year(), now.Month())

	b.sendText(message.Chat.ID, fmt.Sprintf("🔍 Ищу лучший кэшбэк для \"%s\" в группе \"%s\"...", category, groupName))

	// Получаем все кэшбэки по точной категории
	allRules, err := b.getAllCashbacksByCategory(groupName, category, monthYear)
	
	// Если нашли точные совпадения - показываем все
	if err == nil && len(allRules) > 0 {
		log.Printf("✅ Найдено %d активных кешбеков для категории '%s'", len(allRules), category)
		b.sendText(message.Chat.ID, formatAllCashbackResults(allRules, category, false))
		return
	}
	
	// Не нашли точную категорию - пробуем найти похожие (если не пропускаем)
	log.Printf("⚠️ Не найдено активных кешбеков для '%s', ищу похожие категории", category)
	if !skipSuggestion {
		b.trySuggestSimilarCategory(message, category, groupName, monthYear)
	} else {
		// skipSuggestion=true означает, что уже была попытка с исправлением
		// Пробуем "Все покупки" как последний вариант
		log.Printf("⚠️ Уже была попытка исправления, пробуем 'Все покупки'")
		allPurchasesRules, errAll := b.getAllCashbacksByCategory(groupName, "Все покупки", monthYear)
		if errAll == nil && len(allPurchasesRules) > 0 {
			log.Printf("✅ Найдено %d кешбеков для 'Все покупки' как fallback", len(allPurchasesRules))
			b.sendText(message.Chat.ID, formatAllCashbackResults(allPurchasesRules, category, true))
			return
		}
		log.Printf("❌ 'Все покупки' тоже не найдены, показываю 'не найдено'")
		b.sendText(message.Chat.ID, formatNotFoundMessage(category, monthYear))
	}
}

// trySuggestSimilarCategory пытается найти похожую категорию.
func (b *Bot) trySuggestSimilarCategory(message *tgbotapi.Message, category, groupName, monthYear string) {
	categories, err := b.client.ListAllCategories(groupName, monthYear)
	log.Printf("🔍 Получено категорий из API: %d, ошибка: %v", len(categories), err)

	if err != nil || len(categories) == 0 {
		b.sendText(message.Chat.ID, formatNotFoundMessage(category, monthYear))
		return
	}

	similar, simPercent, distance := findSimilarCategory(category, categories)
	log.Printf("🔍 Сравнение: '%s' → '%s' (расстояние: %d, похожесть: %.1f%%)",
		category, similar, distance, simPercent)

	// Если это точное совпадение (100%), значит категория существует, но все кешбеки истекли
	// В этом случае НЕ выполняем поиск снова (чтобы избежать бесконечного цикла)
	// Вместо этого сразу пробуем fallback на "Все покупки"
	if simPercent == 100.0 && strings.EqualFold(category, similar) {
		log.Printf("⚠️ Категория '%s' существует, но все кешбеки истекли. Пропускаю повторный поиск.", category)
		// Логика fallback уже выполнена в handleBestQueryWithCorrection, просто показываем "не найдено"
		b.sendText(message.Chat.ID, formatNotFoundMessage(category, monthYear))
		return
	}

	if simPercent > 60.0 {
		b.suggestCategoryCorrection(message, category, similar, simPercent, distance)
		return
	}

	if simPercent > 40.0 && distance <= max(len(category)/2, 4) {
		b.suggestWeakCategoryCorrection(message, category, similar, simPercent, distance)
		return
	}

	// Ничего похожего не нашли - пробуем "Все покупки" как fallback
	log.Printf("❌ Похожесть слишком низкая (%.1f%%), пробую 'Все покупки'", simPercent)
	allPurchasesRules, errAll := b.getAllCashbacksByCategory(groupName, "Все покупки", monthYear)
	if errAll == nil && len(allPurchasesRules) > 0 {
		b.sendText(message.Chat.ID, formatAllCashbackResults(allPurchasesRules, category, true))
		return
	}
	
	// Даже "Все покупки" не найдены - показываем "не найдено"
	log.Printf("❌ 'Все покупки' тоже не найдены")
	b.sendText(message.Chat.ID, formatNotFoundMessage(category, monthYear))
}

// suggestCategoryCorrection предлагает уверенное исправление категории.
func (b *Bot) suggestCategoryCorrection(message *tgbotapi.Message, original, suggested string, simPercent float64, distance int) {
	text := fmt.Sprintf(
		"❌ Категория не найдена\n\n"+
			"📁 Вы написали: \"%s\"\n"+
			"💡 Возможно, вы имели в виду: \"%s\"\n\n"+
			"❓ Искать с исправленным названием?",
		original, suggested,
	)

	log.Printf("✅ Предлагаю исправление: '%s' → '%s' (расстояние: %d, похожесть: %.1f%%)",
		original, suggested, distance, simPercent)

	b.setState(message.From.ID, StateAwaitingCategoryCorrection, &ParsedData{Category: suggested}, nil, 0)
	b.sendWithButtons(message.Chat.ID, text, ButtonsConfirmSimple)
}

// suggestWeakCategoryCorrection предлагает слабое исправление категории.
func (b *Bot) suggestWeakCategoryCorrection(message *tgbotapi.Message, original, suggested string, simPercent float64, distance int) {
	text := fmt.Sprintf(
		"❌ Категория не найдена\n\n"+
			"📁 Вы написали: \"%s\"\n"+
			"💡 Может быть: \"%s\"?\n\n"+
			"❓ Попробовать с этим вариантом?",
		original, suggested,
	)

	log.Printf("⚠️ Слабое предположение: '%s' → '%s' (расстояние: %d, похожесть: %.1f%%)",
		original, suggested, distance, simPercent)

	b.setState(message.From.ID, StateAwaitingCategoryCorrection, &ParsedData{Category: suggested}, nil, 0)
	b.sendWithButtons(message.Chat.ID, text, ButtonsConfirmSimple)
}

// logSuggestions логирует полученные предложения.
func logSuggestions(suggestion *models.SuggestResponse, data *ParsedData) {
	log.Printf("💡 Получены предложения от API: Valid=%v, BankSuggestions=%d, CategorySuggestions=%d",
		suggestion.Valid, len(suggestion.Suggestions.BankName), len(suggestion.Suggestions.Category))

	if len(suggestion.Suggestions.BankName) > 0 {
		log.Printf("   Предложение банка: '%s' (было: '%s')",
			suggestion.Suggestions.BankName[0].Value, data.BankName)
	}
	if len(suggestion.Suggestions.Category) > 0 {
		log.Printf("   Предложение категории: '%s' (было: '%s')",
			suggestion.Suggestions.Category[0].Value, data.Category)
	}
}

// getAllCashbacksByCategory получает все кэшбэки по категории через список всех кэшбэков группы.
// Ищет все категории, которые содержат введенное слово (без учета регистра).
func (b *Bot) getAllCashbacksByCategory(groupName, category, monthYear string) ([]models.CashbackRule, error) {
	// Получаем все кэшбэки группы
	list, err := b.client.ListCashback(groupName, 1000, 0)
	if err != nil {
		return nil, err
	}
	
	// Нормализуем введенную категорию для поиска
	categoryLower := strings.ToLower(strings.TrimSpace(category))
	
	// Фильтруем по категории (поиск по подстроке) и дате
	var filtered []models.CashbackRule
	var matchedButExpired int
	now := time.Now()
	
	for _, rule := range list.Rules {
		// Проверяем, что категория содержит введенное слово (без учета регистра)
		ruleCategoryLower := strings.ToLower(rule.Category)
		containsCategory := strings.Contains(ruleCategoryLower, categoryLower)
		
		// Также проверяем точное совпадение (приоритет)
		exactMatch := strings.EqualFold(rule.Category, category)
		
		if exactMatch || containsCategory {
			if rule.MonthYear.After(now.AddDate(0, 0, -1)) {
				filtered = append(filtered, rule)
			} else {
				matchedButExpired++
				log.Printf("⏰ Найден кешбек для '%s' но он истек: %s (срок до %s)", 
					category, rule.BankName, rule.MonthYear.Format("02.01.2006"))
			}
		}
	}
	
	if len(filtered) == 0 {
		if matchedButExpired > 0 {
			log.Printf("❌ Для '%s' найдено %d кешбеков, но все истекли", category, matchedButExpired)
		}
		return nil, fmt.Errorf("кэшбэк не найден")
	}
	
	// Сортируем: сначала точные совпадения, потом по убыванию процента
	sortCashbackByCategoryAndPercent(filtered, category)
	
	return filtered, nil
}

// sortCashbackByCategoryAndPercent сортирует кэшбэки: сначала точные совпадения категории, потом по убыванию процента.
func sortCashbackByCategoryAndPercent(rules []models.CashbackRule, searchCategory string) {
	for i := 0; i < len(rules)-1; i++ {
		for j := i + 1; j < len(rules); j++ {
			// Приоритет точным совпадениям
			iExact := strings.EqualFold(rules[i].Category, searchCategory)
			jExact := strings.EqualFold(rules[j].Category, searchCategory)
			
			shouldSwap := false
			
			if iExact && !jExact {
				// i - точное совпадение, j - нет, не меняем
				shouldSwap = false
			} else if !iExact && jExact {
				// j - точное совпадение, i - нет, меняем
				shouldSwap = true
			} else {
				// Оба одинаковые по типу совпадения, сортируем по проценту
				if rules[j].CashbackPercent > rules[i].CashbackPercent ||
					(rules[j].CashbackPercent == rules[i].CashbackPercent && rules[j].MaxAmount > rules[i].MaxAmount) {
					shouldSwap = true
				}
			}
			
			if shouldSwap {
				rules[i], rules[j] = rules[j], rules[i]
			}
		}
	}
}

// trySuggestSimilarBank пытается найти похожий банк.
func (b *Bot) trySuggestSimilarBank(message *tgbotapi.Message, bankName, groupName string) {
	banks, err := b.client.ListAllBanks(groupName)
	log.Printf("🔍 Получено банков из API: %d, ошибка: %v", len(banks), err)

	if err != nil || len(banks) == 0 {
		b.sendText(message.Chat.ID, fmt.Sprintf("❌ Кешбек для банка \"%s\" не найден в вашей группе.\n\nИспользуйте /banklist для просмотра всех банков.", bankName))
		return
	}

	similar, simPercent, distance := findSimilarCategory(bankName, banks)
	log.Printf("🔍 Сравнение банков: '%s' → '%s' (расстояние: %d, похожесть: %.1f%%)", bankName, similar, distance, simPercent)

	// Если есть хорошее совпадение (>60% или >40% с расстоянием ≤2)
	if simPercent > 60.0 || (simPercent > 40.0 && distance <= 2) {
		log.Printf("✅ Предлагаю исправление банка: '%s' → '%s' (расстояние: %d, похожесть: %.1f%%)", bankName, similar, distance, simPercent)
		text := fmt.Sprintf("🤔 Возможно, вы имели в виду банк \"%s\"?", similar)
		buttons := [][]tgbotapi.KeyboardButton{
			{
				tgbotapi.NewKeyboardButton(fmt.Sprintf("✅ Да, показать для \"%s\"", similar)),
			},
			{
				tgbotapi.NewKeyboardButton("✏️ Изменить вручную"),
				tgbotapi.NewKeyboardButton("🚫 Отменить"),
			},
		}

		reply := tgbotapi.NewMessage(message.Chat.ID, text)
		reply.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			Keyboard:        buttons,
			ResizeKeyboard:  true,
			OneTimeKeyboard: true,
		}

		if _, err := b.api.Send(reply); err != nil {
			log.Printf("❌ Ошибка отправки: %v", err)
		}

		// Сохраняем имя банка и название группы в ParsedData
		b.setState(message.From.ID, StateAwaitingBankCorrection, &ParsedData{BankName: similar, Category: groupName}, nil, 0)
		return
	}

	// Похожих банков нет
	log.Printf("⚠️ Похожие банки не найдены (simPercent: %.1f%%)", simPercent)
	b.sendText(message.Chat.ID, fmt.Sprintf("❌ Кешбек для банка \"%s\" не найден в вашей группе.\n\nИспользуйте /banklist для просмотра всех банков.", bankName))
}

