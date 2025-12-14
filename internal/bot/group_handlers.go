package bot

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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
		b.showAvailableGroups(message.Chat.ID)
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

// showAvailableGroups показывает список доступных групп.
func (b *Bot) showAvailableGroups(chatID int64) {
	groups, err := b.client.GetAllGroups()
	if err != nil {
		b.sendText(chatID, "❌ Ошибка получения списка групп")
		return
	}

	if len(groups) == 0 {
		b.sendText(chatID, "📝 Пока нет групп.\n\n💡 Создайте первую группу: /creategroup Название\n\nИли /cancel для отмены.")
		return
	}

	text := "👥 К какой группе присоединиться?\n\n"
	text += "📋 Доступные группы:\n"
	for i, group := range groups {
		text += fmt.Sprintf("• %s\n", group)
	}
	text += "\n💬 Введите название группы (без команды)\n"
	text += "Или /cancel для отмены."

	b.sendText(chatID, text)
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

	members, err := b.client.GetGroupMembers(groupName)
	if err != nil {
		b.sendText(message.Chat.ID, "❌ Ошибка получения участников")
		return
	}

	text := b.formatGroupInfo(groupName, members)
	b.sendText(message.Chat.ID, text)
}

// formatGroupInfo форматирует информацию о группе.
func (b *Bot) formatGroupInfo(groupName string, members []string) string {
	text := fmt.Sprintf("📊 Группа: %s\n\n", groupName)
	text += fmt.Sprintf("👥 Участников: %d\n\n", len(members))

	// Получаем кэшбэки текущего месяца
	now := time.Now()
	monthYear := fmt.Sprintf("%d-%02d", now.Year(), now.Month())

	log.Printf("🔍 /groupinfo debug: groupName=%s, monthYear=%s", groupName, monthYear)

	list, err := b.client.ListCashback(groupName, 1000, 0)
	if err != nil {
		log.Printf("❌ ListCashback error: %v", err)
		return text
	}

	log.Printf("✅ ListCashback returned %d rules", len(list.Rules))

	if len(list.Rules) == 0 {
		return text
	}

	// Группируем по категориям
	categories := make(map[string][]string)
	matchCount := 0

	for _, rule := range list.Rules {
		ruleMonth := rule.MonthYear.Format("2006-01")
		log.Printf("  📅 Rule ID=%d, category=%s, month=%s (checking against %s)",
			rule.ID, rule.Category, ruleMonth, monthYear)

		if ruleMonth == monthYear {
			matchCount++
			info := fmt.Sprintf("%.1f%% (%s, карта: %s)", rule.CashbackPercent, rule.BankName, rule.UserDisplayName)
			categories[rule.Category] = append(categories[rule.Category], info)
		}
	}

	log.Printf("✅ Matched %d rules for month %s, categories: %d", matchCount, monthYear, len(categories))

	if len(categories) > 0 {
		text += "💰 Кэшбэк в текущем месяце:\n\n"
		for category, infos := range categories {
			text += fmt.Sprintf("📁 %s:\n", category)
			for _, info := range infos {
				text += fmt.Sprintf("   • %s\n", info)
			}
			text += "\n"
		}
	} else {
		text += "💡 Пока нет кэшбэков в текущем месяце"
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

