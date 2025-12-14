package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rymax1e/open-cashback-advisor/internal/models"
)

// Repository представляет репозиторий для работы с правилами кэшбэка.
type Repository struct {
	db *Database
}

// NewRepository создаёт новый репозиторий.
func NewRepository(db *Database) *Repository {
	return &Repository{db: db}
}

// --- Методы для работы с кэшбэком ---

// Create создаёт новое правило кэшбэка.
func (r *Repository) Create(ctx context.Context, rule *models.CashbackRule) error {
	err := r.db.Pool.QueryRow(
		ctx, QueryCreateCashback,
		rule.GroupName, rule.Category, rule.BankName, rule.UserID,
		rule.UserDisplayName, rule.MonthYear, rule.CashbackPercent, rule.MaxAmount,
	).Scan(&rule.ID, &rule.CreatedAt, &rule.UpdatedAt)

	if err != nil {
		return fmt.Errorf("создание правила: %w", err)
	}
	return nil
}

// GetByID получает правило по ID.
func (r *Repository) GetByID(ctx context.Context, id int64) (*models.CashbackRule, error) {
	rule, err := r.scanCashbackRule(ctx, QueryGetCashbackByID, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("правило с ID %d: %w", id, ErrNotFound)
		}
		return nil, fmt.Errorf("получение правила %d: %w", id, err)
	}
	return rule, nil
}

// Update обновляет правило кэшбэка.
func (r *Repository) Update(ctx context.Context, id int64, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return ErrEmptyUpdate
	}

	query, args := r.buildUpdateQuery(id, updates)

	result, err := r.db.Pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("обновление правила %d: %w", id, err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("правило с ID %d: %w", id, ErrNotFound)
	}

	return nil
}

// Delete удаляет правило кэшбэка.
func (r *Repository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.Pool.Exec(ctx, QueryDeleteCashback, id)
	if err != nil {
		return fmt.Errorf("удаление правила %d: %w", id, err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("правило с ID %d: %w", id, ErrNotFound)
	}

	return nil
}

// List получает список правил с пагинацией.
func (r *Repository) List(ctx context.Context, limit, offset int, groupName string) ([]models.CashbackRule, int, error) {
	// Получаем общее количество
	var total int
	err := r.db.Pool.QueryRow(ctx, QueryCountCashbackByGroup, groupName).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("подсчёт правил: %w", err)
	}

	// Получаем правила
	rows, err := r.db.Pool.Query(ctx, QueryListCashbackByGroup, groupName, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("получение списка правил: %w", err)
	}
	defer rows.Close()

	rules, err := r.scanCashbackRules(rows)
	if err != nil {
		return nil, 0, err
	}

	return rules, total, nil
}

// GetBestCashback получает правило с лучшим кэшбэком.
func (r *Repository) GetBestCashback(ctx context.Context, groupName, category string, monthYear time.Time) (*models.CashbackRule, error) {
	rule, err := r.scanCashbackRule(ctx, QueryGetBestCashback, groupName, category, monthYear)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("правила для '%s' в %s: %w", category, monthYear.Format("2006-01"), ErrNotFound)
		}
		return nil, fmt.Errorf("получение лучшего кэшбэка: %w", err)
	}
	return rule, nil
}

// --- Методы для fuzzy поиска ---

// FuzzySearchGroupName выполняет fuzzy-поиск по названию группы.
func (r *Repository) FuzzySearchGroupName(ctx context.Context, value string, threshold float64, limit int) ([]models.FuzzySuggestion, error) {
	return r.fuzzySearch(ctx, FieldGroupName, value, threshold, limit)
}

// FuzzySearchCategory выполняет fuzzy-поиск по категории.
func (r *Repository) FuzzySearchCategory(ctx context.Context, value string, threshold float64, limit int) ([]models.FuzzySuggestion, error) {
	return r.fuzzySearch(ctx, FieldCategory, value, threshold, limit)
}

// FuzzySearchBankName выполняет fuzzy-поиск по названию банка.
func (r *Repository) FuzzySearchBankName(ctx context.Context, value string, threshold float64, limit int) ([]models.FuzzySuggestion, error) {
	return r.fuzzySearch(ctx, FieldBankName, value, threshold, limit)
}

// FuzzySearchUserDisplayName выполняет fuzzy-поиск по имени пользователя.
func (r *Repository) FuzzySearchUserDisplayName(ctx context.Context, value string, threshold float64, limit int) ([]models.FuzzySuggestion, error) {
	return r.fuzzySearch(ctx, FieldUserDisplayName, value, threshold, limit)
}

