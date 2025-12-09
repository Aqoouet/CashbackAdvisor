package bot

// Version версия бота
// Обновляйте при каждом значимом изменении
const Version = "2.0.1"

// BuildInfo возвращает информацию о версии
func BuildInfo() string {
	return "v" + Version
}

