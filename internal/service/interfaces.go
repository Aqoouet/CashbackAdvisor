package service

import (
	"context"

	"github.com/rymax1e/open-cashback-advisor/internal/models"
)

// ServiceInterface определяет контракт для сервиса.
type ServiceInterface interface {
	// Кэшбэк
	Suggest(ctx context.Context, req *models.SuggestRequest) (*models.SuggestResponse, error)
	CreateCashback(ctx context.Context, req *models.CreateCashbackRequest) (*models.CashbackRule, error)
	GetCashback(ctx context.Context, id int64) (*models.CashbackRule, error)
	UpdateCashback(ctx context.Context, id int64, req *models.UpdateCashbackRequest) error
	DeleteCashback(ctx context.Context, id int64) error
	ListCashback(ctx context.Context, req *models.ListCashbackRequest) (*models.ListCashbackResponse, error)
	GetBestCashback(ctx context.Context, req *models.BestCashbackRequest) (*models.CashbackRule, error)

	// Группы
	CreateGroup(ctx context.Context, groupName, creatorID string) error
	GetUserGroup(ctx context.Context, userID string) (string, error)
	SetUserGroup(ctx context.Context, userID, groupName string) error
	GroupExists(ctx context.Context, groupName string) (bool, error)
	GetAllGroups(ctx context.Context) ([]string, error)
	GetGroupMembers(ctx context.Context, groupName string) ([]string, error)
}

// Проверка реализации интерфейса.
var _ ServiceInterface = (*Service)(nil)

