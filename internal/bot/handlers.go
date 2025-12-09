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

// Bot представляет Telegram бота
type Bot struct {
	api       *tgbotapi.BotAPI
	client    *APIClient
	userStates map[int64]*UserState
}

// UserState хранит состояние пользователя
type UserState struct {
	State      string
	Data       *ParsedData
	Suggestion *models.SuggestResponse
	RuleID     int64 // Для обновления/удаления
}

// NewBot создает нового бота
func NewBot(token string, apiClient *APIClient, debug bool) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать бота: %w", err)
	}

	api.Debug = debug
	log.Printf("✅ Авторизован как @%s", api.Self.UserName)

	return &Bot{
		api:        api,
		client:     apiClient,
		userStates: make(map[int64]*UserState),
	}, nil
}

// Start запускает бота
func (b *Bot) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	log.Println("🤖 Бот запущен и ожидает сообщений...")

	for update := range updates {
		if update.Message != nil {
			b.handleMessage(update.Message)
		} else if update.CallbackQuery != nil {
			b.handleCallback(update.CallbackQuery)
		}
	}
}

// handleMessage обрабатывает текстовые сообщения
func (b *Bot) handleMessage(message *tgbotapi.Message) {
	userID := message.From.ID
	
	log.Printf("📨 Сообщение от @%s: %s", message.From.UserName, message.Text)

	// Обработка команд
	if message.IsCommand() {
		switch message.Command() {
		case "start":
			b.handleStart(message)
		case "help":
			b.handleHelp(message)
		case "add":
			b.handleAddCommand(message)
		case "list":
			b.handleList(message)
		case "best":
			b.handleBestCommand(message)
		case "update":
			b.handleUpdateCommand(message)
		case "delete":
			b.handleDeleteCommand(message)
		case "cancel":
			b.handleCancel(message)
		default:
			b.sendMessage(message.Chat.ID, "❌ Неизвестная команда. Используйте /help для справки.")
		}
		return
	}

	// Обработка состояний пользователя
	state, exists := b.userStates[userID]
	if exists {
		switch state.State {
		case "awaiting_confirmation":
			b.handleConfirmation(message, state)
			return
		case "awaiting_bank_correction":
			b.handleBankCorrection(message, state)
			return
		case "awaiting_update_data":
			b.handleUpdateData(message, state)
			return
		case "awaiting_delete_confirmation":
			b.handleDeleteConfirmation(message, state)
			return
		}
	}

	// Если нет запятой - это поиск лучшего кэшбэка по категории
	if !strings.Contains(message.Text, ",") {
		b.handleBestQueryByCategory(message)
		return
	}

	// Есть запятая - это добавление правила
	b.handleNewRule(message)
}

// handleStart обрабатывает команду /start
func (b *Bot) handleStart(message *tgbotapi.Message) {
	text := fmt.Sprintf("👋 Привет! Я помогаю не упустить выгодный кэшбэк.\n\n"+
		"🎯 Что я умею:\n"+
		"• Запоминаю условия кэшбэка от разных банков\n"+
		"• Подсказываю, где сейчас самый выгодный кэшбэк\n"+
		"• Исправляю опечатки в названиях банков и категорий\n"+
		"• Показываю все твои сохранённые правила\n\n"+
		"✍️ Добавить правило (месяц опционален):\n"+
		"\"Тинькофф, Такси, 5%%, 3000\"\n"+
		"\"Сбер, Перекресток доставка, 12%%, 30000, январь\"\n\n"+
		"🔍 Найти лучший кэшбэк (без запятых):\n"+
		"\"Такси\" - покажет для текущего месяца\n"+
		"\"Перекресток доставка\"\n\n"+
		"📋 Команды:\n"+
		"/list - все правила (от всех пользователей)\n"+
		"/best - найти лучший кэшбэк\n"+
		"/update ID - обновить своё правило\n"+
		"/delete ID - удалить своё правило\n"+
		"/help - подробная справка\n\n"+
		"Я пойму, проверю и сохраню! 😊\n\n"+
		"ℹ️ Версия: %s", BuildInfo())

	b.sendMessage(message.Chat.ID, text)
}

