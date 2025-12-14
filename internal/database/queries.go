// Package database содержит SQL запросы и работу с базой данных.
package database

// SQL запросы для работы с кэшбэком.
const (
	// QueryCreateCashback — создание нового правила.
	QueryCreateCashback = `
		INSERT INTO cashback_rules (
			group_name, category, bank_name, user_id, user_display_name,
			month_year, cashback_percent, max_amount
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`

	// QueryGetCashbackByID — получение правила по ID.
	QueryGetCashbackByID = `
		SELECT id, group_name, category, bank_name, user_id, user_display_name,
			   month_year, cashback_percent, max_amount, created_at, updated_at
		FROM cashback_rules
		WHERE id = $1`

	// QueryDeleteCashback — удаление правила.
	QueryDeleteCashback = `DELETE FROM cashback_rules WHERE id = $1`

	// QueryCountCashbackByGroup — подсчёт правил группы.
	QueryCountCashbackByGroup = `
		SELECT COUNT(*) 
		FROM cashback_rules cr
		INNER JOIN user_groups ug ON cr.user_id = ug.user_id
		WHERE ug.group_name = $1`

	// QueryListCashbackByGroup — список правил группы.
	QueryListCashbackByGroup = `
		SELECT cr.id, cr.group_name, cr.category, cr.bank_name, cr.user_id, cr.user_display_name,
			   cr.month_year, cr.cashback_percent, cr.max_amount, cr.created_at, cr.updated_at
		FROM cashback_rules cr
		INNER JOIN user_groups ug ON cr.user_id = ug.user_id
		WHERE ug.group_name = $1
		ORDER BY cr.created_at DESC 
		LIMIT $2 OFFSET $3`

	// QueryGetBestCashback — получение лучшего кэшбэка.
	QueryGetBestCashback = `
		SELECT cr.id, cr.group_name, cr.category, cr.bank_name, cr.user_id, cr.user_display_name,
			   cr.month_year, cr.cashback_percent, cr.max_amount, cr.created_at, cr.updated_at
		FROM cashback_rules cr
		INNER JOIN user_groups ug ON cr.user_id = ug.user_id
		WHERE ug.group_name = $1 AND cr.category = $2 AND cr.month_year = $3
		ORDER BY cr.cashback_percent DESC, cr.max_amount DESC
		LIMIT 1`

	// QueryFuzzySearch — fuzzy поиск по полю (шаблон).
	QueryFuzzySearchTemplate = `
		SELECT DISTINCT %s, similarity(%s, $1) as sim
		FROM cashback_rules
		WHERE similarity(%s, $1) >= $2
		ORDER BY sim DESC
		LIMIT $3`
)

// SQL запросы для работы с группами.
const (
	// QuerySetUserGroup — установка группы пользователя.
	QuerySetUserGroup = `
		INSERT INTO user_groups (user_id, group_name, updated_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id) 
		DO UPDATE SET group_name = $2, updated_at = CURRENT_TIMESTAMP`

	// QueryGetUserGroup — получение группы пользователя.
	QueryGetUserGroup = `SELECT group_name FROM user_groups WHERE user_id = $1`

	// QueryCreateGroup — создание группы.
	QueryCreateGroup = `
		INSERT INTO groups (group_name, created_by)
		VALUES ($1, $2)
		ON CONFLICT (group_name) DO NOTHING`

	// QueryGroupExists — проверка существования группы.
	QueryGroupExists = `SELECT EXISTS(SELECT 1 FROM groups WHERE group_name = $1)`

	// QueryGetGroupMembers — получение участников группы.
	QueryGetGroupMembers = `SELECT user_id FROM user_groups WHERE group_name = $1`

	// QueryGetAllGroups — получение всех групп.
	QueryGetAllGroups = `SELECT group_name FROM groups ORDER BY created_at DESC`
)

// Поля для fuzzy поиска.
const (
	FieldGroupName       = "group_name"
	FieldCategory        = "category"
	FieldBankName        = "bank_name"
	FieldUserDisplayName = "user_display_name"
)

