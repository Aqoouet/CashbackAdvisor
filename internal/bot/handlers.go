package bot

import (
	"fmt"
	"log"
	"strconv"
	"strings"

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
		case "cancel":
			b.handleCancel(message)
		default:
			b.sendMessage(message.Chat.ID, "❌ Неизвестная команда. Используйте /help для справки.")
		}
		return
	}

	// Обработка состояний пользователя
	state, exists := b.userStates[userID]
	if exists && state.State == "awaiting_confirmation" {
		b.handleConfirmation(message, state)
		return
	}

	// Проверяем, не запрос ли это на поиск лучшего кэшбэка
	if b.isBestCashbackQuery(message.Text) {
		b.handleBestQuery(message)
		return
	}

	// Обычное сообщение - парсим как новое правило
	b.handleNewRule(message)
}

// handleStart обрабатывает команду /start
func (b *Bot) handleStart(message *tgbotapi.Message) {
	text := "👋 Привет! Я помогаю не упустить выгодный кэшбэк.\n\n" +
		"🎯 Что я умею:\n" +
		"• Запоминаю условия кэшбэка от разных банков\n" +
		"• Подсказываю, где сейчас самый выгодный кэшбэк\n" +
		"• Исправляю опечатки в названиях банков и категорий\n" +
		"• Показываю все твои сохранённые правила\n\n" +
		"✍️ Добавить правило - просто напиши:\n" +
		"\"Тинькофф такси 5% 3000р декабрь\"\n\n" +
		"🔍 Найти лучший кэшбэк - напиши:\n" +
		"\"Лучший кэшбэк такси декабрь\" или /best\n\n" +
		"📋 Команды:\n" +
		"/list - мои правила\n" +
		"/best - найти лучший кэшбэк\n" +
		"/help - подробная справка\n\n" +
		"Я пойму, проверю и сохраню! 😊"

	b.sendMessage(message.Chat.ID, text)
}

// handleHelp обрабатывает команду /help
func (b *Bot) handleHelp(message *tgbotapi.Message) {
	text := "📖 Подробная справка:\n\n" +
		"🔹 /add - Добавить новое правило кэшбэка\n" +
		"🔹 /list - Показать мои правила\n" +
		"🔹 /best - Найти лучший кэшбэк\n" +
		"🔹 /cancel - Отменить текущую операцию\n\n" +
		"💡 Примеры добавления правил:\n" +
		"• \"Сбер супермаркеты 10% 5000р январь\"\n" +
		"• \"Альфа рестораны 7.5% 4000 февраль\"\n" +
		"• \"ВТБ такси 5 процентов 3000 рублей март\"\n\n" +
		"🔍 Примеры поиска лучшего кэшбэка:\n" +
		"• \"Лучший кэшбэк такси декабрь\"\n" +
		"• \"Где выгоднее рестораны январь\"\n" +
		"• Просто напиши категорию и месяц!\n\n" +
		"✨ Бот умеет исправлять опечатки! 😊"

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

	// Проверяем, что все данные есть
	missing := ValidateParsedData(data)
	if len(missing) > 0 {
		text := "⚠️ Не хватает данных:\n" + strings.Join(missing, ", ") + "\n\n" +
			"Пример: \"Тинькофф такси 5% 3000р декабрь\""
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

	// Если есть ошибки валидации
	if !suggestion.Valid {
		text := "❌ Ошибки валидации:\n" + strings.Join(suggestion.Errors, "\n")
		b.sendMessage(message.Chat.ID, text)
		return
	}

	// Если есть предложения по исправлению
	hasSuggestions := len(suggestion.Suggestions.BankName) > 0 ||
		len(suggestion.Suggestions.Category) > 0 ||
		len(suggestion.Suggestions.GroupName) > 0 ||
		len(suggestion.Suggestions.UserDisplayName) > 0

	if hasSuggestions {
		text := "💡 Возможно, вы имели в виду:\n\n"
		
		if len(suggestion.Suggestions.BankName) > 0 {
			text += fmt.Sprintf("🏦 Банк: %s (вы написали: %s)\n",
				suggestion.Suggestions.BankName[0].Value, data.BankName)
		}
		if len(suggestion.Suggestions.Category) > 0 {
			text += fmt.Sprintf("📁 Категория: %s (вы написали: %s)\n",
				suggestion.Suggestions.Category[0].Value, data.Category)
		}
		
		text += "\n❓ Исправить и сохранить?"
		
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
			"💵 Макс. сумма: %.0f₽",
		rule.ID,
		rule.BankName,
		rule.Category,
		rule.MonthYear.Format("2006-01"),
		rule.CashbackPercent,
		rule.MaxAmount,
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

	text := fmt.Sprintf("📋 Ваши правила (%d):\n\n", list.Total)
	for i, rule := range list.Rules {
		text += fmt.Sprintf(
			"%d. %s - %s\n   %.1f%% до %.0f₽ (%s)\n   ID: %d\n\n",
			i+1,
			rule.BankName,
			rule.Category,
			rule.CashbackPercent,
			rule.MaxAmount,
			rule.MonthYear.Format("01/2006"),
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

// handleBestQuery обрабатывает текстовый запрос на поиск лучшего кэшбэка
func (b *Bot) handleBestQuery(message *tgbotapi.Message) {
	// Парсим категорию и месяц из сообщения
	data, err := ParseMessage(message.Text)
	if err != nil {
		b.sendMessage(message.Chat.ID, "❌ Не удалось понять запрос. Попробуйте: \"Такси декабрь\"")
		return
	}
	
	if data.Category == "" {
		b.sendMessage(message.Chat.ID, "❌ Укажите категорию. Например: \"Такси декабрь\"")
		return
	}
	
	if data.MonthYear == "" {
		b.sendMessage(message.Chat.ID, "❌ Укажите месяц. Например: \"Такси декабрь\"")
		return
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
			"👤 Добавил: %s",
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