// handleHelp обрабатывает команду /help
func (b *Bot) handleHelp(message *tgbotapi.Message) {
	text := fmt.Sprintf("📖 Подробная справка (Версия: %s)\n\n"+
		"🔹 /add - Добавить новое правило кэшбэка\n"+
		"🔹 /list - Показать все правила (от всех пользователей)\n"+
		"🔹 /best - Найти лучший кэшбэк среди всех\n"+
		"🔹 /update ID - Обновить своё правило\n"+
		"🔹 /delete ID - Удалить своё правило\n"+
		"🔹 /cancel - Отменить текущую операцию\n\n"+
		"💡 Формат добавления правил (с запятыми):\n"+
		"Банк, Категория, Процент, Сумма[, Месяц]\n\n"+
		"📝 Примеры добавления:\n"+
		"• \"Тинькофф, Такси, 5%%, 3000\" (месяц = текущий)\n"+
		"• \"Сбер, Супермаркеты, 10, 5000, январь\"\n"+
		"• \"Альфа, Рестораны, 7.5, 4000\"\n"+
		"• \"ВТБ, Перекресток доставка, 12, 30000, март\"\n\n"+
		"🔍 Поиск лучшего кэшбэка (БЕЗ запятых):\n"+
		"Бот найдёт лучшее предложение среди ВСЕХ пользователей!\n"+
		"• \"Такси\" (покажет для текущего месяца)\n"+
		"• \"Перекресток доставка\"\n"+
		"• \"Рестораны\"\n\n"+
		"💡 Идея: Делимся информацией о кэшбэке - помогаем друг другу!\n\n"+
		"✨ Бот умеет исправлять опечатки! 😊", BuildInfo())

	b.sendMessage(message.Chat.ID, text)
}

// handleNewRule обрабатывает новое правило от пользователя
func (b *Bot) handleNewRule(message *tgbotapi.Message) {
	userID := message.From.ID
	
	// Парсим сообщение
	data, err := ParseMessage(message.Text)
	if err != nil {
		b.sendMessage(message.Chat.ID, fmt.Sprintf("❌ Ошибка парсинга: %s", err))
		return
	}

	// Логируем распознанные данные для отладки
	log.Printf("🔍 Распознано: Bank='%s', Category='%s', Percent=%.1f%%, Amount=%.0f, Month='%s'",
		data.BankName, data.Category, data.CashbackPercent, data.MaxAmount, data.MonthYear)

	// Проверяем опечатки в названии банка
	if correctedBank, found := FindSimilarBank(data.BankName); found && correctedBank != data.BankName {
		log.Printf("💡 Исправление банка: '%s' → '%s'", data.BankName, correctedBank)
		
		// Показываем пользователю предложение исправления
		text := fmt.Sprintf("💡 Возможная опечатка в названии банка:\n\n"+
			"Вы написали: \"%s\"\n"+
			"Предлагаю исправить на: \"%s\"\n\n"+
			"❓ Исправить?", data.BankName, correctedBank)
		
		// Сохраняем состояние для подтверждения с исправленным банком
		correctedData := *data
		correctedData.BankName = correctedBank
		
		b.userStates[userID] = &UserState{
			State: "awaiting_bank_correction",
			Data:  &correctedData,
		}
		
		// Отправляем с кнопками
		b.sendMessageWithButtons(message.Chat.ID, text, [][]string{
			{"✅ Да, исправить", "❌ Нет, оставить как есть"},
		})
		return
	}

	// Проверяем, что все данные есть
	missing := ValidateParsedData(data)
	if len(missing) > 0 {
		text := "⚠️ Не хватает данных:\n" + strings.Join(missing, ", ") + "\n\n" +
			"Формат: Банк, Категория, Процент, Сумма[, Месяц]\n" +
			"Пример: \"Тинькофф, Такси, 5%, 3000\" (месяц опционален)"
		b.sendMessage(message.Chat.ID, text)
		return
	}

	// Показываем распознанные данные
	b.sendMessage(message.Chat.ID, FormatParsedData(data))
	b.sendMessage(message.Chat.ID, "🔍 Проверяю данные...")

	// Вызываем /suggest для проверки
	suggestReq := &models.SuggestRequest{
		GroupName:       "Общие", // Можно сделать настраиваемым
		Category:        data.Category,
		BankName:        data.BankName,
		UserDisplayName: getUserDisplayName(message.From),
		MonthYear:       data.MonthYear,
		CashbackPercent: data.CashbackPercent,
		MaxAmount:       data.MaxAmount,
	}

	suggestion, err := b.client.Suggest(suggestReq)
	if err != nil {
		b.sendMessage(message.Chat.ID, fmt.Sprintf("❌ Ошибка проверки: %s", err))
		return
	}

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

	// Если есть ошибки валидации
	if !suggestion.Valid {
		text := "❌ Ошибки валидации:\n" + strings.Join(suggestion.Errors, "\n")
		b.sendMessage(message.Chat.ID, text)
		return
	}

	// Если есть предложения по исправлению
	// Фильтруем только те предложения, которые реально отличаются
	var realSuggestions []string
	hasRealSuggestions := false
	
	if len(suggestion.Suggestions.BankName) > 0 {
		suggestedBank := suggestion.Suggestions.BankName[0].Value
		originalBank := strings.TrimSpace(data.BankName)
		suggestedBankTrimmed := strings.TrimSpace(suggestedBank)
		
		// Сравниваем точно (с учетом пробелов внутри), но без учета регистра и лишних пробелов по краям
		if originalBank != suggestedBankTrimmed {
			realSuggestions = append(realSuggestions, fmt.Sprintf("🏦 Банк: %s → %s",
				originalBank, suggestedBankTrimmed))
			hasRealSuggestions = true
		}
	}
	
	if len(suggestion.Suggestions.Category) > 0 {
		suggestedCategory := suggestion.Suggestions.Category[0].Value
		originalCategory := strings.TrimSpace(data.Category)
		suggestedCategoryTrimmed := strings.TrimSpace(suggestedCategory)
		
		// Сравниваем точно (с учетом пробелов внутри)
		if originalCategory != suggestedCategoryTrimmed {
			realSuggestions = append(realSuggestions, fmt.Sprintf("📁 Категория: %s → %s",
				originalCategory, suggestedCategoryTrimmed))
			hasRealSuggestions = true
		}
	}

	if hasRealSuggestions {
		text := "💡 Возможно, вы имели в виду:\n\n"
		text += strings.Join(realSuggestions, "\n")
		text += "\n\n❓ Исправить и сохранить?"
		
		// Сохраняем состояние для подтверждения
		b.userStates[userID] = &UserState{
			State:      "awaiting_confirmation",
			Data:       data,
			Suggestion: suggestion,
		}
		
		// Отправляем с кнопками
		b.sendMessageWithButtons(message.Chat.ID, text, [][]string{
			{"✅ Да, исправить", "❌ Нет, оставить как есть"},
			{"🚫 Отмена"},
		})
	} else {
		// Нет предложений - сразу сохраняем
		b.saveRule(message.Chat.ID, message.From, data, false)
	}
}

