package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rymax1e/open-cashback-advisor/internal/models"
)

// Repository представляет репозиторий для работы с правилами кэшбэка
type Repository struct {
	db *Database
}

// NewRepository создает новый репозиторий
func NewRepository(db *Database) *Repository {
	return &Repository{db: db}
}

// Create создает новое правило кэшбэка
func (r *Repository) Create(ctx context.Context, rule *models.CashbackRule) error {
	query := `
		INSERT INTO cashback_rules (
			group_name, category, bank_name, user_id, user_display_name,
			month_year, cashback_percent, max_amount
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`

	err := r.db.Pool.QueryRow(
		ctx, query,
		rule.GroupName, rule.Category, rule.BankName, rule.UserID,
		rule.UserDisplayName, rule.MonthYear, rule.CashbackPercent, rule.MaxAmount,
	).Scan(&rule.ID, &rule.CreatedAt, &rule.UpdatedAt)

	if err != nil {
		return fmt.Errorf("не удалось создать правило: %w", err)
	}

	return nil
}

// GetByID получает правило по ID
func (r *Repository) GetByID(ctx context.Context, id int64) (*models.CashbackRule, error) {
	query := `
		SELECT id, group_name, category, bank_name, user_id, user_display_name,
			   month_year, cashback_percent, max_amount, created_at, updated_at
		FROM cashback_rules
		WHERE id = $1
	`

	var rule models.CashbackRule
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&rule.ID, &rule.GroupName, &rule.Category, &rule.BankName,
		&rule.UserID, &rule.UserDisplayName, &rule.MonthYear,
		&rule.CashbackPercent, &rule.MaxAmount, &rule.CreatedAt, &rule.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("правило с ID %d не найдено", id)
		}
		return nil, fmt.Errorf("не удалось получить правило: %w", err)
	}

	return &rule, nil
}

// Update обновляет правило кэшбэка
func (r *Repository) Update(ctx context.Context, id int64, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return fmt.Errorf("нет полей для обновления")
	}

	// Динамическое построение запроса
	query := "UPDATE cashback_rules SET "
	args := []interface{}{}
	argPosition := 1

	for field, value := range updates {
		if argPosition > 1 {
			query += ", "
		}
		query += fmt.Sprintf("%s = $%d", field, argPosition)
		args = append(args, value)
		argPosition++
	}

	query += fmt.Sprintf(" WHERE id = $%d", argPosition)
	args = append(args, id)

	result, err := r.db.Pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("не удалось обновить правило: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("правило с ID %d не найдено", id)
	}

	return nil
}

// Delete удаляет правило кэшбэка
func (r *Repository) Delete(ctx context.Context, id int64) error {
	query := "DELETE FROM cashback_rules WHERE id = $1"

	result, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("не удалось удалить правило: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("правило с ID %d не найдено", id)
	}

	return nil
}

// List получает список правил с пагинацией
func (r *Repository) List(ctx context.Context, limit, offset int, groupName string) ([]models.CashbackRule, int, error) {
	// Запрос на получение общего количества через JOIN с user_groups
	countQuery := `
		SELECT COUNT(*) 
		FROM cashback_rules cr
		INNER JOIN user_groups ug ON cr.user_id = ug.user_id
		WHERE ug.group_name = $1
	`
	
	var total int
	err := r.db.Pool.QueryRow(ctx, countQuery, groupName).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("не удалось получить количество записей: %w", err)
	}

	// Запрос на получение правил через JOIN с user_groups
	query := `
		SELECT cr.id, cr.group_name, cr.category, cr.bank_name, cr.user_id, cr.user_display_name,
			   cr.month_year, cr.cashback_percent, cr.max_amount, cr.created_at, cr.updated_at
		FROM cashback_rules cr
		INNER JOIN user_groups ug ON cr.user_id = ug.user_id
		WHERE ug.group_name = $1
		ORDER BY cr.created_at DESC 
		LIMIT $2 OFFSET $3
	`
	args := []interface{}{groupName, limit, offset}

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("не удалось получить список правил: %w", err)
	}
	defer rows.Close()

	var rules []models.CashbackRule
	for rows.Next() {
		var rule models.CashbackRule
		err := rows.Scan(
			&rule.ID, &rule.GroupName, &rule.Category, &rule.BankName,
			&rule.UserID, &rule.UserDisplayName, &rule.MonthYear,
			&rule.CashbackPercent, &rule.MaxAmount, &rule.CreatedAt, &rule.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("не удалось прочитать правило: %w", err)
		}
		rules = append(rules, rule)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("ошибка при чтении результатов: %w", err)
	}

	return rules, total, nil
}

// GetBestCashback получает правило с максимальным кэшбэком для заданных параметров
func (r *Repository) GetBestCashback(ctx context.Context, groupName, category string, monthYear time.Time) (*models.CashbackRule, error) {
	// Ищем лучший кешбэк среди всех пользователей группы через JOIN с user_groups
	query := `
		SELECT cr.id, cr.group_name, cr.category, cr.bank_name, cr.user_id, cr.user_display_name,
			   cr.month_year, cr.cashback_percent, cr.max_amount, cr.created_at, cr.updated_at
		FROM cashback_rules cr
		INNER JOIN user_groups ug ON cr.user_id = ug.user_id
		WHERE ug.group_name = $1 AND cr.category = $2 AND cr.month_year = $3
		ORDER BY cr.cashback_percent DESC, cr.max_amount DESC
		LIMIT 1
	`

	var rule models.CashbackRule
	err := r.db.Pool.QueryRow(ctx, query, groupName, category, monthYear).Scan(
		&rule.ID, &rule.GroupName, &rule.Category, &rule.BankName,
		&rule.UserID, &rule.UserDisplayName, &rule.MonthYear,
		&rule.CashbackPercent, &rule.MaxAmount, &rule.CreatedAt, &rule.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("правила для указанных параметров не найдены")
		}
		return nil, fmt.Errorf("не удалось получить лучший кэшбэк: %w", err)
	}

	return &rule, nil
}

