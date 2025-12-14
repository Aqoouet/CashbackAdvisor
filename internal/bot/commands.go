package bot

import (
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/rymax1e/open-cashback-advisor/internal/models"
)

// CommandHelp содержит справку по команде.
type CommandHelp struct {
	Name        string
	ShortDesc   string
	LongDesc    string
	Usage       string
	Examples    []string
}

// commandHelpMap содержит детальную справку по всем командам.
var commandHelpMap = map[string]CommandHelp{
	"start": {
		Name:      "/start",
		ShortDesc: "Начало работы с ботом",
		LongDesc: "Команда /start показывает приветствие и краткую справку о возможностях бота.\n\n" +
			"⚠️ ВАЖНО: Бот НЕ ищет информацию в интернете! Он показывает только кэшбэк, " +
			"добавленный участниками вашей группы.",
		Usage:    "/start",
		Examples: []string{"/start"},
	},
	"help": {
		Name:      "/help",
		ShortDesc: "Справка по командам",
		LongDesc: "Команда /help показывает список всех доступных команд.\n\n" +
			"Вы можете получить детальную справку по конкретной команде, указав её название.",
		Usage:    "/help [команда]",
		Examples: []string{"/help", "/help add", "/help best"},
	},
	"add": {
		Name:      "/add",
		ShortDesc: "Добавить кэшбэк",
		LongDesc: "Добавляет новый кэшбэк в базу данных группы.\n\n" +
			"Формат ввода: Банк, Категория, Процент, Сумма[, Дата окончания]\n\n" +
			"Поддерживается мультистрочный ввод - вы можете добавить несколько кэшбэков одним сообщением.",
		Usage: "Отправьте данные через запятую",
		Examples: []string{
			"Тинькофф, Такси, 5%, 3000",
			"Сбер, Супермаркеты, 10, 5000, 31.01.2025",
			"Альфа, Рестораны, 7.5, 4000, 28.02.2025",
		},
	},
	"best": {
		Name:      "/best",
		ShortDesc: "Найти лучший кэшбэк",
		LongDesc: "Ищет все кэшбэки по указанной категории и показывает их, отсортированными по убыванию процента.\n\n" +
			"Бот умеет исправлять опечатки и предлагает похожие категории, если точного совпадения не найдено.",
		Usage: "Просто напишите категорию (без запятых)",
		Examples: []string{
			"Такси",
			"Рестораны",
			"Супермаркеты",
		},
	},
	"list": {
		Name:      "/list",
		ShortDesc: "Список всех кэшбэков группы",
		LongDesc: "Показывает кэшбэки группы в табличном виде с возможностью пагинации.\n\n" +
			"По умолчанию показывает последние 5 записей.",
		Usage: "/list [параметры]",
		Examples: []string{
			"/list - последние 5",
			"/list all - все записи",
			"/list 1-10 - записи с 1 по 10",
			"/list 1-5,8,10 - записи 1-5, 8 и 10",
		},
	},
	"update": {
		Name:      "/update",
		ShortDesc: "Обновить свой кэшбэк",
		LongDesc: "Обновляет существующий кэшбэк по его ID.\n\n" +
			"Бот покажет текущую строку в формате для копирования - вы можете скопировать, " +
			"изменить нужные поля и отправить обратно.",
		Usage:    "/update <ID>",
		Examples: []string{"/update 5", "/update 12"},
	},
	"delete": {
		Name:      "/delete",
		ShortDesc: "Удалить свой кэшбэк",
		LongDesc:  "Удаляет кэшбэк по указанному ID. Вы можете удалять только свои записи.",
		Usage:     "/delete <ID>",
		Examples:  []string{"/delete 5", "/delete 12"},
	},
	"bankinfo": {
		Name:      "/bankinfo",
		ShortDesc: "Информация о кэшбэках банка",
		LongDesc: "Показывает все активные кэшбэки указанного банка в вашей группе.\n\n" +
			"Бот автоматически исправляет опечатки в названии банка.",
		Usage:    "/bankinfo <название банка>",
		Examples: []string{"/bankinfo Тинькофф", "/bankinfo Сбер", "/bankinfo Альфа"},
	},
	"categorylist": {
		Name:      "/categorylist",
		ShortDesc: "Список всех активных категорий",
		LongDesc:  "Показывает все уникальные категории, по которым есть активный (не истекший) кэшбэк в группе.",
		Usage:     "/categorylist",
		Examples:  []string{"/categorylist"},
	},
	"banklist": {
		Name:      "/banklist",
		ShortDesc: "Список всех активных банков",
		LongDesc:  "Показывает все уникальные банки, по которым есть активный (не истекший) кэшбэк в группе.",
		Usage:     "/banklist",
		Examples:  []string{"/banklist"},
	},
	"userinfo": {
		Name:      "/userinfo",
		ShortDesc: "Кэшбэки пользователя",
		LongDesc: "Показывает все кэшбэки указанного пользователя.\n\n" +
			"Без параметров показывает ваши собственные кэшбэки.",
		Usage:    "/userinfo [ID]",
		Examples: []string{"/userinfo", "/userinfo 123456789"},
	},
	"userlist": {
		Name:      "/userlist",
		ShortDesc: "Список пользователей группы",
		LongDesc: "Показывает всех пользователей группы в табличном виде.\n\n" +
			"Поддерживает пагинацию для удобного просмотра.",
		Usage: "/userlist [параметры]",
		Examples: []string{
			"/userlist - все пользователи",
			"/userlist 1-10 - с 1 по 10",
			"/userlist 1,3,5 - пользователи 1, 3 и 5",
		},
	},
	"creategroup": {
		Name:      "/creategroup",
		ShortDesc: "Создать новую группу",
		LongDesc: "Создаёт новую группу с указанным названием.\n\n" +
			"Вы автоматически становитесь участником созданной группы.",
		Usage:    "/creategroup <название>",
		Examples: []string{"/creategroup Семья", "/creategroup Друзья"},
	},
	"joingroup": {
		Name:      "/joingroup",
		ShortDesc: "Присоединиться к группе",
		LongDesc: "Присоединяет вас к существующей группе.\n\n" +
			"Группа должна быть предварительно создана командой /creategroup.",
		Usage:    "/joingroup <название>",
		Examples: []string{"/joingroup Семья", "/joingroup Друзья"},
	},
	"groupinfo": {
		Name:      "/groupinfo",
		ShortDesc: "Информация о группе",
		LongDesc:  "Показывает информацию о вашей текущей группе или указанной группе.",
		Usage:     "/groupinfo [название]",
		Examples:  []string{"/groupinfo", "/groupinfo Семья"},
	},
	"cancel": {
		Name:      "/cancel",
		ShortDesc: "Отменить текущую операцию",
		LongDesc:  "Отменяет текущую операцию и сбрасывает состояние диалога.",
		Usage:     "/cancel",
		Examples:  []string{"/cancel"},
	},
}