// handleBankCorrection обрабатывает подтверждение исправления банка
func (b *Bot) handleBankCorrection(message *tgbotapi.Message, state *UserState) {
	text := strings.ToLower(strings.TrimSpace(message.Text))
	
	if strings.Contains(text, "да") || strings.Contains(text, "исправить") || text == "✅ да, исправить" {
		// Используем исправленные данные
		log.Printf("✅ Пользователь подтвердил исправление банка: %s", state.Data.BankName)
		
		// Продолжаем с исправленными данными - проверяем через API
		b.continueWithValidation(message, state.Data)
	} else {
		// Используем оригинальные данные без исправления
		log.Printf("❌ Пользователь отклонил исправление банка")
		
		b.sendMessage(message.Chat.ID, "Хорошо, оставляю как есть.")
		
		// Продолжаем валидацию с оригинальным названием
		// Для простоты просто завершим - пользователь может отправить заново
		delete(b.userStates, message.From.ID)
		b.sendMessage(message.Chat.ID, "Отправьте правило заново, если хотите продолжить.")
	}
}

// continueWithValidation продолжает валидацию данных через API
func (b *Bot) continueWithValidation(message *tgbotapi.Message, data *ParsedData) {
	userID := message.From.ID
	
	// Показываем распознанные данные
	b.sendMessage(message.Chat.ID, FormatParsedData(data))
	b.sendMessage(message.Chat.ID, "🔍 Проверяю данные...")

	// Вызываем /suggest для проверки
	suggestReq := &models.SuggestRequest{
		GroupName:       "Общие",
		Category:        data.Category,
		BankName:        data.BankName,
		UserDisplayName: getUserDisplayName(message.From),
		MonthYear:       data.MonthYear,
		CashbackPercent: data.CashbackPercent,
		MaxAmount:       data.MaxAmount,
	}

	suggestion, err := b.client.Suggest(suggestReq)
	if err != nil {
		b.sendMessage(message.Chat.ID, fmt.Sprintf("❌ Ошибка проверки: %s", err))
		delete(b.userStates, userID)
		return
	}

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

	// Если есть ошибки валидации
	if !suggestion.Valid {
		text := "❌ Ошибки валидации:\n" + strings.Join(suggestion.Errors, "\n")
		b.sendMessage(message.Chat.ID, text)
		delete(b.userStates, userID)
		return
	}

	// Если есть предложения по исправлению
	// Фильтруем только те предложения, которые реально отличаются
	var realSuggestions []string
	hasRealSuggestions := false
	
	if len(suggestion.Suggestions.BankName) > 0 {
		suggestedBank := suggestion.Suggestions.BankName[0].Value
		originalBank := strings.TrimSpace(data.BankName)
		suggestedBankTrimmed := strings.TrimSpace(suggestedBank)
		
		if originalBank != suggestedBankTrimmed {
			realSuggestions = append(realSuggestions, fmt.Sprintf("🏦 Банк: %s → %s",
				originalBank, suggestedBankTrimmed))
			hasRealSuggestions = true
		}
	}
	
	if len(suggestion.Suggestions.Category) > 0 {
		suggestedCategory := suggestion.Suggestions.Category[0].Value
		originalCategory := strings.TrimSpace(data.Category)
		suggestedCategoryTrimmed := strings.TrimSpace(suggestedCategory)
		
		if originalCategory != suggestedCategoryTrimmed {
			realSuggestions = append(realSuggestions, fmt.Sprintf("📁 Категория: %s → %s",
				originalCategory, suggestedCategoryTrimmed))
			hasRealSuggestions = true
		}
	}

	if hasRealSuggestions {
		text := "💡 Возможно, вы имели в виду:\n\n"
		text += strings.Join(realSuggestions, "\n")
		text += "\n\n❓ Исправить и сохранить?"
		
		// Сохраняем состояние для подтверждения
		b.userStates[userID] = &UserState{
			State:      "awaiting_confirmation",
			Data:       data,
			Suggestion: suggestion,
		}
		
		// Отправляем с кнопками
		b.sendMessageWithButtons(message.Chat.ID, text, [][]string{
			{"✅ Да, исправить", "❌ Нет, оставить как есть"},
			{"🚫 Отмена"},
		})
	} else {
		// Все отлично, сохраняем сразу
		b.saveRule(message.Chat.ID, message.From, data, false)
		delete(b.userStates, userID)
	}
}

