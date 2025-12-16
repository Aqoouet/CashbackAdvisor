package bot

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rymax1e/open-cashback-advisor/internal/models"
)

// handleConfirmation обрабатывает подтверждение исправлений.
func (b *Bot) handleConfirmation(message *tgbotapi.Message, state *UserState) {
	text := strings.ToLower(strings.TrimSpace(message.Text))
	userID := message.From.ID

	switch {
	case isYesAnswer(text):
		// Применяем исправления
		data := state.Data
		if len(state.Suggestion.Suggestions.BankName) > 0 {
			data.BankName = state.Suggestion.Suggestions.BankName[0].Value
		}
		if len(state.Suggestion.Suggestions.Category) > 0 {
			data.Category = state.Suggestion.Suggestions.Category[0].Value
		}
		b.saveCashback(message.Chat.ID, message.From, data, false)

	case isNoAnswer(text):
		// Сохраняем как есть
		b.saveCashback(message.Chat.ID, message.From, state.Data, true)

	case isManualEditAnswer(text):
		// Переход в режим ручного ввода
		b.setState(userID, StateAwaitingManualInput, state.Data, state.Suggestion, 0)
		b.sendText(message.Chat.ID, "✏️ Отправьте данные в формате:\n"+
			"Банк, Категория, Процент, Сумма[, Дата окончания]\n\n"+
			"Или /cancel для отмены.")
		return

	case isCancelAnswer(text):
		b.sendText(message.Chat.ID, "🚫 Операция отменена")

	default:
		b.sendText(message.Chat.ID, "❓ Пожалуйста, выберите один из вариантов")
		return
	}

	b.clearState(userID)
}

// handleBankCorrection обрабатывает подтверждение исправления банка.
func (b *Bot) handleBankCorrection(message *tgbotapi.Message, state *UserState) {
	text := strings.ToLower(strings.TrimSpace(message.Text))
	userID := message.From.ID

	// Проверяем, это запрос из /bankinfo или из /add
	// Для /bankinfo: есть BankName и Category (groupName), но НЕТ CashbackPercent и MaxAmount
	// Для /add: есть все поля включая CashbackPercent и MaxAmount
	isBankInfoContext := state.Data != nil && 
		state.Data.BankName != "" && 
		state.Data.Category != "" &&
		state.Data.CashbackPercent == 0 && 
		state.Data.MaxAmount == 0
	
	if isBankInfoContext {
		groupName := state.Data.Category // Временно сохранили название группы в поле Category
		bankName := state.Data.BankName
		
		switch {
		case isYesAnswer(text):
			log.Printf("✅ Пользователь подтвердил исправление банка для /bankinfo: %s", bankName)
			b.clearState(userID)
			
			// Получаем данные по банку
			rules, err := b.client.GetCashbackByBank(groupName, bankName)
			if err != nil || len(rules) == 0 {
				b.sendText(message.Chat.ID, fmt.Sprintf("❌ Кешбек для банка \"%s\" не найден в вашей группе.", bankName))
				return
			}
			
			b.sendText(message.Chat.ID, formatBankInfo(bankName, rules))
			
		case isManualEditAnswer(text):
			log.Printf("✏️ Пользователь выбрал ручной ввод для /bankinfo")
			b.setState(userID, StateAwaitingBankInfoName, nil, nil, 0)
			b.sendText(message.Chat.ID, "🏦 Введите название банка.\n\nИли /cancel для отмены.")
			
		default:
			log.Printf("❌ Пользователь отклонил исправление банка для /bankinfo")
			b.clearState(userID)
			b.sendText(message.Chat.ID, "🚫 Операция отменена.")
		}
		return
	}

	// Обработка для добавления кешбека (старая логика)
	switch {
	case isYesAnswer(text):
		log.Printf("✅ Пользователь подтвердил исправление банка: %s", state.Data.BankName)
		b.continueWithValidation(message, state.Data)
		
	case isManualEditAnswer(text):
		// Переход в режим ручного ввода
		b.setState(userID, StateAwaitingManualInput, state.Data, nil, 0)
		b.sendText(message.Chat.ID, "✏️ Отправьте данные в формате:\n"+
			"Банк, Категория, Процент, Сумма[, Дата окончания]\n\n"+
			"Или /cancel для отмены.")
		
	default:
		log.Printf("❌ Пользователь отклонил исправление банка")
		b.sendText(message.Chat.ID, "Хорошо, оставляю как есть.")
		b.clearState(userID)
		b.sendText(message.Chat.ID, "Отправьте данные заново, если хотите продолжить.")
	}
}