// handleStart обрабатывает команду /start.
func (b *Bot) handleStart(message *tgbotapi.Message) {
	text := fmt.Sprintf(`👋 Привет! Я помогаю не упустить выгодный кэшбэк.

🎯 Что я умею:
• Запоминаю условия кэшбэка от разных банков
• Подсказываю, где сейчас самый выгодный кэшбэк
• Исправляю опечатки в названиях банков и категорий
• Работаю с группами - делитесь с друзьями!

⚠️ ВАЖНО: 
Бот НЕ находит информацию в интернете!
Он показывает только кэшбэк, добавленный участниками вашей группы.
Без группы бот работать не будет!

👥 Группы (обязательно!):
/creategroup Название - создать
/joingroup Название - присоединиться
/groupinfo - инфо о вашей группе

✍️ Добавить кешбек (дата опциональна):
"Тинькофф, Такси, 5%%, 3000"
"Сбер, Перекресток доставка, 12%%, 30000, 31.01.2025"

🔍 Найти лучший кэшбэк (без запятых):
"Такси" - покажет для текущего месяца
"Перекресток доставка"

📋 Команды:
/list - все кешбека группы
/best - найти лучший кэшбэк
/update ID - обновить свой кешбек
/delete ID - удалить свой кешбек
/help - подробная справка

Я пойму, проверю и сохраню! 😊

ℹ️ Версия: %s`, BuildInfo())

	b.sendText(message.Chat.ID, text)
}

