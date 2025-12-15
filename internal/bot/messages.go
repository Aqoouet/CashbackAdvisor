package bot

import (
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rymax1e/open-cashback-advisor/internal/models"
)

// sendText отправляет текстовое сообщение с клавиатурой по умолчанию.
func (b *Bot) sendText(chatID int64, text string) {
	// Получаем текущую страницу пользователя
	page := 0
	if state, exists := b.userStates[chatID]; exists {
		page = state.KeyboardPage
	}
	
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"

	kb := tgbotapi.NewReplyKeyboard(buildKeyboardWithPage(nil, page)...)
	kb.ResizeKeyboard = true
	msg.ReplyMarkup = kb

	if _, err := b.api.Send(msg); err != nil {
		log.Printf("❌ Ошибка отправки сообщения: %v", err)
	}
}

// sendTextPlain отправляет текстовое сообщение БЕЗ HTML парсинга (для таблиц).
func (b *Bot) sendTextPlain(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	// НЕ используем ParseMode для совместимости с таблицами

	kb := tgbotapi.NewReplyKeyboard(buildKeyboard(nil)...)
	kb.ResizeKeyboard = true
	msg.ReplyMarkup = kb

	if _, err := b.api.Send(msg); err != nil {
		log.Printf("❌ Ошибка отправки сообщения: %v", err)
	}
}

// sendWithButtons отправляет сообщение с указанными кнопками.
func (b *Bot) sendWithButtons(chatID int64, text string, buttons [][]string) {
	msg := tgbotapi.NewMessage(chatID, text)

	keyboard := buildKeyboard(buttons)
	kb := tgbotapi.NewReplyKeyboard(keyboard...)
	kb.ResizeKeyboard = true
	msg.ReplyMarkup = kb

	if _, err := b.api.Send(msg); err != nil {
		log.Printf("❌ Ошибка отправки сообщения: %v", err)
	}
}

// FormatParsedData форматирует распознанные данные для отображения.
func FormatParsedData(data *ParsedData) string {
	return fmt.Sprintf(
		"📋 Распознанные данные:\n\n"+
			"🏦 Банк: %s\n"+
			"📁 Категория: %s\n"+
			"📅 Действует до: %s\n"+
			"💰 Кэшбэк: %.1f%%\n"+
			"💵 Макс. сумма: %.0f₽",
		data.BankName,
		data.Category,
		data.MonthYear,
		data.CashbackPercent,
		data.MaxAmount,
	)
}

// formatCashbackRule форматирует правило кэшбэка для отображения.
func formatCashbackRule(rule *models.CashbackRule) string {
	return fmt.Sprintf(
		"🆔 ID: %d\n"+
			"🏦 Банк: %s\n"+
			"📁 Категория: %s\n"+
			"📅 Действует до: %s\n"+
			"💰 Кэшбэк: %.1f%%\n"+
			"💵 Макс. сумма: %.0f₽\n"+
			"👤 Карта: %s",
		rule.ID,
		rule.BankName,
		rule.Category,
		rule.MonthYear.Format("02.01.2006"),
		rule.CashbackPercent,
		rule.MaxAmount,
		rule.UserDisplayName,
	)
}

// formatSavedCashback форматирует сохранённый кэшбэк.
func formatSavedCashback(rule *models.CashbackRule) string {
	return fmt.Sprintf(
		"✅ Кешбек успешно сохранён!\n\n"+
			"🆔 ID: %d\n"+
			"🏦 Банк: %s\n"+
			"📁 Категория: %s\n"+
			"📅 Действует до: %s\n"+
			"💰 Кэшбэк: %.1f%%\n"+
			"💵 Макс. сумма: %.0f₽\n"+
			"👤 Карта: %s",
		rule.ID,
		rule.BankName,
		rule.Category,
		rule.MonthYear.Format("02.01.2006"),
		rule.CashbackPercent,
		rule.MaxAmount,
		rule.UserDisplayName,
	)
}

// formatBestCashback форматирует лучший кэшбэк с учетом fallback.
func formatBestCashback(rule *models.CashbackRule, requestedCategory string, isFallback bool) string {
	var text string
	
	if isFallback {
		text = fmt.Sprintf(
			"💡 Кэшбэк для категории \"%s\" не найден.\n"+
				"Показываю кэшбэк на \"Все покупки\":\n\n"+
				"🏦 Банк: %s\n"+
				"📅 Действует до: %s\n"+
				"💰 Кэшбэк: %.1f%%\n"+
				"💵 Макс. сумма: %.0f₽\n"+
				"👤 Карта: %s",
			requestedCategory,
			rule.BankName,
			rule.MonthYear.Format("02.01.2006"),
			rule.CashbackPercent,
			rule.MaxAmount,
			rule.UserDisplayName,
		)
	} else {
		text = fmt.Sprintf(
		"🏆 Лучший кэшбэк для \"%s\":\n\n"+
			"🏦 Банк: %s\n"+
				"📅 Действует до: %s\n"+
			"💰 Кэшбэк: %.1f%%\n"+
			"💵 Макс. сумма: %.0f₽\n"+
			"👤 Карта: %s",
		rule.Category,
		rule.BankName,
			rule.MonthYear.Format("02.01.2006"),
		rule.CashbackPercent,
		rule.MaxAmount,
		rule.UserDisplayName,
	)
	}
	
	return text
}