// handleConfirmation обрабатывает подтверждение пользователя
func (b *Bot) handleConfirmation(message *tgbotapi.Message, state *UserState) {
	text := strings.ToLower(strings.TrimSpace(message.Text))
	
	switch {
	case strings.Contains(text, "да") || strings.Contains(text, "исправить"):
		// Применяем исправления
		data := state.Data
		if len(state.Suggestion.Suggestions.BankName) > 0 {
			data.BankName = state.Suggestion.Suggestions.BankName[0].Value
		}
		if len(state.Suggestion.Suggestions.Category) > 0 {
			data.Category = state.Suggestion.Suggestions.Category[0].Value
		}
		b.saveRule(message.Chat.ID, message.From, data, false)
		
	case strings.Contains(text, "нет") || strings.Contains(text, "оставить"):
		// Сохраняем как есть
		b.saveRule(message.Chat.ID, message.From, state.Data, true)
		
	case strings.Contains(text, "отмена"):
		b.sendMessage(message.Chat.ID, "🚫 Операция отменена")
		delete(b.userStates, message.From.ID)
		
	default:
		b.sendMessage(message.Chat.ID, "❓ Пожалуйста, выберите один из вариантов")
		return
	}
	
	delete(b.userStates, message.From.ID)
}

// saveRule сохраняет правило через API
func (b *Bot) saveRule(chatID int64, user *tgbotapi.User, data *ParsedData, force bool) {
	req := &models.CreateCashbackRequest{
		GroupName:       "Общие",
		Category:        data.Category,
		BankName:        data.BankName,
		UserID:          strconv.FormatInt(user.ID, 10),
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
		b.sendMessage(chatID, fmt.Sprintf("❌ Ошибка сохранения: %s", err))
		return
	}

	text := fmt.Sprintf(
		"✅ Правило успешно сохранено!\n\n"+
			"🆔 ID: %d\n"+
			"🏦 Банк: %s\n"+
			"📁 Категория: %s\n"+
			"📅 Месяц: %s\n"+
			"💰 Кэшбэк: %.1f%%\n"+
			"💵 Макс. сумма: %.0f₽\n"+
			"👤 Карта: %s",
		rule.ID,
		rule.BankName,
		rule.Category,
		rule.MonthYear.Format("2006-01"),
		rule.CashbackPercent,
		rule.MaxAmount,
		rule.UserDisplayName,
	)

	b.sendMessage(chatID, text)
}

