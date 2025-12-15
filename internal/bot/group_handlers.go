package bot

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rymax1e/open-cashback-advisor/internal/models"
)

// handleCreateGroup обрабатывает команду /creategroup.
func (b *Bot) handleCreateGroup(message *tgbotapi.Message) {
	args := strings.Fields(message.Text)
	if len(args) < 2 {
		// Устанавливаем состояние ожидания названия группы
		b.setState(message.From.ID, StateAwaitingCreateGroupName, nil, nil, 0)
		b.sendText(message.Chat.ID, "👥 Создание новой группы\n\n"+
			"💬 Введите название группы (без команды)\n\n"+
			"Примеры:\n"+
			"• Семья\n"+
			"• Работа\n"+
			"• Друзья\n\n"+
			"🔒 Совет: Используйте уникальное название, чтобы посторонние не могли случайно присоединиться к вашей группе.\n\n"+
			"Или /cancel для отмены.")
		return
	}

	groupName := strings.Join(args[1:], " ")
	userIDStr := strconv.FormatInt(message.From.ID, 10)

	// Проверяем, не состоит ли пользователь уже в группе
	if currentGroup, err := b.client.GetUserGroup(userIDStr); err == nil {
		b.sendText(message.Chat.ID, fmt.Sprintf("⚠️ Вы уже состоите в группе \"%s\"", currentGroup))
		return
	}

	// Создаём группу
	err := b.client.CreateGroup(groupName, userIDStr)
	if err != nil {
		b.sendText(message.Chat.ID, fmt.Sprintf("❌ %s", err))
		return
	}

	b.sendText(message.Chat.ID, fmt.Sprintf(
		"✅ Группа \"%s\" успешно создана!\n\nВы можете пригласить друзей командой:\n/joingroup %s",
		groupName, groupName,
	))
}

// handleJoinGroup обрабатывает команду /joingroup.
func (b *Bot) handleJoinGroup(message *tgbotapi.Message) {
	args := strings.Fields(message.Text)
	userIDStr := strconv.FormatInt(message.From.ID, 10)

	if len(args) < 2 {
		// Устанавливаем состояние ожидания названия группы
		b.setState(message.From.ID, StateAwaitingJoinGroupName, nil, nil, 0)
		b.sendText(message.Chat.ID, "👥 Присоединение к группе\n\n"+
			"💬 Введите название группы (без команды)\n\n"+
			"Примеры:\n"+
			"• Семья\n"+
			"• Работа\n"+
			"• Друзья\n\n"+
			"⚠️ Группа должна существовать\n"+
			"Или /cancel для отмены.")
		return
	}

	groupName := strings.Join(args[1:], " ")

	// Проверяем существование группы
	if !b.client.GroupExists(groupName) {
		b.sendText(message.Chat.ID, fmt.Sprintf(
			"❌ Группа \"%s\" не существует.\n\nСоздайте её: /creategroup %s",
			groupName, groupName,
		))
		return
	}

	// Проверяем текущую группу пользователя
	if currentGroup, err := b.client.GetUserGroup(userIDStr); err == nil {
		if currentGroup == groupName {
			b.sendText(message.Chat.ID, fmt.Sprintf("⚠️ Вы уже состоите в группе \"%s\"", currentGroup))
			return
		}
		log.Printf("👥 Пользователь @%s переключается из группы \"%s\" в группу \"%s\"",
			message.From.UserName, currentGroup, groupName)
	}

	// Присоединяемся к группе
	err := b.client.JoinGroup(userIDStr, groupName)
	if err != nil {
		b.sendText(message.Chat.ID, fmt.Sprintf("❌ Ошибка: %s", err))
		return
	}

	b.sendText(message.Chat.ID, fmt.Sprintf("✅ Вы присоединились к группе \"%s\"!", groupName))
}


