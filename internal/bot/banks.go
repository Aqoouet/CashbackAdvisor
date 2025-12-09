package bot

// KnownBanks список известных банков для автоисправления
var KnownBanks = []string{
	"Тинькофф",
	"Сбер",
	"Сбербанк",
	"Альфа-Банк",
	"Альфа",
	"ВТБ",
	"Райффайзен",
	"Газпромбанк",
	"Открытие",
	"Росбанк",
	"МТС Банк",
	"Совкомбанк",
	"Ак Барс",
	"Уралсиб",
	"Промсвязьбанк",
	"Банк Санкт-Петербург",
	"Хоум Кредит",
	"Русский Стандарт",
	"Почта Банк",
}

// FindSimilarBank находит похожий банк из списка известных
func FindSimilarBank(input string) (string, bool) {
	if input == "" {
		return "", false
	}

	similar, distance := findSimilarCategory(input, KnownBanks)
	simPercent := similarity(input, similar)

	// Если похожесть > 60% - предлагаем исправление
	if simPercent > 60.0 {
		return similar, true
	}

	return "", false
}