// handleList обрабатывает команду /list
func (b *Bot) handleList(message *tgbotapi.Message) {
	userID := strconv.FormatInt(message.From.ID, 10)
	
	list, err := b.client.ListCashback(userID, 10, 0)
	if err != nil {
		b.sendMessage(message.Chat.ID, fmt.Sprintf("❌ Ошибка: %s", err))
		return
	}

	if len(list.Rules) == 0 {
		b.sendMessage(message.Chat.ID, "📝 У вас пока нет правил кэшбэка.\n\nИспользуйте /add для добавления.")
		return
	}

	text := fmt.Sprintf("📋 Все правила кэшбэка (%d):\n\n", list.Total)
	for i, rule := range list.Rules {
		text += fmt.Sprintf(
			"%d. %s - %s\n   %.1f%% до %.0f₽ (%s)\n   👤 Карта: %s\n   ID: %d\n\n",
			i+1,
			rule.BankName,
			rule.Category,
			rule.CashbackPercent,
			rule.MaxAmount,
			rule.MonthYear.Format("01/2006"),
			rule.UserDisplayName,
			rule.ID,
		)
	}

	b.sendMessage(message.Chat.ID, text)
}

// handleAddCommand обрабатывает команду /add
func (b *Bot) handleAddCommand(message *tgbotapi.Message) {
	text := "📝 Отправьте данные о кэшбэке.\n\n" +
		"Пример: \"Тинькофф такси 5% 3000р декабрь\"\n\n" +
		"Или используйте /cancel для отмены."
	
	b.sendMessage(message.Chat.ID, text)
}

// handleBestCommand обрабатывает команду /best
func (b *Bot) handleBestCommand(message *tgbotapi.Message) {
	text := "🔍 Для поиска лучшего кэшбэка отправьте:\n\n" +
		"📝 Просто напишите категорию и месяц:\n\n" +
		"Примеры:\n" +
		"• \"Лучший кэшбэк такси декабрь\"\n" +
		"• \"Где выгоднее рестораны январь\"\n" +
		"• \"Такси декабрь\"\n" +
		"• \"Супермаркеты февраль\""
	
	b.sendMessage(message.Chat.ID, text)
}

// isBestCashbackQuery проверяет, является ли сообщение запросом на поиск лучшего кэшбэка
func (b *Bot) isBestCashbackQuery(text string) bool {
	textLower := strings.ToLower(text)
	
	keywords := []string{
		"лучший кэшбэк", "лучший кешбек", "лучший cashback",
		"где выгоднее", "где лучше", "самый выгодный",
		"найди лучший", "покажи лучший",
	}
	
	for _, keyword := range keywords {
		if strings.Contains(textLower, keyword) {
			return true
		}
	}
	
	// Также проверяем паттерн "категория + месяц" без других данных
	// (нет процентов, сумм, названия банка)
	hasPercent := strings.Contains(textLower, "%") || strings.Contains(textLower, "процент")
	hasAmount := strings.Contains(textLower, "₽") || strings.Contains(textLower, "руб")
	hasBank := false
	
	banks := []string{"тинькофф", "сбер", "альфа", "втб", "райффайзен"}
	for _, bank := range banks {
		if strings.Contains(textLower, bank) {
			hasBank = true
			break
		}
	}
	
	// Если нет банка, процентов и сумм, но есть категория - это запрос на лучший кэшбэк
	if !hasBank && !hasPercent && !hasAmount {
		categories := []string{"такси", "ресторан", "кафе", "супермаркет", "аптек", "азс", "кино", "транспорт"}
		for _, cat := range categories {
			if strings.Contains(textLower, cat) {
				return true
			}
		}
	}
	
	return false
}