// FuzzySearchGroupName выполняет fuzzy-поиск по названию группы
func (r *Repository) FuzzySearchGroupName(ctx context.Context, value string, threshold float64, limit int) ([]models.FuzzySuggestion, error) {
	return r.fuzzySearch(ctx, "group_name", value, threshold, limit)
}

// FuzzySearchCategory выполняет fuzzy-поиск по категории
func (r *Repository) FuzzySearchCategory(ctx context.Context, value string, threshold float64, limit int) ([]models.FuzzySuggestion, error) {
	return r.fuzzySearch(ctx, "category", value, threshold, limit)
}

// FuzzySearchBankName выполняет fuzzy-поиск по названию банка
func (r *Repository) FuzzySearchBankName(ctx context.Context, value string, threshold float64, limit int) ([]models.FuzzySuggestion, error) {
	return r.fuzzySearch(ctx, "bank_name", value, threshold, limit)
}

// FuzzySearchUserDisplayName выполняет fuzzy-поиск по отображаемому имени пользователя
func (r *Repository) FuzzySearchUserDisplayName(ctx context.Context, value string, threshold float64, limit int) ([]models.FuzzySuggestion, error) {
	return r.fuzzySearch(ctx, "user_display_name", value, threshold, limit)
}

// fuzzySearch выполняет fuzzy-поиск по указанному полю
func (r *Repository) fuzzySearch(ctx context.Context, field, value string, threshold float64, limit int) ([]models.FuzzySuggestion, error) {
	query := fmt.Sprintf(`
		SELECT DISTINCT %s, similarity(%s, $1) as sim
		FROM cashback_rules
		WHERE similarity(%s, $1) >= $2
		ORDER BY sim DESC
		LIMIT $3
	`, field, field, field)

	rows, err := r.db.Pool.Query(ctx, query, value, threshold, limit)
	if err != nil {
		return nil, fmt.Errorf("не удалось выполнить fuzzy-поиск по полю %s: %w", field, err)
	}
	defer rows.Close()

	var suggestions []models.FuzzySuggestion
	for rows.Next() {
		var suggestion models.FuzzySuggestion
		err := rows.Scan(&suggestion.Value, &suggestion.Similarity)
		if err != nil {
			return nil, fmt.Errorf("не удалось прочитать результат поиска: %w", err)
		}
		suggestions = append(suggestions, suggestion)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при чтении результатов: %w", err)
	}

	return suggestions, nil
}

// --- Методы для работы с группами пользователей ---

// SetUserGroup устанавливает группу пользователя
func (r *Repository) SetUserGroup(ctx context.Context, userID, groupName string) error {
	query := `
		INSERT INTO user_groups (user_id, group_name, updated_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id) 
		DO UPDATE SET group_name = $2, updated_at = CURRENT_TIMESTAMP
	`
	
	_, err := r.db.Pool.Exec(ctx, query, userID, groupName)
	if err != nil {
		return fmt.Errorf("не удалось установить группу: %w", err)
	}
	
	return nil
}

// GetUserGroup получает группу пользователя
func (r *Repository) GetUserGroup(ctx context.Context, userID string) (string, error) {
	query := `SELECT group_name FROM user_groups WHERE user_id = $1`
	
	var groupName string
	err := r.db.Pool.QueryRow(ctx, query, userID).Scan(&groupName)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("пользователь не в группе")
		}
		return "", fmt.Errorf("ошибка получения группы: %w", err)
	}
	
	return groupName, nil
}

// CreateGroup создает новую группу
func (r *Repository) CreateGroup(ctx context.Context, groupName, creatorID string) error {
	query := `
		INSERT INTO groups (group_name, created_by)
		VALUES ($1, $2)
		ON CONFLICT (group_name) DO NOTHING
	`
	
	_, err := r.db.Pool.Exec(ctx, query, groupName, creatorID)
	if err != nil {
		return fmt.Errorf("не удалось создать группу: %w", err)
	}
	
	// Добавляем создателя в группу
	return r.SetUserGroup(ctx, creatorID, groupName)
}

// GroupExists проверяет существование группы
func (r *Repository) GroupExists(ctx context.Context, groupName string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM groups WHERE group_name = $1)`
	
	var exists bool
	err := r.db.Pool.QueryRow(ctx, query, groupName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("ошибка проверки группы: %w", err)
	}
	
	return exists, nil
}

// GetGroupMembers возвращает участников группы
func (r *Repository) GetGroupMembers(ctx context.Context, groupName string) ([]string, error) {
	query := `SELECT user_id FROM user_groups WHERE group_name = $1`
	
	rows, err := r.db.Pool.Query(ctx, query, groupName)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения участников: %w", err)
	}
	defer rows.Close()
	
	var members []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		members = append(members, userID)
	}
	
	return members, nil
}

// GetAllGroups возвращает список всех групп
func (r *Repository) GetAllGroups(ctx context.Context) ([]string, error) {
	query := `SELECT group_name FROM groups ORDER BY created_at DESC`
	
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения групп: %w", err)
	}
	defer rows.Close()
	
	var groups []string
	for rows.Next() {
		var groupName string
		if err := rows.Scan(&groupName); err != nil {
			return nil, err
		}
		groups = append(groups, groupName)
	}
	
	return groups, nil
}

