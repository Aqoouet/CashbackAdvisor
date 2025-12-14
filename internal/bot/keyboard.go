package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Константы для кнопок.
const (
	BtnYesCorrect    = "✅ Да, исправить"
	BtnNoKeepAsIs    = "❌ Нет, оставить как есть"
	BtnManualEdit    = "✏️ Изменить вручную"
	BtnCancel        = "🚫 Отмена"
	BtnYesDelete     = "✅ Да, удалить"
	BtnCancelShort   = "❌ Отмена"
	BtnNavPrev       = "◀️"
	BtnNavNext       = "▶️"
)

// Предустановленные наборы кнопок.
var (
	// ButtonsConfirm — кнопки подтверждения исправления с ручным вводом.
	ButtonsConfirm = [][]string{
		{BtnYesCorrect, BtnNoKeepAsIs},
		{BtnManualEdit},
		{BtnCancel},
	}

	// ButtonsConfirmSimple — простые кнопки да/нет с ручным вводом.
	ButtonsConfirmSimple = [][]string{
		{BtnYesCorrect, BtnNoKeepAsIs},
		{BtnManualEdit},
	}

	// ButtonsDelete — кнопки подтверждения удаления.
	ButtonsDelete = [][]string{
		{BtnYesDelete, BtnCancelShort},
	}
)

// Все доступные команды для пагинации.
var allCommands = []string{
	"/start", "/help", "/add", "/best",
	"/list", "/update", "/delete", "/bankinfo",
	"/categorylist", "/banklist", "/userinfo", "/groupinfo",
	"/joingroup", "/creategroup",
}

// getTotalCommandPages возвращает общее количество страниц команд.
func getTotalCommandPages() int {
	const commandsPerPage = 4
	return (len(allCommands) + commandsPerPage - 1) / commandsPerPage
}

// getCommandPage возвращает страницу команд с навигацией.
func getCommandPage(page int) [][]string {
	const commandsPerPage = 4
	const commandsPerRow = 2
	
	totalPages := getTotalCommandPages()
	
	// Нормализуем номер страницы
	if page < 0 {
		page = 0
	}
	if page >= totalPages {
		page = totalPages - 1
	}
	
	start := page * commandsPerPage
	end := start + commandsPerPage
	if end > len(allCommands) {
		end = len(allCommands)
	}
	
	pageCommands := allCommands[start:end]
	
	// Формируем кнопки по 2 в ряд
	var rows [][]string
	for i := 0; i < len(pageCommands); i += commandsPerRow {
		end := i + commandsPerRow
		if end > len(pageCommands) {
			end = len(pageCommands)
		}
		rows = append(rows, pageCommands[i:end])
	}
	
	// Добавляем навигацию, если страниц больше одной
	if totalPages > 1 {
		navRow := []string{}
		if page > 0 {
			navRow = append(navRow, BtnNavPrev)
		}
		// Добавляем кнопку /cancel посередине
		navRow = append(navRow, "/cancel")
		// Не добавляем индикатор, чтобы избежать распознавания как команды
		if page < totalPages-1 {
			navRow = append(navRow, BtnNavNext)
		}
		if len(navRow) > 0 {
			rows = append(rows, navRow)
		}
	}
	
	return rows
}

// buildKeyboard формирует клавиатуру Telegram с пагинацией команд.
func buildKeyboard(buttons [][]string) [][]tgbotapi.KeyboardButton {
	return buildKeyboardWithPage(buttons, 0)
}

// buildKeyboardWithPage формирует клавиатуру с указанной страницей команд.
func buildKeyboardWithPage(buttons [][]string, page int) [][]tgbotapi.KeyboardButton {
	var keyboard [][]tgbotapi.KeyboardButton

	// Основные кнопки (если есть)
	for _, row := range buttons {
		var keyboardRow []tgbotapi.KeyboardButton
		for _, btn := range row {
			keyboardRow = append(keyboardRow, tgbotapi.NewKeyboardButton(btn))
		}
		keyboard = append(keyboard, keyboardRow)
	}

	// Добавляем кнопки команд с пагинацией
	commandRows := getCommandPage(page)
	for _, row := range commandRows {
		var keyboardRow []tgbotapi.KeyboardButton
		for _, btn := range row {
			keyboardRow = append(keyboardRow, tgbotapi.NewKeyboardButton(btn))
		}
		keyboard = append(keyboard, keyboardRow)
	}

	return keyboard
}