// handleBestQueryByCategory обрабатывает запрос на поиск лучшего кэшбэка по категории
// Всё сообщение = категория, месяц = текущий по умолчанию
func (b *Bot) handleBestQueryByCategory(message *tgbotapi.Message) {
	category := normalizeString(message.Text)
	
	if category == "" {
		b.sendMessage(message.Chat.ID, "❌ Укажите категорию. Например: \"Такси\"")
		return
	}
	
	// Используем текущий месяц по умолчанию
	now := time.Now()
	monthYear := fmt.Sprintf("%d-%02d", now.Year(), now.Month())
	
	b.sendMessage(message.Chat.ID, fmt.Sprintf("🔍 Ищу лучший кэшбэк для \"%s\" в этом месяце...", category))
	
	// Вызываем API для поиска лучшего кэшбэка
	rule, err := b.client.GetBestCashback("Общие", category, monthYear)
	if err != nil {
		// Пытаемся найти похожие категории
		categories, err2 := b.client.ListAllCategories("Общие", monthYear)
		if err2 == nil && len(categories) > 0 {
			similar, distance := findSimilarCategory(category, categories)
			simPercent := similarity(category, similar)
			
			// Если нашли похожую категорию (похожесть > 60%)
			if simPercent > 60.0 {
				b.sendMessage(message.Chat.ID, fmt.Sprintf("❌ Категория не найдена\n\n"+
					"📁 Вы написали: \"%s\"\n"+
					"💡 Возможно, вы имели в виду: \"%s\"\n\n"+
					"Попробуйте ещё раз с правильным названием!", 
					category, similar))
				log.Printf("🔍 Поиск похожей категории: '%s' → '%s' (расстояние: %d, похожесть: %.1f%%)",
					category, similar, distance, simPercent)
				return
			}
		}
		
		// Если не нашли похожих
		b.sendMessage(message.Chat.ID, fmt.Sprintf("❌ Кэшбэк не найден\n\n"+
			"📁 Категория: \"%s\"\n"+
			"📅 Месяц: %s\n\n"+
			"💡 Похоже, у вас ещё нет правил для этой категории.\n\n"+
			"Чтобы добавить правило, напишите через запятую:\n"+
			"Банк, %s, Процент, Сумма", 
			category, monthYear, category))
		return
	}
	
	text := fmt.Sprintf(
		"🏆 Лучший кэшбэк для \"%s\":\n\n"+
			"📅 Месяц: %s\n"+
			"🏦 Банк: %s\n"+
			"💰 Кэшбэк: %.1f%%\n"+
			"💵 Макс. сумма: %.0f₽\n"+
			"👤 Карта: %s",
		rule.Category,
		rule.MonthYear.Format("01/2006"),
		rule.BankName,
		rule.CashbackPercent,
		rule.MaxAmount,
		rule.UserDisplayName,
	)

	b.sendMessage(message.Chat.ID, text)
}

// handleBestQuery обрабатывает текстовый запрос на поиск лучшего кэшбэка
func (b *Bot) handleBestQuery(message *tgbotapi.Message) {
	// Парсим категорию и месяц из сообщения
	data, err := ParseMessage(message.Text)
	if err != nil {
		b.sendMessage(message.Chat.ID, "❌ Не удалось понять запрос. Попробуйте: \"Такси\"")
		return
	}
	
	if data.Category == "" {
		b.sendMessage(message.Chat.ID, "❌ Укажите категорию. Например: \"Такси\"")
		return
	}
	
	// Если месяц не указан, используем текущий
	if data.MonthYear == "" {
		now := time.Now()
		data.MonthYear = fmt.Sprintf("%d-%02d", now.Year(), now.Month())
	}
	
	b.sendMessage(message.Chat.ID, "🔍 Ищу лучший кэшбэк...")
	
	// Вызываем API для поиска лучшего кэшбэка
	rule, err := b.client.GetBestCashback("Общие", data.Category, data.MonthYear)
	if err != nil {
		b.sendMessage(message.Chat.ID, fmt.Sprintf("❌ Не найдено правил для категории \"%s\" в %s\n\nДобавьте правило или попробуйте другую категорию.", data.Category, data.MonthYear))
		return
	}
	
	text := fmt.Sprintf(
		"🏆 Лучший кэшбэк найден!\n\n"+
			"📁 Категория: %s\n"+
			"📅 Месяц: %s\n\n"+
			"🥇 Лучшее предложение:\n"+
			"🏦 Банк: %s\n"+
			"💰 Кэшбэк: %.1f%%\n"+
			"💵 Макс. сумма: %.0f₽\n"+
			"👤 Карта: %s",
		rule.Category,
		rule.MonthYear.Format("01/2006"),
		rule.BankName,
		rule.CashbackPercent,
		rule.MaxAmount,
		rule.UserDisplayName,
	)
	
	b.sendMessage(message.Chat.ID, text)
}