// handleCategoryCorrection обрабатывает подтверждение исправления категории при поиске.
func (b *Bot) handleCategoryCorrection(message *tgbotapi.Message, state *UserState) {
	text := strings.ToLower(strings.TrimSpace(message.Text))
	userID := message.From.ID

	switch {
	case isYesAnswer(text):
		correctedCategory := state.Data.Category
		log.Printf("✅ Пользователь подтвердил исправление категории: %s", correctedCategory)
		b.clearState(userID)
		b.handleBestQueryWithCorrection(message, correctedCategory, true)
		
	case isManualEditAnswer(text):
		// Переход в режим ручного ввода для поиска
		b.clearState(userID)
		b.sendText(message.Chat.ID, "✏️ Введите название категории для поиска:\n\n"+
			"Или /cancel для отмены.")
		
	default:
		log.Printf("❌ Пользователь отклонил исправление категории")
		b.clearState(userID)
		b.sendText(message.Chat.ID, "Хорошо, попробуйте ввести название категории по-другому.")
	}
}

// handleUpdateData обрабатывает ввод новых данных для обновления.
func (b *Bot) handleUpdateData(message *tgbotapi.Message, state *UserState) {
	data, err := ParseMessage(message.Text)
	if err != nil {
		b.sendText(message.Chat.ID, fmt.Sprintf("❌ Ошибка парсинга: %s", err))
		return
	}

	missing := ValidateParsedData(data)
	if len(missing) > 0 {
		text := "⚠️ Не хватает данных:\n" + strings.Join(missing, ", ") + "\n\n" +
			"Формат: Банк, Категория, Процент, Сумма[, Месяц]"
		b.sendText(message.Chat.ID, text)
		return
	}

	userIDStr := strconv.FormatInt(message.From.ID, 10)
	groupName, err := b.client.GetUserGroup(userIDStr)
	if err != nil {
		b.sendText(message.Chat.ID, "❌ Вы должны быть в группе")
		return
	}

	req := &models.UpdateCashbackRequest{
		GroupName:       groupName,
		Category:        data.Category,
		BankName:        data.BankName,
		MonthYear:       data.MonthYear,
		CashbackPercent: data.CashbackPercent,
		MaxAmount:       data.MaxAmount,
	}

	_, err = b.client.UpdateCashback(state.RuleID, req)
	if err != nil {
		b.sendText(message.Chat.ID, fmt.Sprintf("❌ Ошибка обновления: %s", err))
		b.clearState(message.From.ID)
		return
	}

	// Получаем обновленные данные
	rule, err := b.client.GetCashbackByID(state.RuleID)
	if err != nil {
		b.sendText(message.Chat.ID, "✅ Кешбек обновлён!")
		b.clearState(message.From.ID)
		return
	}

	text := fmt.Sprintf(
		"✅ Кешбек обновлён!\n\n"+
			"🆔 ID: %d\n"+
			"🏦 Банк: %s\n"+
			"📁 Категория: %s\n"+
			"📅 До: %s\n"+
			"💰 Кэшбэк: %.1f%%\n"+
			"💵 Макс. сумма: %.0f₽",
		rule.ID,
		rule.BankName,
		rule.Category,
		rule.MonthYear.Format("02.01.2006"),
		rule.CashbackPercent,
		rule.MaxAmount,
	)

	b.sendText(message.Chat.ID, text)
	b.clearState(message.From.ID)
}

// handleDeleteConfirmation обрабатывает подтверждение удаления.
func (b *Bot) handleDeleteConfirmation(message *tgbotapi.Message, state *UserState) {
	text := strings.ToLower(strings.TrimSpace(message.Text))

	if isDeleteConfirm(text) {
		err := b.client.DeleteCashback(state.RuleID)
		if err != nil {
			b.sendText(message.Chat.ID, fmt.Sprintf("❌ Ошибка удаления: %s", err))
		} else {
			b.sendText(message.Chat.ID, fmt.Sprintf("✅ %% кешбек ID %d успешно удалён!", state.RuleID))
		}
	} else {
		b.sendText(message.Chat.ID, "❌ Удаление отменено.")
	}

	b.clearState(message.From.ID)
}

// isYesAnswer проверяет, является ли ответ положительным.
func isYesAnswer(text string) bool {
	return strings.Contains(text, "да") ||
		strings.Contains(text, "исправить") ||
		text == "✅ да, исправить"
}

// isNoAnswer проверяет, является ли ответ отрицательным.
func isNoAnswer(text string) bool {
	return strings.Contains(text, "нет") ||
		strings.Contains(text, "оставить")
}

