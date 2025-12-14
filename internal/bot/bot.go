// Package bot предоставляет функциональность Telegram бота для управления кэшбэком.
package bot

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rymax1e/open-cashback-advisor/internal/models"
)

// UserStateType определяет тип состояния пользователя.
type UserStateType string

// Константы состояний пользователя.
const (
	StateNone                   UserStateType = ""
	StateAwaitingConfirmation   UserStateType = "awaiting_confirmation"
	StateAwaitingBankCorrection UserStateType = "awaiting_bank_correction"
	StateAwaitingCategoryCorrection UserStateType = "awaiting_category_correction"
	StateAwaitingUpdateData     UserStateType = "awaiting_update_data"
	StateAwaitingDeleteConfirm  UserStateType = "awaiting_delete_confirmation"
	StateAwaitingGroupName      UserStateType = "awaiting_group_name"
)

// UserState хранит состояние диалога с пользователем.
type UserState struct {
	State      UserStateType
	Data       *ParsedData
	Suggestion *models.SuggestResponse
	RuleID     int64
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
	userID := message.From.ID

	log.Printf("📨 Сообщение от @%s: %s", message.From.UserName, message.Text)

	// Обработка команд
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

	// Определяем тип сообщения по наличию запятой
	if strings.Contains(message.Text, ",") {
		b.handleNewCashback(message, userID)
	} else {
		b.handleBestQueryByCategory(message)
	}
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
	state, exists := b.userStates[message.From.ID]
	if !exists {
		return false
	}

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
	default:
		return false
	}

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