// handleCancel обрабатывает команду /cancel
func (b *Bot) handleCancel(message *tgbotapi.Message) {
	delete(b.userStates, message.From.ID)
	b.sendMessage(message.Chat.ID, "🚫 Операция отменена")
}

// handleCallback обрабатывает callback от inline кнопок
func (b *Bot) handleCallback(callback *tgbotapi.CallbackQuery) {
	// Здесь можно обрабатывать нажатия на inline кнопки
	b.api.Send(tgbotapi.NewCallback(callback.ID, ""))
}

// sendMessage отправляет текстовое сообщение
func (b *Bot) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	if _, err := b.api.Send(msg); err != nil {
		log.Printf("❌ Ошибка отправки сообщения: %v", err)
	}
}

// handleUpdateCommand обрабатывает команду /update ID
func (b *Bot) handleUpdateCommand(message *tgbotapi.Message) {
	args := strings.Fields(message.Text)
	if len(args) < 2 {
		b.sendMessage(message.Chat.ID, "❌ Укажите ID правила.\n\nПример: /update 5")
		return
	}

	id, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		b.sendMessage(message.Chat.ID, "❌ Неверный формат ID. Используйте число.")
		return
	}

	// Запрашиваем правило у API
	rule, err := b.client.GetCashbackByID(id)
	if err != nil {
		b.sendMessage(message.Chat.ID, fmt.Sprintf("❌ Правило с ID %d не найдено.", id))
		return
	}

	// Проверяем, что это правило пользователя
	if rule.UserID != strconv.FormatInt(message.From.ID, 10) {
		b.sendMessage(message.Chat.ID, "❌ Вы можете обновлять только свои правила.")
		return
	}

	// Показываем текущие данные
	text := fmt.Sprintf("📝 Обновление правила ID: %d\n\n"+
		"Текущие данные:\n"+
		"🏦 Банк: %s\n"+
		"📁 Категория: %s\n"+
		"📅 Месяц: %s\n"+
		"💰 Кэшбэк: %.1f%%\n"+
		"💵 Макс. сумма: %.0f₽\n\n"+
		"Отправьте новые данные через запятую:\n"+
		"Банк, Категория, Процент, Сумма[, Месяц]",
		rule.ID,
		rule.BankName,
		rule.Category,
		rule.MonthYear.Format("01/2006"),
		rule.CashbackPercent,
		rule.MaxAmount,
	)

	b.sendMessage(message.Chat.ID, text)

	// Сохраняем состояние ожидания данных
	b.userStates[message.From.ID] = &UserState{
		State:  "awaiting_update_data",
		RuleID: id,
	}
}

// handleUpdateData обрабатывает ввод новых данных для обновления
func (b *Bot) handleUpdateData(message *tgbotapi.Message, state *UserState) {
	// Парсим новые данные
	data, err := ParseMessage(message.Text)
	if err != nil {
		b.sendMessage(message.Chat.ID, fmt.Sprintf("❌ Ошибка парсинга: %s", err))
		return
	}

	// Проверяем данные
	missing := ValidateParsedData(data)
	if len(missing) > 0 {
		text := "⚠️ Не хватает данных:\n" + strings.Join(missing, ", ") + "\n\n" +
			"Формат: Банк, Категория, Процент, Сумма[, Месяц]"
		b.sendMessage(message.Chat.ID, text)
		return
	}

	// Обновляем правило через API
	req := &models.UpdateCashbackRequest{
		GroupName:       "Общие",
		Category:        data.Category,
		BankName:        data.BankName,
		MonthYear:       data.MonthYear,
		CashbackPercent: data.CashbackPercent,
		MaxAmount:       data.MaxAmount,
	}

	rule, err := b.client.UpdateCashback(state.RuleID, req)
	if err != nil {
		b.sendMessage(message.Chat.ID, fmt.Sprintf("❌ Ошибка обновления: %s", err))
		delete(b.userStates, message.From.ID)
		return
	}

	text := fmt.Sprintf("✅ Правило обновлено!\n\n"+
		"🆔 ID: %d\n"+
		"🏦 Банк: %s\n"+
		"📁 Категория: %s\n"+
		"📅 Месяц: %s\n"+
		"💰 Кэшбэк: %.1f%%\n"+
		"💵 Макс. сумма: %.0f₽",
		rule.ID,
		rule.BankName,
		rule.Category,
		rule.MonthYear.Format("2006-01"),
		rule.CashbackPercent,
		rule.MaxAmount,
	)

	b.sendMessage(message.Chat.ID, text)
	delete(b.userStates, message.From.ID)
}

