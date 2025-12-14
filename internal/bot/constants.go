// Package bot содержит константы и конфигурацию бота.
package bot

import "time"

// Лимиты и таймауты.
const (
	// DefaultListLimit — лимит по умолчанию для списков.
	DefaultListLimit = 100

	// MaxListLimit — максимальный лимит для списков.
	MaxListLimit = 1000

	// UpdateTimeout — таймаут для получения обновлений от Telegram.
	UpdateTimeout = 60

	// HTTPClientTimeout — таймаут HTTP клиента.
	HTTPClientTimeout = 30 * time.Second
)

// Пороги для fuzzy matching.
const (
	// SimilarityThresholdHigh — высокий порог похожести (уверенное совпадение).
	SimilarityThresholdHigh = 60.0

	// SimilarityThresholdLow — низкий порог похожести (слабое совпадение).
	SimilarityThresholdLow = 40.0

	// BankSimilarityThreshold — порог похожести для банков.
	BankSimilarityThreshold = 60.0
)

// Эмодзи для сообщений.
const (
	EmojiSuccess     = "✅"
	EmojiError       = "❌"
	EmojiWarning     = "⚠️"
	EmojiInfo        = "ℹ️"
	EmojiQuestion    = "❓"
	EmojiSearch      = "🔍"
	EmojiSave        = "💾"
	EmojiBank        = "🏦"
	EmojiCategory    = "📁"
	EmojiCalendar    = "📅"
	EmojiPercent     = "💰"
	EmojiAmount      = "💵"
	EmojiUser        = "👤"
	EmojiGroup       = "👥"
	EmojiList        = "📋"
	EmojiTrophy      = "🏆"
	EmojiID          = "🆔"
	EmojiBulb        = "💡"
	EmojiCancel      = "🚫"
	EmojiRobot       = "🤖"
	EmojiRocket      = "🚀"
	EmojiMessage     = "📨"
	EmojiHello       = "👋"
	EmojiTarget      = "🎯"
	EmojiPencil      = "✍️"
	EmojiBook        = "📖"
	EmojiCard        = "💳"
	EmojiBlueCircle  = "🔹"
	EmojiStar        = "✨"
	EmojiHandshake   = "🤝"
	EmojiChart       = "📊"
)

// Текстовые шаблоны ошибок.
const (
	ErrMsgUnknownCommand   = "❌ Неизвестная команда. Используйте /help для справки."
	ErrMsgNotInGroup       = "⚠️ Вы не состоите в группе!\n\nСначала создайте группу или присоединитесь к существующей:\n/creategroup название - создать новую группу\n/joingroup название - присоединиться к группе"
	ErrMsgMustBeInGroup    = "❌ Вы должны быть в группе. Используйте /creategroup или /joingroup"
	ErrMsgSpecifyGroupName = "❌ Укажите название группы.\n\nПример: /creategroup Семья"
	ErrMsgGroupNotExists   = "❌ Группа \"%s\" не существует"
	ErrMsgAlreadyInGroup   = "⚠️ Вы уже состоите в группе \"%s\""
	ErrMsgSpecifyID        = "❌ Укажите ID %% кешбека.\n\nПример: /%s 5"
	ErrMsgInvalidID        = "❌ Неверный формат ID. Используйте число."
	ErrMsgRuleNotFound     = "❌ %% кешбек с ID %d не найден."
	ErrMsgNotYourRule      = "❌ Вы можете %s только свой %% кешбек."
	ErrMsgParseError       = "❌ Ошибка парсинга: %s"
	ErrMsgAPIError         = "❌ Ошибка: %s"
	ErrMsgValidationError  = "❌ Ошибки валидации:\n%s"
	ErrMsgMissingData      = "⚠️ Не хватает данных:\n%s\n\nФормат: Банк, Категория, Процент, Сумма[, Месяц]\nПример: \"Тинькофф, Такси, 5%%, 3000\" (месяц опционален)"
	ErrMsgSpecifyCategory  = "❌ Укажите категорию. Например: \"Такси\""
)

// Текстовые шаблоны успешных сообщений.
const (
	MsgOperationCancelled = "🚫 Операция отменена"
	MsgGroupCreated       = "✅ Группа \"%s\" успешно создана!\n\nВы можете пригласить друзей командой:\n/joingroup %s"
	MsgGroupJoined        = "✅ Вы присоединились к группе \"%s\"!"
	MsgRuleDeleted        = "✅ %% кешбек ID %d успешно удалён!"
	MsgDeleteCancelled    = "❌ Удаление отменено."
	MsgChooseOption       = "❓ Пожалуйста, выберите один из вариантов"
	MsgCheckingData       = "🔍 Проверяю данные..."
	MsgSearching          = "🔍 Ищу лучший кэшбэк для \"%s\" в группе \"%s\"..."
	MsgNoGroupsYet        = "📝 Пока нет групп.\n\nСоздайте первую группу: /creategroup Название"
	MsgNoCashbackYet      = "📝 Пока нет % кешбека в группе.\n\nДобавьте первым!"
	MsgKeepAsIs           = "Хорошо, оставляю как есть."
	MsgSendAgain          = "Отправьте данные заново, если хотите продолжить."
	MsgTryDifferentName   = "Хорошо, попробуйте ввести название категории по-другому."
)

// Форматы дат.
const (
	DateFormatYearMonth  = "2006-01"
	DateFormatMonthYear  = "01/2006"
	DateFormatDisplay    = "01/2006"
)

// API endpoints (относительные пути).
const (
	EndpointCashback       = "/api/v1/cashback"
	EndpointCashbackSuggest = "/api/v1/cashback/suggest"
	EndpointCashbackBest   = "/api/v1/cashback/best"
	EndpointGroups         = "/api/v1/groups"
	EndpointGroupsCheck    = "/api/v1/groups/check"
	EndpointGroupsMembers  = "/api/v1/groups/members"
	EndpointUserGroup      = "/api/v1/users/%s/group"
)