// formatAllCashbackResults форматирует все найденные кэшбэки по категории.
func formatAllCashbackResults(rules []models.CashbackRule, requestedCategory string, isFallback bool) string {
	if len(rules) == 0 {
		return "❌ Кэшбэк не найден"
	}
	
	var text string
	
	if isFallback {
		text = fmt.Sprintf("💡 Кэшбэк для категории \"%s\" не найден.\n"+
			"Показываю кэшбэк на \"Все покупки\" (%d вариант", requestedCategory, len(rules))
		if len(rules) == 1 {
			text += "):\n\n"
		} else if len(rules) < 5 {
			text += "а):\n\n"
		} else {
			text += "ов):\n\n"
		}
	} else {
		text = fmt.Sprintf("🏆 Все кэшбэки для \"%s\" (%d вариант", requestedCategory, len(rules))
		if len(rules) == 1 {
			text += "):\n\n"
		} else if len(rules) < 5 {
			text += "а):\n\n"
		} else {
			text += "ов):\n\n"
		}
	}
	
	for i, rule := range rules {
		medal := ""
		if i == 0 {
			medal = "🥇 "
		} else if i == 1 {
			medal = "🥈 "
		} else if i == 2 {
			medal = "🥉 "
		} else {
			medal = fmt.Sprintf("%d. ", i+1)
		}
		
		text += fmt.Sprintf(
			"%s🏦 %s\n"+
				"   📁 %s\n"+
				"   💰 %.1f%% до %.0f₽\n"+
				"   📅 До %s\n"+
				"   👤 %s\n\n",
			medal,
			rule.BankName,
			rule.Category,
			rule.CashbackPercent,
			rule.MaxAmount,
			rule.MonthYear.Format("02.01.2006"),
			rule.UserDisplayName,
		)
	}
	
	return text
}

// formatCashbackList форматирует список кэшбэков.
func formatCashbackList(rules []models.CashbackRule, total int) string {
	if len(rules) == 0 {
		return "📝 Пока нет кешбека в группе.\n\nДобавьте первым!"
	}

	text := fmt.Sprintf("📋 Все кешбека группы (%d):\n\n", total)

	for i, rule := range rules {
		text += fmt.Sprintf(
			"%d. %s - %s\n   %.1f%% до %.0f₽ (до %s)\n   👤 Карта: %s\n   ID: %d\n\n",
			i+1,
			rule.BankName,
			rule.Category,
			rule.CashbackPercent,
			rule.MaxAmount,
			rule.MonthYear.Format("02.01.2006"),
			rule.UserDisplayName,
			rule.ID,
		)
	}

	return text
}

// formatCashbackListTable форматирует список кэшбэков в табличном виде.
func formatCashbackListTable(rules []models.CashbackRule, total int, showAll bool, indices []int) string {
	if len(rules) == 0 {
		return "📝 Пока нет кешбека в группе.\n\nДобавьте первым!"
	}

	var header string
	if showAll {
		header = fmt.Sprintf("📋 Все кешбека группы (%d):\n\n", total)
	} else if indices == nil {
		header = fmt.Sprintf("📋 Последние 5 кешбеков (всего %d):\n\n", total)
	} else {
		header = fmt.Sprintf("📋 Выбранные кешбеки (всего %d):\n\n", total)
	}

	text := header
	
	for i, rule := range rules {
		text += fmt.Sprintf(
			"%d. 🏦 %s\n"+
			"   📁 %s\n"+
			"   💰 %.1f%% до %.0f₽\n"+
			"   📅 До %s\n"+
			"   👤 %s (ID: %d)\n\n",
			i+1,
			rule.BankName,
			rule.Category,
			rule.CashbackPercent,
			rule.MaxAmount,
			rule.MonthYear.Format("02.01.2006"),
			rule.UserDisplayName,
			rule.ID,
		)
	}
	
	if !showAll && indices == nil && total > 5 {
		text += "💡 Используйте:\n"
		text += "• /list all - показать все\n"
		text += "• /list 1-10 - показать с 1 по 10\n"
		text += "• /list 1-5,8 - показать 1-5 и 8"
	}

	return text
}

// truncateString обрезает строку до указанной длины
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-2] + ".."
}

// formatBankInfo форматирует информацию о кэшбэках банка.
func formatBankInfo(bankName string, rules []models.CashbackRule) string {
	text := fmt.Sprintf("🏦 Активные кэшбэки банка \"%s\" (%d):\n\n", bankName, len(rules))

	for i, rule := range rules {
		text += fmt.Sprintf(
			"%d. 📁 %s\n"+
				"   💰 %.1f%% до %.0f₽\n"+
				"   📅 До %s\n"+
				"   👤 %s\n\n",
			i+1,
			rule.Category,
			rule.CashbackPercent,
			rule.MaxAmount,
			rule.MonthYear.Format("02.01.2006"),
			rule.UserDisplayName,
		)
	}

	return text
}