// isCancelAnswer проверяет, является ли ответ отменой.
func isCancelAnswer(text string) bool {
	return strings.Contains(text, "отмена")
}

// isDeleteConfirm проверяет, подтверждает ли пользователь удаление.
func isDeleteConfirm(text string) bool {
	return strings.Contains(text, "да") ||
		strings.Contains(text, "удалить")
}

// isManualEditAnswer проверяет, хочет ли пользователь ввести вручную.
func isManualEditAnswer(text string) bool {
	return strings.Contains(text, "изменить вручную") ||
		strings.Contains(text, "✏️")
}

// handleManualInput обрабатывает ручной ввод данных.
func (b *Bot) handleManualInput(message *tgbotapi.Message, state *UserState) {
	// Парсим новые данные
	data, err := ParseMessage(message.Text)
	if err != nil {
		b.sendText(message.Chat.ID, fmt.Sprintf("❌ Ошибка парсинга: %s", err))
		return
	}

	// Проверяем полноту данных
	missing := ValidateParsedData(data)
	if len(missing) > 0 {
		text := "⚠️ Не хватает данных:\n" + strings.Join(missing, ", ") + "\n\n" +
			"Формат: Банк, Категория, Процент, Сумма[, Дата окончания]"
		b.sendText(message.Chat.ID, text)
		return
	}

	log.Printf("✅ Ручной ввод: Bank='%s', Category='%s', Percent=%.1f%%, Amount=%.0f",
		data.BankName, data.Category, data.CashbackPercent, data.MaxAmount)

	// Сохраняем без дополнительной валидации
	b.saveCashback(message.Chat.ID, message.From, data, true)
	b.clearState(message.From.ID)
}

// handleAddDataInput обрабатывает ввод данных для команды /add.
func (b *Bot) handleAddDataInput(message *tgbotapi.Message) {
	userID := message.From.ID
	
	// Проверка на отмену
	if isCancelAnswer(message.Text) {
		b.clearState(userID)
		b.sendText(message.Chat.ID, "🚫 Операция отменена")
		return
	}
	
	// Очищаем состояние и обрабатываем как новый кэшбэк
	b.clearState(userID)
	b.handleNewCashback(message, userID)
}

// handleBestCategoryInput обрабатывает ввод категории для команды /best.
func (b *Bot) handleBestCategoryInput(message *tgbotapi.Message) {
	userID := message.From.ID
	category := strings.TrimSpace(message.Text)
	
	// Проверка на отмену
	if isCancelAnswer(category) {
		b.clearState(userID)
		b.sendText(message.Chat.ID, "🚫 Операция отменена")
		return
	}
	
	// Очищаем состояние и выполняем поиск
	b.clearState(userID)
	b.handleBestQueryWithCorrection(message, category, false)
}

// handleBankInfoNameInput обрабатывает ввод названия банка для команды /bankinfo.
func (b *Bot) handleBankInfoNameInput(message *tgbotapi.Message) {
	userID := message.From.ID
	bankName := strings.TrimSpace(message.Text)
	
	// Проверка на отмену
	if isCancelAnswer(bankName) {
		b.clearState(userID)
		b.sendText(message.Chat.ID, "🚫 Операция отменена")
		return
	}
	
	// Валидация: название банка не должно быть пустым и не слишком коротким
	if len(bankName) < 2 {
		b.sendText(message.Chat.ID, "❌ Название банка слишком короткое. Введите корректное название или /cancel для отмены.")
		return
	}
	
	// Валидация: название не должно содержать только цифры
	if isOnlyDigits(bankName) {
		b.sendText(message.Chat.ID, "❌ Название банка не может состоять только из цифр. Введите корректное название или /cancel для отмены.")
		return
	}
	
	// Очищаем состояние
	b.clearState(userID)
	
	// Получаем группу
	userIDStr := strconv.FormatInt(userID, 10)
	groupName, err := b.client.GetUserGroup(userIDStr)
	if err != nil {
		b.sendText(message.Chat.ID, "❌ Вы должны быть в группе. Используйте /creategroup или /joingroup")
		return
	}
	
	// Получаем данные
	rules, err := b.client.GetCashbackByBank(groupName, bankName)
	if err != nil || len(rules) == 0 {
		// Не найден точный банк - ищем похожие
		log.Printf("⚠️ Банк '%s' не найден, ищу похожие банки", bankName)
		b.trySuggestSimilarBank(message, bankName, groupName)
		return
	}
	
	b.sendText(message.Chat.ID, formatBankInfo(bankName, rules))
}

