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
func (r *Repository) List(ctx context.Context, limit, offset int, userID string) ([]models.CashbackRule, int, error) {
	// Запрос на получение общего количества
	countQuery := "SELECT COUNT(*) FROM cashback_rules"
	countArgs := []interface{}{}
	if userID != "" {
		countQuery += " WHERE user_id = $1"
		countArgs = append(countArgs, userID)
	}

	var total int
	err := r.db.Pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("не удалось получить количество записей: %w", err)
	}

	// Запрос на получение правил
	query := `
		SELECT id, group_name, category, bank_name, user_id, user_display_name,
			   month_year, cashback_percent, max_amount, created_at, updated_at
		FROM cashback_rules
	`
	args := []interface{}{}
	if userID != "" {
		query += " WHERE user_id = $1"
		args = append(args, userID)
		query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $2 OFFSET $3")
		args = append(args, limit, offset)
	} else {
		query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $1 OFFSET $2")
		args = append(args, limit, offset)
	}

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
	query := `
		SELECT id, group_name, category, bank_name, user_id, user_display_name,
			   month_year, cashback_percent, max_amount, created_at, updated_at
		FROM cashback_rules
		WHERE group_name = $1 AND category = $2 AND month_year = $3
		ORDER BY cashback_percent DESC, max_amount DESC
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