// fuzzySearch выполняет fuzzy-поиск по указанному полю.
func (r *Repository) fuzzySearch(ctx context.Context, field, value string, threshold float64, limit int) ([]models.FuzzySuggestion, error) {
	query := fmt.Sprintf(QueryFuzzySearchTemplate, field, field, field)

	rows, err := r.db.Pool.Query(ctx, query, value, threshold, limit)
	if err != nil {
		return nil, fmt.Errorf("fuzzy-поиск по %s: %w", field, err)
	}
	defer rows.Close()

	var suggestions []models.FuzzySuggestion
	for rows.Next() {
		var s models.FuzzySuggestion
		if err := rows.Scan(&s.Value, &s.Similarity); err != nil {
			return nil, fmt.Errorf("чтение результата fuzzy-поиска: %w", err)
		}
		suggestions = append(suggestions, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("итерация результатов: %w", err)
	}

	return suggestions, nil
}

// --- Методы для работы с группами ---

// SetUserGroup устанавливает группу пользователя.
func (r *Repository) SetUserGroup(ctx context.Context, userID, groupName string) error {
	_, err := r.db.Pool.Exec(ctx, QuerySetUserGroup, userID, groupName)
	if err != nil {
		return fmt.Errorf("установка группы пользователя: %w", err)
	}
	return nil
}

// GetUserGroup получает группу пользователя.
func (r *Repository) GetUserGroup(ctx context.Context, userID string) (string, error) {
	var groupName string
	err := r.db.Pool.QueryRow(ctx, QueryGetUserGroup, userID).Scan(&groupName)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("пользователь %s: %w", userID, ErrNotFound)
		}
		return "", fmt.Errorf("получение группы пользователя: %w", err)
	}
	return groupName, nil
}

// CreateGroup создаёт новую группу.
func (r *Repository) CreateGroup(ctx context.Context, groupName, creatorID string) error {
	_, err := r.db.Pool.Exec(ctx, QueryCreateGroup, groupName, creatorID)
	if err != nil {
		return fmt.Errorf("создание группы: %w", err)
	}

	// Добавляем создателя в группу
	return r.SetUserGroup(ctx, creatorID, groupName)
}

// GroupExists проверяет существование группы.
func (r *Repository) GroupExists(ctx context.Context, groupName string) (bool, error) {
	var exists bool
	err := r.db.Pool.QueryRow(ctx, QueryGroupExists, groupName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("проверка существования группы: %w", err)
	}
	return exists, nil
}

// GetGroupMembers возвращает участников группы.
func (r *Repository) GetGroupMembers(ctx context.Context, groupName string) ([]string, error) {
	rows, err := r.db.Pool.Query(ctx, QueryGetGroupMembers, groupName)
	if err != nil {
		return nil, fmt.Errorf("получение участников группы: %w", err)
	}
	defer rows.Close()

	var members []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("чтение участника: %w", err)
		}
		members = append(members, userID)
	}

	return members, nil
}

// GetAllGroups возвращает список всех групп.
func (r *Repository) GetAllGroups(ctx context.Context) ([]string, error) {
	rows, err := r.db.Pool.Query(ctx, QueryGetAllGroups)
	if err != nil {
		return nil, fmt.Errorf("получение списка групп: %w", err)
	}
	defer rows.Close()

	var groups []string
	for rows.Next() {
		var groupName string
		if err := rows.Scan(&groupName); err != nil {
			return nil, fmt.Errorf("чтение группы: %w", err)
		}
		groups = append(groups, groupName)
	}

	return groups, nil
}

// --- Вспомогательные методы ---

// scanCashbackRule сканирует одно правило из запроса.
func (r *Repository) scanCashbackRule(ctx context.Context, query string, args ...interface{}) (*models.CashbackRule, error) {
	var rule models.CashbackRule
	err := r.db.Pool.QueryRow(ctx, query, args...).Scan(
		&rule.ID, &rule.GroupName, &rule.Category, &rule.BankName,
		&rule.UserID, &rule.UserDisplayName, &rule.MonthYear,
		&rule.CashbackPercent, &rule.MaxAmount, &rule.CreatedAt, &rule.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// scanCashbackRules сканирует список правил из rows.
func (r *Repository) scanCashbackRules(rows pgx.Rows) ([]models.CashbackRule, error) {
	var rules []models.CashbackRule
	for rows.Next() {
		var rule models.CashbackRule
		err := rows.Scan(
			&rule.ID, &rule.GroupName, &rule.Category, &rule.BankName,
			&rule.UserID, &rule.UserDisplayName, &rule.MonthYear,
			&rule.CashbackPercent, &rule.MaxAmount, &rule.CreatedAt, &rule.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("чтение правила: %w", err)
		}
		rules = append(rules, rule)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("итерация результатов: %w", err)
	}

	return rules, nil
}

// buildUpdateQuery строит динамический UPDATE запрос.
func (r *Repository) buildUpdateQuery(id int64, updates map[string]interface{}) (string, []interface{}) {
	query := "UPDATE cashback_rules SET "
	args := make([]interface{}, 0, len(updates)+1)
	argPos := 1

	for field, value := range updates {
		if argPos > 1 {
			query += ", "
		}
		query += fmt.Sprintf("%s = $%d", field, argPos)
		args = append(args, value)
		argPos++
	}

	query += fmt.Sprintf(" WHERE id = $%d", argPos)
	args = append(args, id)

	return query, args
}