// handleHelp обрабатывает команду /help [command_name].
func (b *Bot) handleHelp(message *tgbotapi.Message) {
	// Парсим аргументы
	args := strings.TrimPrefix(message.Text, "/help")
	args = strings.TrimSpace(args)

	// Если указана конкретная команда
	if args != "" {
		b.handleCommandHelp(message, args)
		return
	}

	// Общая справка
	text := fmt.Sprintf(`📖 Подробная справка (Версия: %s)

⚠️ ВАЖНАЯ ИНФОРМАЦИЯ:
Бот НЕ ищет информацию в интернете!
Он показывает только кэшбэк, который добавили участники вашей группы.

💡 Назначение бота:
• Сохранять информацию о кэшбэке ваших карт
• Делиться кэшбэком с друзьями в группе
• Быстро находить лучшее предложение

🔐 Группы нужны для:
• Разделения кэшбэка разных коллективов
• Совместного использования информации
• Без группы бот НЕ РАБОТАЕТ!

👥 Группы:
• /creategroup (название) - Создать группу
• /joingroup (название) - Присоединиться
• /groupinfo [название] - Информация

💳 Основные команды:
• /add - Добавить кешбек
• /best (категория) - Найти лучший кэшбэк
• /list [параметры] - Список кешбеков
• /update (ID) - Обновить кешбек
• /delete (ID) - Удалить кешбек

📊 Информация:
• /bankinfo (банк) - Кэшбэки по банку
• /categorylist - Список категорий
• /banklist - Список банков
• /userinfo [ID] - Кэшбэки пользователя
• /userlist [параметры] - Список пользователей

💡 Детальная справка по команде:
/help (команда)

Например: /help add, /help best

✨ Бот умеет исправлять опечатки! 😊`, BuildInfo())

	b.sendText(message.Chat.ID, text)
}

// handleCommandHelp показывает детальную справку по конкретной команде.
func (b *Bot) handleCommandHelp(message *tgbotapi.Message, commandName string) {
	// Убираем / если пользователь ввёл /help /add
	commandName = strings.TrimPrefix(commandName, "/")

	help, exists := commandHelpMap[commandName]
	if !exists {
		b.sendText(message.Chat.ID, fmt.Sprintf("❌ Команда /%s не найдена.\n\n"+
			"Используйте /help для просмотра всех команд.", commandName))
		return
	}

	text := fmt.Sprintf("📖 Справка по команде %s\n\n", help.Name)
	text += fmt.Sprintf("📝 %s\n\n", help.ShortDesc)
	text += fmt.Sprintf("%s\n\n", help.LongDesc)
	text += fmt.Sprintf("💡 Использование:\n%s\n\n", help.Usage)

	if len(help.Examples) > 0 {
		text += "📚 Примеры:\n"
		for _, example := range help.Examples {
			text += fmt.Sprintf("• %s\n", example)
		}
	}

	b.sendText(message.Chat.ID, text)
}

// handleAddCommand обрабатывает команду /add.
func (b *Bot) handleAddCommand(message *tgbotapi.Message) {
	text := `📝 Отправьте данные о кэшбэке.

Формат: Банк, Категория, Процент, Сумма[, Дата окончания]

Примеры:
• "Тинькофф, Такси, 5%, 3000"
• "Сбер, Супермаркеты, 10, 5000, 31.01.2025"

Или используйте /cancel для отмены.`

	b.sendText(message.Chat.ID, text)
}

