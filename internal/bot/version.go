package bot

// Version версия бота
// Обновляйте при каждом значимом изменении
const Version = "2.1.0"

// BuildInfo возвращает информацию о версии
func BuildInfo() string {
	return "v" + Version
}