// formatCategoryList форматирует список активных категорий.
func formatCategoryList(categories []string) string {
	text := fmt.Sprintf("📁 Активные категории (%d):\n\n", len(categories))

	for i, category := range categories {
		text += fmt.Sprintf("%d. %s\n", i+1, category)
	}

	text += "\n💡 Используйте /best для поиска лучшего кэшбэка по категории"

	return text
}

// formatBankList форматирует список активных банков.
func formatBankList(banks []string) string {
	text := fmt.Sprintf("🏦 Активные банки (%d):\n\n", len(banks))

	for i, bank := range banks {
		text += fmt.Sprintf("%d. %s\n", i+1, bank)
	}

	text += "\n💡 Используйте /bankinfo [название] для просмотра кэшбэков банка"

	return text
}

// formatUserInfo форматирует информацию о кэшбэках пользователя.
func formatUserInfo(rules []models.CashbackRule, groupName string) string {
	if len(rules) == 0 {
		return "📝 Нет кэшбэков"
	}

	userName := rules[0].UserDisplayName
	
	// Подсчитываем активные кешбеки
	now := time.Now()
	activeCount := 0
	for _, rule := range rules {
		if rule.MonthYear.After(now.AddDate(0, 0, -1)) {
			activeCount++
		}
	}
	
	text := fmt.Sprintf("👤 Кэшбэки пользователя <b>%s</b>\n\n", userName)
	text += fmt.Sprintf("👥 Группа: %s\n", groupName)
	text += fmt.Sprintf("💳 Всего кешбеков: %d (активных: %d)\n\n", len(rules), activeCount)

	for i, rule := range rules {
		// Помечаем истекшие кешбеки
		statusIcon := ""
		if rule.MonthYear.Before(now.AddDate(0, 0, -1)) {
			statusIcon = " ⏰"
		}
		
		text += fmt.Sprintf(
			"%d. 🏦 %s%s\n"+
				"   📁 %s\n"+
				"   💰 %.1f%% до %.0f₽\n"+
				"   📅 До %s\n"+
				"   🆔 ID: %d\n\n",
			i+1,
			rule.BankName,
			statusIcon,
			rule.Category,
			rule.CashbackPercent,
			rule.MaxAmount,
			rule.MonthYear.Format("02.01.2006"),
			rule.ID,
		)
	}

	return text
}

// formatUserListTable форматирует список пользователей в табличном виде.
func formatUserListTable(users []models.UserInfo, total int) string {
	if len(users) == 0 {
		return "📝 Нет пользователей"
	}

	text := fmt.Sprintf("👥 Пользователи группы \"%s\" (показано %d из %d):\n\n", 
		users[0].GroupName, len(users), total)
	
	for i, user := range users {
		text += fmt.Sprintf(
			"%d. 👤 %s\n   ID: %s\n\n",
			i+1,
			user.UserDisplayName,
			user.UserID,
		)
	}
	if len(users) < total {
		text += "💡 Используйте:\n"
		text += "• /userlist - все пользователи\n"
		text += "• /userlist 1-10 - с 1 по 10"
	}

	return text
}

// formatUpdatePrompt форматирует запрос на обновление с текущей строкой для копирования.
func formatUpdatePrompt(rule *models.CashbackRule) string {
	return fmt.Sprintf(
		"📝 Обновление кешбека ID: %d\n\n"+
			"Текущие данные:\n"+
			"🏦 Банк: %s\n"+
			"📁 Категория: %s\n"+
			"📅 Действует до: %s\n"+
			"💰 Кэшбэк: %.1f%%\n"+
			"💵 Макс. сумма: %.0f₽\n\n"+
			"✏️ Скопируйте строку ниже, измените и отправьте новые данные:",
		rule.ID,
		rule.BankName,
		rule.Category,
		rule.MonthYear.Format("02.01.2006"),
		rule.CashbackPercent,
		rule.MaxAmount,
	)
}

// formatDeletePrompt форматирует запрос на удаление.
func formatDeletePrompt(rule *models.CashbackRule) string {
	return fmt.Sprintf(
		"⚠️ Вы уверены, что хотите удалить кешбек?\n\n"+
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
}

// formatNotFoundMessage форматирует сообщение о ненайденном кэшбэке.
func formatNotFoundMessage(category, monthYear string) string {
	return fmt.Sprintf(
		"❌ Кэшбэк не найден\n\n"+
			"📁 Категория: \"%s\"\n"+
			"📅 Месяц: %s\n\n"+
			"💡 Похоже, ещё нет кешбека для этой категории.\n\n"+
			"Чтобы добавить, напишите через запятую:\n"+
			"Банк, %s, Процент, Сумма[, Дата окончания]",
		category, monthYear, category,
	)
}

