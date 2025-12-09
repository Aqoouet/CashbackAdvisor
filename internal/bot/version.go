package bot

// Version версия бота
// Обновляйте при каждом значимом изменении
const Version = "1.2.1"

// BuildInfo возвращает информацию о версии
func BuildInfo() string {
	return "v" + Version
}