// isOnlyDigits проверяет, состоит ли строка только из цифр.
func isOnlyDigits(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// handleUpdateIDInput обрабатывает ввод ID для команды /update.
func (b *Bot) handleUpdateIDInput(message *tgbotapi.Message) {
	userID := message.From.ID
	idStr := strings.TrimSpace(message.Text)
	
	// Проверка на отмену
	if isCancelAnswer(idStr) {
		b.clearState(userID)
		b.sendText(message.Chat.ID, "🚫 Операция отменена")
		return
	}
	
	// Парсим ID
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		b.sendText(message.Chat.ID, "❌ Неверный формат ID. Введите число или /cancel для отмены.")
		return
	}
	
	// Получаем правило
	rule, err := b.client.GetCashbackByID(id)
	if err != nil {
		b.sendText(message.Chat.ID, fmt.Sprintf("❌ Кешбек с ID %d не найден.", id))
		b.clearState(userID)
		return
	}
	
	// Проверяем владельца
	if rule.UserID != strconv.FormatInt(userID, 10) {
		b.sendText(message.Chat.ID, "❌ Вы можете обновлять только свой кешбек.")
		b.clearState(userID)
		return
	}
	
	// Переходим к ожиданию данных для обновления
	// Отправляем первое сообщение с инструкцией
	b.sendText(message.Chat.ID, formatUpdatePrompt(rule))
	
	// Отправляем второе сообщение только со строкой для копирования
	copyLine := fmt.Sprintf("%s, %s, %.1f, %.0f, %s",
		rule.BankName,
		rule.Category,
		rule.CashbackPercent,
		rule.MaxAmount,
		rule.MonthYear.Format("02.01.2006"),
	)
	b.sendTextPlain(message.Chat.ID, copyLine)
	
	b.setState(userID, StateAwaitingUpdateData, nil, nil, id)
}

// handleDeleteIDInput обрабатывает ввод ID для команды /delete.
func (b *Bot) handleDeleteIDInput(message *tgbotapi.Message) {
	userID := message.From.ID
	idStr := strings.TrimSpace(message.Text)
	
	// Проверка на отмену
	if isCancelAnswer(idStr) {
		b.clearState(userID)
		b.sendText(message.Chat.ID, "🚫 Операция отменена")
		return
	}
	
	// Парсим ID
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		b.sendText(message.Chat.ID, "❌ Неверный формат ID. Введите число или /cancel для отмены.")
		return
	}
	
	// Получаем правило
	rule, err := b.client.GetCashbackByID(id)
	if err != nil {
		b.sendText(message.Chat.ID, fmt.Sprintf("❌ Кешбек с ID %d не найден.", id))
		b.clearState(userID)
		return
	}
	
	// Проверяем владельца
	if rule.UserID != strconv.FormatInt(userID, 10) {
		b.sendText(message.Chat.ID, "❌ Вы можете удалять только свой кешбек.")
		b.clearState(userID)
		return
	}
	
	// Переходим к подтверждению удаления
	text := fmt.Sprintf(
		"⚠️ Вы уверены, что хотите удалить этот кешбек?\n\n"+
			"🏦 Банк: %s\n"+
			"📁 Категория: %s\n"+
			"💰 %.1f%%%% до %.0f₽\n"+
			"📅 До %s\n\n"+
			"❓ Удалить?",
		rule.BankName, rule.Category, rule.CashbackPercent,
		rule.MaxAmount, rule.MonthYear.Format("02.01.2006"),
	)
	
	b.setState(userID, StateAwaitingDeleteConfirm, nil, nil, id)
	b.sendWithButtons(message.Chat.ID, text, ButtonsDelete)
}

