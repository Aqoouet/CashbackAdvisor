package bot

import "github.com/rymax1e/open-cashback-advisor/internal/models"

// APIClientInterface определяет контракт для API клиента.
// Позволяет легко создавать mock-объекты для тестирования.
type APIClientInterface interface {
	// Кэшбэк
	Suggest(req *models.SuggestRequest) (*models.SuggestResponse, error)
	CreateCashback(req *models.CreateCashbackRequest) (*models.CashbackRule, error)
	GetCashbackByID(id int64) (*models.CashbackRule, error)
	UpdateCashback(id int64, req *models.UpdateCashbackRequest) (*models.CashbackRule, error)
	DeleteCashback(id int64) error
	ListCashback(groupName string, limit, offset int) (*models.ListCashbackResponse, error)
	GetBestCashback(groupName, category, monthYear string) (*models.CashbackRule, error)
	ListAllCategories(groupName, monthYear string) ([]string, error)

	// Группы
	GetUserGroup(userID string) (string, error)
	CreateGroup(groupName, creatorID string) error
	JoinGroup(userID, groupName string) error
	GroupExists(groupName string) bool
	GetAllGroups() ([]string, error)
	GetGroupMembers(groupName string) ([]string, error)
}

// Проверка, что APIClient реализует интерфейс.
var _ APIClientInterface = (*APIClient)(nil)

