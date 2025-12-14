package database

import (
	"context"
	"errors"
	"time"

	"github.com/rymax1e/open-cashback-advisor/internal/models"
)

// Стандартные ошибки репозитория.
var (
	ErrNotFound       = errors.New("запись не найдена")
	ErrNoRowsAffected = errors.New("нет затронутых строк")
	ErrEmptyUpdate    = errors.New("нет полей для обновления")
)

// RepositoryInterface определяет контракт для репозитория.
type RepositoryInterface interface {
	// Кэшбэк
	Create(ctx context.Context, rule *models.CashbackRule) error
	GetByID(ctx context.Context, id int64) (*models.CashbackRule, error)
	Update(ctx context.Context, id int64, updates map[string]interface{}) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, limit, offset int, groupName string) ([]models.CashbackRule, int, error)
	GetBestCashback(ctx context.Context, groupName, category string, monthYear time.Time) (*models.CashbackRule, error)
	GetAllCashbackByCategory(ctx context.Context, groupName, category string, monthYear time.Time) ([]models.CashbackRule, error)

	// Fuzzy поиск
	FuzzySearchGroupName(ctx context.Context, value string, threshold float64, limit int) ([]models.FuzzySuggestion, error)
	FuzzySearchCategory(ctx context.Context, value string, threshold float64, limit int) ([]models.FuzzySuggestion, error)
	FuzzySearchBankName(ctx context.Context, value string, threshold float64, limit int) ([]models.FuzzySuggestion, error)
	FuzzySearchUserDisplayName(ctx context.Context, value string, threshold float64, limit int) ([]models.FuzzySuggestion, error)

	// Группы
	SetUserGroup(ctx context.Context, userID, groupName string) error
	GetUserGroup(ctx context.Context, userID string) (string, error)
	CreateGroup(ctx context.Context, groupName, creatorID string) error
	GroupExists(ctx context.Context, groupName string) (bool, error)
	GetGroupMembers(ctx context.Context, groupName string) ([]string, error)
	GetAllGroups(ctx context.Context) ([]string, error)

	// Дополнительные методы
	GetCashbackByBank(ctx context.Context, groupName, bankName string, monthYear time.Time) ([]models.CashbackRule, error)
	GetActiveCategories(ctx context.Context, groupName string, monthYear time.Time) ([]string, error)
	GetActiveBanks(ctx context.Context, groupName string, monthYear time.Time) ([]string, error)
}

// Проверка реализации интерфейса.
var _ RepositoryInterface = (*Repository)(nil)