// handleBestCommand обрабатывает команду /best.
func (b *Bot) handleBestCommand(message *tgbotapi.Message) {
	// Устанавливаем состояние ожидания категории
	b.setState(message.From.ID, StateAwaitingBestCategory, nil, nil, 0)
	
	text := `🔍 Введите категорию для поиска лучшего кэшбэка.

Примеры:
• Такси
• Супермаркеты
• Фастфуд
• Рестораны

Или /cancel для отмены.`

	b.sendText(message.Chat.ID, text)
}

// handleList обрабатывает команду /list с поддержкой пагинации.
// Форматы:
// /list - последние 5 строк
// /list all - все строки
// /list 1-10 - строки с 1 по 10
// /list 1-5,8,10 - строки с 1 по 5, а также 8 и 10
func (b *Bot) handleList(message *tgbotapi.Message) {
	userIDStr := strconv.FormatInt(message.From.ID, 10)
	groupName, err := b.client.GetUserGroup(userIDStr)
	if err != nil {
		b.sendText(message.Chat.ID, "❌ Вы должны быть в группе. Используйте /creategroup или /joingroup")
		return
	}

	// Парсим аргументы команды
	args := strings.TrimPrefix(message.Text, "/list")
	args = strings.TrimSpace(args)
	
	indices, showAll, err := ParseListArguments(args)
	if err != nil {
		b.sendText(message.Chat.ID, fmt.Sprintf("❌ Неверный формат: %s\n\n"+
			"Примеры:\n"+
			"• /list - последние 5\n"+
			"• /list all - все\n"+
			"• /list 1-10 - с 1 по 10\n"+
			"• /list 1-5,8,10 - с 1 по 5, а также 8 и 10", err))
		return
	}

	// Получаем все записи группы
	list, err := b.client.ListCashback(groupName, 1000, 0)
	if err != nil {
		b.sendText(message.Chat.ID, fmt.Sprintf("❌ Ошибка: %s", err))
		return
	}

	// Фильтруем записи по индексам
	var filtered []models.CashbackRule
	if showAll {
		filtered = list.Rules
	} else if indices == nil {
		// По умолчанию - последние 5 (т.к. list.Rules уже отсортирован по created_at DESC)
		limit := 5
		if len(list.Rules) < limit {
			limit = len(list.Rules)
		}
		filtered = list.Rules[:limit]
	} else {
		// Выбираем по индексам
		for _, idx := range indices {
			if idx > 0 && idx <= len(list.Rules) {
				filtered = append(filtered, list.Rules[idx-1])
			}
		}
	}

	if len(filtered) == 0 {
		b.sendText(message.Chat.ID, "📝 Нет записей для отображения.")
		return
	}

	b.sendTextPlain(message.Chat.ID, formatCashbackListTable(filtered, list.Total, showAll, indices))
}

// handleUpdateCommand обрабатывает команду /update ID.
func (b *Bot) handleUpdateCommand(message *tgbotapi.Message) {
	args := strings.Fields(message.Text)
	if len(args) < 2 {
		// Устанавливаем состояние ожидания ID
		b.setState(message.From.ID, StateAwaitingUpdateID, nil, nil, 0)
		b.sendText(message.Chat.ID, "🔢 Введите ID кешбека для обновления.\n\nИспользуйте /list для просмотра всех ID.\n\nИли /cancel для отмены.")
		return
	}

	id, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		b.sendText(message.Chat.ID, "❌ Неверный формат ID. Используйте число.")
		return
	}

	rule, err := b.client.GetCashbackByID(id)
	if err != nil {
		b.sendText(message.Chat.ID, fmt.Sprintf("❌ %% кешбек с ID %d не найден.", id))
		return
	}

	// Проверяем владельца
	if rule.UserID != strconv.FormatInt(message.From.ID, 10) {
		b.sendText(message.Chat.ID, "❌ Вы можете обновлять только свой %% кешбек.")
		return
	}

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
	
	b.setState(message.From.ID, StateAwaitingUpdateData, nil, nil, id)
}

