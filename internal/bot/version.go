package bot

// Version версия бота
// Обновляйте при каждом значимом изменении
const Version = "1.0.3"

// BuildInfo возвращает информацию о версии
func BuildInfo() string {
	return "v" + Version
}