// handleGroupInfo обрабатывает команду /groupinfo.
func (b *Bot) handleGroupInfo(message *tgbotapi.Message) {
	args := strings.Fields(message.Text)
	userIDStr := strconv.FormatInt(message.From.ID, 10)

	var groupName string
	if len(args) < 2 {
		var err error
		groupName, err = b.client.GetUserGroup(userIDStr)
		if err != nil {
			b.sendText(message.Chat.ID, "❌ Вы не состоите в группе")
			return
		}
	} else {
		groupName = strings.Join(args[1:], " ")
		if !b.client.GroupExists(groupName) {
			b.sendText(message.Chat.ID, fmt.Sprintf("❌ Группа \"%s\" не существует", groupName))
			return
		}
	}

	// Получаем информацию о пользователях
	users, err := b.client.GetGroupUsers(groupName)
	if err != nil {
		b.sendText(message.Chat.ID, "❌ Ошибка получения участников")
		return
	}

	// Получаем все кешбеки группы для подсчета активности
	list, err := b.client.ListCashback(groupName, 1000, 0)
	if err != nil {
		b.sendText(message.Chat.ID, "❌ Ошибка получения данных")
		return
	}

	text := b.formatGroupInfo(groupName, users, list.Rules)
	b.sendText(message.Chat.ID, text)
}

// formatGroupInfo форматирует информацию о группе.
func (b *Bot) formatGroupInfo(groupName string, users []models.UserInfo, rules []models.CashbackRule) string {
	text := fmt.Sprintf("📊 Информация о группе\n\n")
	text += fmt.Sprintf("👥 Группа: <b>%s</b>\n", groupName)
	text += fmt.Sprintf("📌 Участников: %d\n", len(users))
	text += fmt.Sprintf("💳 Всего кешбеков: %d\n\n", len(rules))

	if len(users) == 0 {
		text += "📝 Пока нет участников в группе."
		return text
	}

	// Подсчитываем статистику по каждому пользователю
	userStats := make(map[string]struct {
		Name          string
		TotalRules    int
		ActiveRules   int
		LastAddedDate time.Time
	})

	now := time.Now()
	for _, rule := range rules {
		stats := userStats[rule.UserID]
		stats.Name = rule.UserDisplayName
		stats.TotalRules++
		
		// Считаем активные (не истекшие) кешбеки
		if rule.MonthYear.After(now.AddDate(0, 0, -1)) {
			stats.ActiveRules++
		}
		
		// Отслеживаем последнюю дату добавления
		if rule.CreatedAt.After(stats.LastAddedDate) {
			stats.LastAddedDate = rule.CreatedAt
		}
		
		userStats[rule.UserID] = stats
	}

	// Формируем список участников
	text += "👤 Участники:\n\n"
	
	for i, user := range users {
		stats := userStats[user.UserID]
		text += fmt.Sprintf("%d. <b>%s</b>\n", i+1, user.UserDisplayName)
		
		if stats.TotalRules > 0 {
			text += fmt.Sprintf("   💳 Кешбеков: %d (активных: %d)\n", stats.TotalRules, stats.ActiveRules)
			
			if !stats.LastAddedDate.IsZero() {
				// Форматируем дату последней активности
				daysSince := int(now.Sub(stats.LastAddedDate).Hours() / 24)
				if daysSince == 0 {
					text += "   📅 Последняя активность: сегодня\n"
				} else if daysSince == 1 {
					text += "   📅 Последняя активность: вчера\n"
				} else if daysSince < 7 {
					text += fmt.Sprintf("   📅 Последняя активность: %d дн. назад\n", daysSince)
				} else {
					text += fmt.Sprintf("   📅 Последняя активность: %s\n", stats.LastAddedDate.Format("02.01.2006"))
				}
			}
		} else {
			text += "   📝 Еще не добавлял кешбеки\n"
		}
		
		text += "\n"
	}

	return text
}

// handleGroupNameInput обрабатывает ввод названия группы.
func (b *Bot) handleGroupNameInput(message *tgbotapi.Message) {
	groupName := strings.TrimSpace(message.Text)
	userIDStr := strconv.FormatInt(message.From.ID, 10)

	err := b.client.CreateGroup(groupName, userIDStr)
	if err != nil {
		b.sendText(message.Chat.ID, fmt.Sprintf("❌ %s", err))
		b.clearState(message.From.ID)
		return
	}

	b.sendText(message.Chat.ID, fmt.Sprintf("✅ Группа \"%s\" создана!", groupName))
	b.clearState(message.From.ID)
}

