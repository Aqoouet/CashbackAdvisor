// Package bot предоставляет функциональность Telegram бота для управления кэшбэком.
package bot

import (
	"fmt"
	"log"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rymax1e/open-cashback-advisor/internal/models"
)

// UserStateType определяет тип состояния пользователя.
type UserStateType string

// Константы состояний пользователя.
const (
	StateNone                       UserStateType = ""
	StateAwaitingConfirmation       UserStateType = "awaiting_confirmation"
	StateAwaitingBankCorrection     UserStateType = "awaiting_bank_correction"
	StateAwaitingCategoryCorrection UserStateType = "awaiting_category_correction"
	StateAwaitingUpdateData         UserStateType = "awaiting_update_data"
	StateAwaitingDeleteConfirm      UserStateType = "awaiting_delete_confirmation"
	StateAwaitingGroupName          UserStateType = "awaiting_group_name"
	StateAwaitingManualInput        UserStateType = "awaiting_manual_input"
	StateAwaitingAddData            UserStateType = "awaiting_add_data"
	StateAwaitingBestCategory       UserStateType = "awaiting_best_category"
	StateAwaitingBankInfoName       UserStateType = "awaiting_bankinfo_name"
	StateAwaitingUpdateID           UserStateType = "awaiting_update_id"
	StateAwaitingDeleteID           UserStateType = "awaiting_delete_id"
	StateAwaitingJoinGroupName      UserStateType = "awaiting_joingroup_name"
	StateAwaitingCreateGroupName    UserStateType = "awaiting_creategroup_name"
)

// UserState хранит состояние диалога с пользователем.
type UserState struct {
	State       UserStateType
	Data        *ParsedData
	Suggestion  *models.SuggestResponse
	RuleID      int64
	KeyboardPage int // Текущая страница клавиатуры
}

// Bot представляет Telegram бота для работы с кэшбэком.
type Bot struct {
	api        *tgbotapi.BotAPI
	client     *APIClient
	userStates map[int64]*UserState
}

// NewBot создаёт нового бота.
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

// Start запускает основной цикл обработки сообщений.
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

// handleMessage маршрутизирует входящие сообщения.
func (b *Bot) handleMessage(message *tgbotapi.Message) {
	log.Printf("📨 Сообщение от @%s: %s", message.From.UserName, message.Text)

	// Обработка кнопок навигации (до обработки команд)
	if b.handleNavigationButtons(message) {
		return
	}

	// Обработка команд - только если команда отправлена вручную (не через кнопку)
	// Кнопки ReplyKeyboard уже вставляют текст в поле ввода, пользователь должен нажать "Отправить" сам
	// Но мы все равно обрабатываем команды, когда пользователь их отправляет
	if message.IsCommand() {
		b.routeCommand(message)
		return
	}

	// Проверяем членство в группе
	if !b.checkGroupMembership(message) {
		return
	}

	// Обработка состояний пользователя
	if b.handleUserState(message) {
		return
	}

	// Если сообщение не команда и нет активного состояния - подсказываем использовать команды
	b.sendText(message.Chat.ID, "ℹ️ Используйте команды для работы с ботом.\n\n"+
		"Например:\n"+
		"/add - Добавить кэшбэк\n"+
		"/best - Найти лучший кэшбэк\n"+
		"/list - Список кэшбэков\n\n"+
		"Полный список команд: /help")
}

// routeCommand маршрутизирует команды к соответствующим обработчикам.
func (b *Bot) routeCommand(message *tgbotapi.Message) {
	switch message.Command() {
	case "start":
		b.handleStart(message)
	case "help":
		b.handleHelp(message)
	case "creategroup":
		b.handleCreateGroup(message)
	case "joingroup":
		b.handleJoinGroup(message)
	case "groupinfo":
		b.handleGroupInfo(message)
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
	case "bankinfo":
		b.handleBankInfo(message)
	case "categorylist":
		b.handleCategoryList(message)
	case "banklist":
		b.handleBankList(message)
	case "userinfo":
		b.handleUserInfo(message)
	case "userlist":
		b.handleUserList(message)
	case "cancel":
		b.handleCancel(message)
	default:
		b.sendText(message.Chat.ID, "❌ Неизвестная команда. Используйте /help для справки.")
	}
}

// checkGroupMembership проверяет, состоит ли пользователь в группе.
func (b *Bot) checkGroupMembership(message *tgbotapi.Message) bool {
	userIDStr := strconv.FormatInt(message.From.ID, 10)
	_, err := b.client.GetUserGroup(userIDStr)
	if err != nil {
		b.sendText(message.Chat.ID,
			"⚠️ Вы не состоите в группе!\n\n"+
				"Сначала создайте группу или присоединитесь к существующей:\n"+
				"/creategroup название - создать новую группу\n"+
				"/joingroup название - присоединиться к группе")
		return false
	}
	return true
}

