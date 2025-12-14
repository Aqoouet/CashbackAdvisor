package bot

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rymax1e/open-cashback-advisor/internal/models"
)

// sendText отправляет текстовое сообщение с клавиатурой по умолчанию.
func (b *Bot) sendText(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"

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

// formatUpdatePrompt форматирует запрос на обновление.
func formatUpdatePrompt(rule *models.CashbackRule) string {
	return fmt.Sprintf(
		"📝 Обновление кешбека ID: %d\n\n"+
			"Текущие данные:\n"+
			"🏦 Банк: %s\n"+
			"📁 Категория: %s\n"+
			"📅 Действует до: %s\n"+
			"💰 Кэшбэк: %.1f%%\n"+
			"💵 Макс. сумма: %.0f₽\n\n"+
			"Отправьте новые данные через запятую:\n"+
			"Банк, Категория, Процент, Сумма[, Дата окончания]",
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