// handleDeleteCommand обрабатывает команду /delete ID
func (b *Bot) handleDeleteCommand(message *tgbotapi.Message) {
	args := strings.Fields(message.Text)
	if len(args) < 2 {
		b.sendMessage(message.Chat.ID, "❌ Укажите ID правила.\n\nПример: /delete 5")
		return
	}

	id, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		b.sendMessage(message.Chat.ID, "❌ Неверный формат ID. Используйте число.")
		return
	}

	// Запрашиваем правило у API для проверки
	rule, err := b.client.GetCashbackByID(id)
	if err != nil {
		b.sendMessage(message.Chat.ID, fmt.Sprintf("❌ Правило с ID %d не найдено.", id))
		return
	}

	// Проверяем, что это правило пользователя
	if rule.UserID != strconv.FormatInt(message.From.ID, 10) {
		b.sendMessage(message.Chat.ID, "❌ Вы можете удалять только свои правила.")
		return
	}

	// Показываем данные и запрашиваем подтверждение
	text := fmt.Sprintf("⚠️ Вы уверены, что хотите удалить правило?\n\n"+
		"🆔 ID: %d\n"+
		"🏦 Банк: %s\n"+
		"📁 Категория: %s\n"+
		"💰 Кэшбэк: %.1f%%\n"+
		"💵 Макс. сумма: %.0f₽",
		rule.ID,
		rule.BankName,
		rule.Category,
		rule.CashbackPercent,
		rule.MaxAmount,
	)

	b.sendMessageWithButtons(message.Chat.ID, text, [][]string{
		{"✅ Да, удалить", "❌ Отмена"},
	})

	// Сохраняем состояние ожидания подтверждения
	b.userStates[message.From.ID] = &UserState{
		State:  "awaiting_delete_confirmation",
		RuleID: id,
	}
}

// handleDeleteConfirmation обрабатывает подтверждение удаления
func (b *Bot) handleDeleteConfirmation(message *tgbotapi.Message, state *UserState) {
	text := strings.ToLower(strings.TrimSpace(message.Text))

	if strings.Contains(text, "да") || strings.Contains(text, "удалить") {
		// Удаляем правило
		err := b.client.DeleteCashback(state.RuleID)
		if err != nil {
			b.sendMessage(message.Chat.ID, fmt.Sprintf("❌ Ошибка удаления: %s", err))
		} else {
			b.sendMessage(message.Chat.ID, fmt.Sprintf("✅ Правило ID %d успешно удалено!", state.RuleID))
		}
	} else {
		b.sendMessage(message.Chat.ID, "❌ Удаление отменено.")
	}

	delete(b.userStates, message.From.ID)
}

// sendMessageWithButtons отправляет сообщение с кнопками
func (b *Bot) sendMessageWithButtons(chatID int64, text string, buttons [][]string) {
	msg := tgbotapi.NewMessage(chatID, text)
	
	var keyboard [][]tgbotapi.KeyboardButton
	for _, row := range buttons {
		var keyboardRow []tgbotapi.KeyboardButton
		for _, btn := range row {
			keyboardRow = append(keyboardRow, tgbotapi.NewKeyboardButton(btn))
		}
		keyboard = append(keyboard, keyboardRow)
	}
	
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(keyboard...)
	
	if _, err := b.api.Send(msg); err != nil {
		log.Printf("❌ Ошибка отправки сообщения: %v", err)
	}
}

// getUserDisplayName получает отображаемое имя пользователя
func getUserDisplayName(user *tgbotapi.User) string {
	if user.FirstName != "" && user.LastName != "" {
		return user.FirstName + " " + user.LastName
	}
	if user.FirstName != "" {
		return user.FirstName
	}
	if user.UserName != "" {
		return user.UserName
	}
	return fmt.Sprintf("User%d", user.ID)
}