// handleDeleteCommand обрабатывает команду /delete ID.
func (b *Bot) handleDeleteCommand(message *tgbotapi.Message) {
	args := strings.Fields(message.Text)
	if len(args) < 2 {
		// Устанавливаем состояние ожидания ID
		b.setState(message.From.ID, StateAwaitingDeleteID, nil, nil, 0)
		b.sendText(message.Chat.ID, "🔢 Введите ID кешбека для удаления.\n\nИспользуйте /list для просмотра всех ID.\n\nИли /cancel для отмены.")
		return
	}

	id, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		b.sendText(message.Chat.ID, "❌ Неверный формат ID. Используйте число.")
		return
	}

	rule, err := b.client.GetCashbackByID(id)
	if err != nil {
		b.sendText(message.Chat.ID, fmt.Sprintf("❌ %% кешбек с ID %d не найден.", id))
		return
	}

	// Проверяем владельца
	if rule.UserID != strconv.FormatInt(message.From.ID, 10) {
		b.sendText(message.Chat.ID, "❌ Вы можете удалять только свой %% кешбек.")
		return
	}

	b.sendWithButtons(message.Chat.ID, formatDeletePrompt(rule), ButtonsDelete)
	b.setState(message.From.ID, StateAwaitingDeleteConfirm, nil, nil, id)
}

// handleCancel обрабатывает команду /cancel.
func (b *Bot) handleCancel(message *tgbotapi.Message) {
	b.clearState(message.From.ID)
	b.sendText(message.Chat.ID, "🚫 Операция отменена")
}

// handleBankInfo обрабатывает команду /bankinfo bank_name.
func (b *Bot) handleBankInfo(message *tgbotapi.Message) {
	args := strings.TrimPrefix(message.Text, "/bankinfo")
	args = strings.TrimSpace(args)

	if args == "" {
		// Устанавливаем состояние ожидания названия банка
		b.setState(message.From.ID, StateAwaitingBankInfoName, nil, nil, 0)
		b.sendText(message.Chat.ID, "🏦 Введите название банка.\n\nПримеры:\n• Тинькофф\n• Сбер\n• Альфа-Банк\n\nИли /cancel для отмены.")
		return
	}

	userIDStr := strconv.FormatInt(message.From.ID, 10)
	groupName, err := b.client.GetUserGroup(userIDStr)
	if err != nil {
		b.sendText(message.Chat.ID, "❌ Вы должны быть в группе. Используйте /creategroup или /joingroup")
		return
	}

	// Попытка найти похожий банк
	correctedBank, found := FindSimilarBank(args)
	bankToSearch := args
	if found && correctedBank != args {
		bankToSearch = correctedBank
	}

	rules, err := b.client.GetCashbackByBank(groupName, bankToSearch)
	if err != nil {
		b.sendText(message.Chat.ID, fmt.Sprintf("❌ Кэшбэки для банка \"%s\" не найдены.\n\n"+
			"💡 Используйте /banklist для просмотра доступных банков.", args))
		return
	}

	b.sendText(message.Chat.ID, formatBankInfo(bankToSearch, rules))
}

// handleCategoryList обрабатывает команду /categorylist.
func (b *Bot) handleCategoryList(message *tgbotapi.Message) {
	userIDStr := strconv.FormatInt(message.From.ID, 10)
	groupName, err := b.client.GetUserGroup(userIDStr)
	if err != nil {
		b.sendText(message.Chat.ID, "❌ Вы должны быть в группе. Используйте /creategroup или /joingroup")
		return
	}

	categories, err := b.client.GetActiveCategories(groupName)
	if err != nil {
		b.sendText(message.Chat.ID, "❌ Ошибка получения категорий")
		return
	}

	if len(categories) == 0 {
		b.sendText(message.Chat.ID, "📝 Пока нет активных категорий в группе.")
		return
	}

	b.sendText(message.Chat.ID, formatCategoryList(categories))
}

