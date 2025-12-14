package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Константы для кнопок.
const (
	BtnYesCorrect  = "✅ Да, исправить"
	BtnNoKeepAsIs  = "❌ Нет, оставить как есть"
	BtnCancel      = "🚫 Отмена"
	BtnYesDelete   = "✅ Да, удалить"
	BtnCancelShort = "❌ Отмена"
)

// Предустановленные наборы кнопок.
var (
	// ButtonsConfirm — кнопки подтверждения исправления.
	ButtonsConfirm = [][]string{
		{BtnYesCorrect, BtnNoKeepAsIs},
		{BtnCancel},
	}

	// ButtonsConfirmSimple — простые кнопки да/нет.
	ButtonsConfirmSimple = [][]string{
		{BtnYesCorrect, BtnNoKeepAsIs},
	}

	// ButtonsDelete — кнопки подтверждения удаления.
	ButtonsDelete = [][]string{
		{BtnYesDelete, BtnCancelShort},
	}
)

// defaultCommandButtons возвращает кнопки быстрого доступа к командам.
func defaultCommandButtons() [][]string {
	return [][]string{
		{"/help", "/list"},
		{"/update", "/groupinfo"},
	}
}

// buildKeyboard формирует клавиатуру Telegram.
func buildKeyboard(buttons [][]string) [][]tgbotapi.KeyboardButton {
	var keyboard [][]tgbotapi.KeyboardButton

	// Основные кнопки
	for _, row := range buttons {
		var keyboardRow []tgbotapi.KeyboardButton
		for _, btn := range row {
			keyboardRow = append(keyboardRow, tgbotapi.NewKeyboardButton(btn))
		}
		keyboard = append(keyboard, keyboardRow)
	}

	// Добавляем кнопки команд
	for _, row := range defaultCommandButtons() {
		var keyboardRow []tgbotapi.KeyboardButton
		for _, btn := range row {
			keyboardRow = append(keyboardRow, tgbotapi.NewKeyboardButton(btn))
		}
		keyboard = append(keyboard, keyboardRow)
	}

	return keyboard
}

