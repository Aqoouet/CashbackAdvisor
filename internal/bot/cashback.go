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

// handleNewCashback обрабатывает добавление нового кэшбэка.
func (b *Bot) handleNewCashback(message *tgbotapi.Message, userID int64) {
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
			"Формат: Банк, Категория, Процент, Сумма[, Месяц]\n" +
			"Пример: \"Тинькофф, Такси, 5%, 3000\" (месяц опционален)"
		b.sendText(message.Chat.ID, text)
		return
	}

	b.continueWithValidation(message, data)
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

	rule, err := b.client.GetBestCashback(groupName, category, monthYear)
	if err != nil {
		if !skipSuggestion {
			b.trySuggestSimilarCategory(message, category, groupName, monthYear)
		} else {
			b.sendText(message.Chat.ID, formatNotFoundMessage(category, monthYear))
		}
		return
	}

	b.sendText(message.Chat.ID, formatBestCashback(rule))
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

	if simPercent > 60.0 {
		b.suggestCategoryCorrection(message, category, similar, simPercent, distance)
		return
	}

	if simPercent > 40.0 && distance <= max(len(category)/2, 4) {
		b.suggestWeakCategoryCorrection(message, category, similar, simPercent, distance)
		return
	}

	log.Printf("❌ Похожесть слишком низкая (%.1f%%), не предлагаю исправление", simPercent)
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