// handleBankList обрабатывает команду /banklist.
func (b *Bot) handleBankList(message *tgbotapi.Message) {
	userIDStr := strconv.FormatInt(message.From.ID, 10)
	groupName, err := b.client.GetUserGroup(userIDStr)
	if err != nil {
		b.sendText(message.Chat.ID, "❌ Вы должны быть в группе. Используйте /creategroup или /joingroup")
		return
	}

	banks, err := b.client.GetActiveBanks(groupName)
	if err != nil {
		b.sendText(message.Chat.ID, "❌ Ошибка получения банков")
		return
	}

	if len(banks) == 0 {
		b.sendText(message.Chat.ID, "📝 Пока нет активных банков в группе.")
		return
	}

	b.sendText(message.Chat.ID, formatBankList(banks))
}

// handleUserInfo обрабатывает команду /userinfo [ID].
func (b *Bot) handleUserInfo(message *tgbotapi.Message) {
	userIDStr := strconv.FormatInt(message.From.ID, 10)
	groupName, err := b.client.GetUserGroup(userIDStr)
	if err != nil {
		b.sendText(message.Chat.ID, "❌ Вы должны быть в группе. Используйте /creategroup или /joingroup")
		return
	}

	// Парсим аргументы
	args := strings.TrimPrefix(message.Text, "/userinfo")
	args = strings.TrimSpace(args)

	targetUserID := userIDStr
	if args != "" {
		// Указан ID другого пользователя
		targetUserID = args
	}

	// Получаем все кэшбэки группы
	list, err := b.client.ListCashback(groupName, 1000, 0)
	if err != nil {
		b.sendText(message.Chat.ID, "❌ Ошибка получения данных")
		return
	}

	// Фильтруем по пользователю
	var userRules []models.CashbackRule
	for _, rule := range list.Rules {
		if rule.UserID == targetUserID {
			userRules = append(userRules, rule)
		}
	}

	if len(userRules) == 0 {
		if targetUserID == userIDStr {
			b.sendText(message.Chat.ID, "📝 У вас пока нет кэшбэков.")
		} else {
			b.sendText(message.Chat.ID, fmt.Sprintf("📝 У пользователя %s пока нет кэшбэков.", targetUserID))
		}
		return
	}

	b.sendText(message.Chat.ID, formatUserInfo(userRules))
}

// handleUserList обрабатывает команду /userlist [a-b,c|all].
func (b *Bot) handleUserList(message *tgbotapi.Message) {
	userIDStr := strconv.FormatInt(message.From.ID, 10)
	groupName, err := b.client.GetUserGroup(userIDStr)
	if err != nil {
		b.sendText(message.Chat.ID, "❌ Вы должны быть в группе. Используйте /creategroup или /joingroup")
		return
	}

	// Парсим аргументы команды
	args := strings.TrimPrefix(message.Text, "/userlist")
	args = strings.TrimSpace(args)

	// Получаем список пользователей
	users, err := b.client.GetGroupUsers(groupName)
	if err != nil {
		b.sendText(message.Chat.ID, "❌ Ошибка получения пользователей")
		return
	}

	if len(users) == 0 {
		b.sendText(message.Chat.ID, "📝 Нет пользователей в группе.")
		return
	}

	// Парсим аргументы для пагинации
	indices, showAll, err := ParseListArguments(args)
	if err != nil {
		b.sendText(message.Chat.ID, fmt.Sprintf("❌ Неверный формат: %s\n\n"+
			"Примеры:\n"+
			"• /userlist - все пользователи\n"+
			"• /userlist all - все\n"+
			"• /userlist 1-5 - с 1 по 5\n"+
			"• /userlist 1,3,5 - 1, 3 и 5", err))
		return
	}

	// Фильтруем пользователей по индексам
	var filtered []models.UserInfo
	if showAll || args == "" {
		filtered = users
	} else if indices != nil {
		// Выбираем по индексам
		for _, idx := range indices {
			if idx > 0 && idx <= len(users) {
				filtered = append(filtered, users[idx-1])
			}
		}
	}

	if len(filtered) == 0 {
		b.sendText(message.Chat.ID, "📝 Нет пользователей для отображения.")
		return
	}

	b.sendTextPlain(message.Chat.ID, formatUserListTable(filtered, len(users)))
}