// handleJoinGroupNameInput обрабатывает ввод названия группы для команды /joingroup.
func (b *Bot) handleJoinGroupNameInput(message *tgbotapi.Message) {
	userID := message.From.ID
	userIDStr := strconv.FormatInt(userID, 10)
	groupName := strings.TrimSpace(message.Text)
	
	log.Printf("🔍 [JOINGROUP_INPUT] Начало обработки ввода для пользователя @%s (ID: %s), введено: \"%s\"", 
		message.From.UserName, userIDStr, groupName)
	
	// Проверка на отмену
	if isCancelAnswer(groupName) {
		log.Printf("🚫 [JOINGROUP_INPUT] Пользователь @%s отменил операцию", message.From.UserName)
		b.clearState(userID)
		b.sendText(message.Chat.ID, "🚫 Операция отменена")
		return
	}
	
	// Очищаем состояние
	log.Printf("🔍 [JOINGROUP_INPUT] Очищаю состояние пользователя @%s", message.From.UserName)
	b.clearState(userID)
	
	log.Printf("🔍 [JOINGROUP_INPUT] Обработанное название группы: \"%s\" (длина: %d)", groupName, len(groupName))
	
	// Проверяем существование группы
	log.Printf("🔍 [JOINGROUP_INPUT] Проверяю существование группы \"%s\"...", groupName)
	groupExists := b.client.GroupExists(groupName)
	log.Printf("🔍 [JOINGROUP_INPUT] Результат проверки существования группы \"%s\": %v", groupName, groupExists)
	
	if !groupExists {
		log.Printf("❌ [JOINGROUP_INPUT] Группа \"%s\" не существует для пользователя @%s", 
			groupName, message.From.UserName)
		b.sendText(message.Chat.ID, fmt.Sprintf(
			"❌ Группа \"%s\" не существует.\n\nСоздайте её: /creategroup %s",
			groupName, groupName,
		))
		return
	}
	
	// Проверяем текущую группу пользователя
	log.Printf("🔍 [JOINGROUP_INPUT] Проверяю текущую группу пользователя @%s (ID: %s)...", 
		message.From.UserName, userIDStr)
	currentGroup, err := b.client.GetUserGroup(userIDStr)
	if err != nil {
		log.Printf("ℹ️ [JOINGROUP_INPUT] Пользователь @%s не состоит ни в какой группе (ошибка: %v)", 
			message.From.UserName, err)
	} else {
		log.Printf("ℹ️ [JOINGROUP_INPUT] Пользователь @%s состоит в группе: \"%s\"", 
			message.From.UserName, currentGroup)
		if currentGroup == groupName {
			log.Printf("⚠️ [JOINGROUP_INPUT] Пользователь @%s уже состоит в группе \"%s\"", 
				message.From.UserName, currentGroup)
			b.sendText(message.Chat.ID, fmt.Sprintf("⚠️ Вы уже состоите в группе \"%s\"", groupName))
			return
		}
		
		log.Printf("👥 [JOINGROUP_INPUT] Пользователь @%s переключается из группы \"%s\" в группу \"%s\"",
			message.From.UserName, currentGroup, groupName)
	}
	
	// Добавляем пользователя в группу
	log.Printf("🔍 [JOINGROUP_INPUT] Пытаюсь присоединить пользователя @%s (ID: %s) к группе \"%s\"...", 
		message.From.UserName, userIDStr, groupName)
	err = b.client.JoinGroup(userIDStr, groupName)
	if err != nil {
		log.Printf("❌ [JOINGROUP_INPUT] Ошибка присоединения пользователя @%s к группе \"%s\": %v", 
			message.From.UserName, groupName, err)
		b.sendText(message.Chat.ID, fmt.Sprintf("❌ %s", err))
		return
	}
	
	log.Printf("✅ [JOINGROUP_INPUT] Пользователь @%s успешно присоединился к группе \"%s\"", 
		message.From.UserName, groupName)
	b.sendText(message.Chat.ID, fmt.Sprintf(
		"✅ Вы присоединились к группе \"%s\"!\n\n"+
			"Теперь вы можете:\n"+
			"• Добавлять кэшбэк: /add\n"+
			"• Искать лучший кэшбэк: /best\n"+
			"• Смотреть список: /list",
		groupName,
	))
}

// handleCreateGroupNameInput обрабатывает ввод названия группы для команды /creategroup.
func (b *Bot) handleCreateGroupNameInput(message *tgbotapi.Message) {
	userID := message.From.ID
	groupName := strings.TrimSpace(message.Text)
	
	// Проверка на отмену
	if isCancelAnswer(groupName) {
		b.clearState(userID)
		b.sendText(message.Chat.ID, "🚫 Операция отменена")
		return
	}
	
	// Очищаем состояние
	b.clearState(userID)
	
	userIDStr := strconv.FormatInt(userID, 10)
	
	// Создаём группу (метод CreateGroup автоматически добавляет создателя в группу)
	err := b.client.CreateGroup(groupName, userIDStr)
	if err != nil {
		b.sendText(message.Chat.ID, fmt.Sprintf("❌ %s", err))
		return
	}
	
	b.sendText(message.Chat.ID, fmt.Sprintf(
		"✅ Группа \"%s\" успешно создана и вы к ней присоединились!\n\n"+
			"Вы можете пригласить друзей командой:\n/joingroup %s",
		groupName, groupName,
	))
}