// handleUserState обрабатывает сообщение в контексте текущего состояния.
// Возвращает true, если сообщение было обработано.
func (b *Bot) handleUserState(message *tgbotapi.Message) bool {
	userID := message.From.ID
	userIDStr := strconv.FormatInt(userID, 10)
	
	state, exists := b.userStates[userID]
	if !exists {
		log.Printf("🔍 [HANDLE_STATE] Пользователь @%s (ID: %s) не имеет активного состояния", 
			message.From.UserName, userIDStr)
		return false
	}

	log.Printf("🔍 [HANDLE_STATE] Пользователь @%s (ID: %s) имеет состояние: %s", 
		message.From.UserName, userIDStr, state.State)

	switch state.State {
	case StateAwaitingConfirmation:
		b.handleConfirmation(message, state)
	case StateAwaitingBankCorrection:
		b.handleBankCorrection(message, state)
	case StateAwaitingCategoryCorrection:
		b.handleCategoryCorrection(message, state)
	case StateAwaitingUpdateData:
		b.handleUpdateData(message, state)
	case StateAwaitingDeleteConfirm:
		b.handleDeleteConfirmation(message, state)
	case StateAwaitingGroupName:
		b.handleGroupNameInput(message)
	case StateAwaitingManualInput:
		b.handleManualInput(message, state)
	case StateAwaitingAddData:
		b.handleAddDataInput(message)
	case StateAwaitingBestCategory:
		b.handleBestCategoryInput(message)
	case StateAwaitingBankInfoName:
		b.handleBankInfoNameInput(message)
	case StateAwaitingUpdateID:
		b.handleUpdateIDInput(message)
	case StateAwaitingDeleteID:
		b.handleDeleteIDInput(message)
	case StateAwaitingJoinGroupName:
		log.Printf("🔍 [HANDLE_STATE] Вызываю handleJoinGroupNameInput для пользователя @%s", message.From.UserName)
		b.handleJoinGroupNameInput(message)
	case StateAwaitingCreateGroupName:
		log.Printf("🔍 [HANDLE_STATE] Вызываю handleCreateGroupNameInput для пользователя @%s", message.From.UserName)
		b.handleCreateGroupNameInput(message)
	default:
		log.Printf("⚠️ [HANDLE_STATE] Неизвестное состояние %s для пользователя @%s", state.State, message.From.UserName)
		return false
	}

	log.Printf("✅ [HANDLE_STATE] Состояние обработано для пользователя @%s", message.From.UserName)
	return true
}

// handleCallback обрабатывает callback от inline кнопок.
func (b *Bot) handleCallback(callback *tgbotapi.CallbackQuery) {
	b.api.Send(tgbotapi.NewCallback(callback.ID, ""))
}

// setState устанавливает состояние пользователя.
func (b *Bot) setState(userID int64, state UserStateType, data *ParsedData, suggestion *models.SuggestResponse, ruleID int64) {
	b.userStates[userID] = &UserState{
		State:      state,
		Data:       data,
		Suggestion: suggestion,
		RuleID:     ruleID,
	}
}

// clearState очищает состояние пользователя.
func (b *Bot) clearState(userID int64) {
	delete(b.userStates, userID)
}

// getUserGroup получает группу пользователя или пустую строку.
func (b *Bot) getUserGroup(userID int64) string {
	userIDStr := strconv.FormatInt(userID, 10)
	groupName, err := b.client.GetUserGroup(userIDStr)
	if err != nil {
		return ""
	}
	return groupName
}

// getUserDisplayName возвращает отображаемое имя пользователя.
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

// handleNavigationButtons обрабатывает нажатие на кнопки навигации.
func (b *Bot) handleNavigationButtons(message *tgbotapi.Message) bool {
	userID := message.From.ID
	
	// Получаем текущую страницу
	state, exists := b.userStates[userID]
	currentPage := 0
	if exists {
		currentPage = state.KeyboardPage
	}
	
	// Вычисляем общее количество страниц
	totalPages := getTotalCommandPages()
	
	// Обрабатываем навигацию
	switch message.Text {
	case BtnNavPrev:
		if currentPage > 0 {
			currentPage--
			b.setKeyboardPage(userID, currentPage)
			// Обновляем клавиатуру
			b.sendTextWithPage(message.Chat.ID, "📋", currentPage)
		}
		return true
		
	case BtnNavNext:
		if currentPage < totalPages-1 {
			currentPage++
			b.setKeyboardPage(userID, currentPage)
			// Обновляем клавиатуру
			b.sendTextWithPage(message.Chat.ID, "📋", currentPage)
		}
		return true
	}
	
	return false
}

// setKeyboardPage устанавливает текущую страницу клавиатуры для пользователя.
func (b *Bot) setKeyboardPage(userID int64, page int) {
	state, exists := b.userStates[userID]
	if !exists {
		state = &UserState{}
		b.userStates[userID] = state
	}
	state.KeyboardPage = page
}

// sendTextWithPage отправляет сообщение с клавиатурой на указанной странице.
func (b *Bot) sendTextWithPage(chatID int64, text string, page int) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"

	kb := tgbotapi.NewReplyKeyboard(buildKeyboardWithPage(nil, page)...)
	kb.ResizeKeyboard = true
	msg.ReplyMarkup = kb

	if _, err := b.api.Send(msg); err != nil {
		log.Printf("❌ Ошибка отправки сообщения: %v", err)
	}
}

