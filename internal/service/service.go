// Package service содержит бизнес-логику приложения.
package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/rymax1e/open-cashback-advisor/internal/database"
	"github.com/rymax1e/open-cashback-advisor/internal/models"
	"github.com/rymax1e/open-cashback-advisor/internal/validator"
)

// Константы для fuzzy поиска.
const (
	fuzzyThresholdGroup    = 0.6
	fuzzyThresholdCategory = 0.6
	fuzzyThresholdBank     = 0.65
	fuzzyThresholdUser     = 0.7
	fuzzyLimit             = 5
)

// Лимиты для пагинации.
const (
	defaultLimit = 20
	maxLimit     = 1000
)

// Ошибки сервиса.
var (
	ErrGroupNotExists = errors.New("группа не существует")
)

// Service представляет бизнес-логику приложения.
type Service struct {
	repo database.RepositoryInterface
}

// NewService создаёт новый сервис.
func NewService(repo database.RepositoryInterface) *Service {
	return &Service{repo: repo}
}

// --- Методы для работы с кэшбэком ---

// Suggest анализирует данные и возвращает предложения.
func (s *Service) Suggest(ctx context.Context, req *models.SuggestRequest) (*models.SuggestResponse, error) {
	validationErrors := validator.ValidateSuggestRequest(
		req.GroupName, req.Category, req.BankName, req.UserDisplayName,
		req.MonthYear, req.CashbackPercent, req.MaxAmount,
	)

	response := &models.SuggestResponse{
		Valid:       len(validationErrors) == 0,
		Suggestions: models.Suggestions{},
		CanProceed:  len(validationErrors) == 0,
	}

	if len(validationErrors) > 0 {
		response.Errors = validationErrors.Strings()
		return response, nil
	}

	// Выполняем fuzzy поиск
	if err := s.fillSuggestions(ctx, req, response); err != nil {
		return nil, err
	}

	return response, nil
}

// fillSuggestions заполняет предложения из fuzzy поиска.
func (s *Service) fillSuggestions(ctx context.Context, req *models.SuggestRequest, resp *models.SuggestResponse) error {
	var err error

	resp.Suggestions.GroupName, err = s.repo.FuzzySearchGroupName(ctx, req.GroupName, fuzzyThresholdGroup, fuzzyLimit)
	if err != nil {
		return fmt.Errorf("fuzzy-поиск group_name: %w", err)
	}

	resp.Suggestions.Category, err = s.repo.FuzzySearchCategory(ctx, req.Category, fuzzyThresholdCategory, fuzzyLimit)
	if err != nil {
		return fmt.Errorf("fuzzy-поиск category: %w", err)
	}

	resp.Suggestions.BankName, err = s.repo.FuzzySearchBankName(ctx, req.BankName, fuzzyThresholdBank, fuzzyLimit)
	if err != nil {
		return fmt.Errorf("fuzzy-поиск bank_name: %w", err)
	}

	resp.Suggestions.UserDisplayName, err = s.repo.FuzzySearchUserDisplayName(ctx, req.UserDisplayName, fuzzyThresholdUser, fuzzyLimit)
	if err != nil {
		return fmt.Errorf("fuzzy-поиск user_display_name: %w", err)
	}

	return nil
}

// CreateCashback создаёт новое правило кэшбэка.
func (s *Service) CreateCashback(ctx context.Context, req *models.CreateCashbackRequest) (*models.CashbackRule, error) {
	validationErrors := validator.ValidateCreateRequest(
		req.GroupName, req.Category, req.BankName, req.UserID,
		req.UserDisplayName, req.MonthYear, req.CashbackPercent, req.MaxAmount,
	)

	if len(validationErrors) > 0 {
		return nil, fmt.Errorf("ошибки валидации: %s", validationErrors.Error())
	}

	monthYear, _ := validator.ValidateMonthYear(req.MonthYear)

	rule := &models.CashbackRule{
		GroupName:       req.GroupName,
		Category:        req.Category,
		BankName:        req.BankName,
		UserID:          req.UserID,
		UserDisplayName: req.UserDisplayName,
		MonthYear:       monthYear,
		CashbackPercent: validator.RoundToTwoDecimals(req.CashbackPercent),
		MaxAmount:       validator.RoundToTwoDecimals(req.MaxAmount),
	}

	if err := s.repo.Create(ctx, rule); err != nil {
		return nil, fmt.Errorf("создание правила: %w", err)
	}

	return rule, nil
}

// GetCashback получает правило по ID.
func (s *Service) GetCashback(ctx context.Context, id int64) (*models.CashbackRule, error) {
	return s.repo.GetByID(ctx, id)
}

// UpdateCashback обновляет правило кэшбэка.
func (s *Service) UpdateCashback(ctx context.Context, id int64, req *models.UpdateCashbackRequest) error {
	updates, err := s.buildUpdates(req)
	if err != nil {
		return err
	}

	return s.repo.Update(ctx, id, updates)
}

// buildUpdates строит карту обновлений из запроса.
func (s *Service) buildUpdates(req *models.UpdateCashbackRequest) (map[string]interface{}, error) {
	updates := make(map[string]interface{})

	if req.GroupName != "" {
		if err := validator.ValidateTextField("group_name", req.GroupName, true); err != nil {
			return nil, err
		}
		updates["group_name"] = req.GroupName
	}

	if req.Category != "" {
		if err := validator.ValidateTextField("category", req.Category, true); err != nil {
			return nil, err
		}
		updates["category"] = req.Category
	}

	if req.BankName != "" {
		if err := validator.ValidateTextField("bank_name", req.BankName, true); err != nil {
			return nil, err
		}
		updates["bank_name"] = req.BankName
	}

	if req.MonthYear != "" {
		monthYear, err := validator.ValidateMonthYear(req.MonthYear)
		if err != nil {
			return nil, err
		}
		updates["month_year"] = monthYear
	}

	if req.CashbackPercent > 0 {
		if err := validator.ValidateCashbackPercent(req.CashbackPercent); err != nil {
			return nil, err
		}
		updates["cashback_percent"] = validator.RoundToTwoDecimals(req.CashbackPercent)
	}

	if req.MaxAmount > 0 {
		if err := validator.ValidateMaxAmount(req.MaxAmount); err != nil {
			return nil, err
		}
		updates["max_amount"] = validator.RoundToTwoDecimals(req.MaxAmount)
	}

	return updates, nil
}

// DeleteCashback удаляет правило кэшбэка.
func (s *Service) DeleteCashback(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

// ListCashback получает список правил с пагинацией.
func (s *Service) ListCashback(ctx context.Context, req *models.ListCashbackRequest) (*models.ListCashbackResponse, error) {
	limit, offset := s.normalizePagination(req.Limit, req.Offset)

	rules, total, err := s.repo.List(ctx, limit, offset, req.GroupName)
	if err != nil {
		return nil, err
	}

	return &models.ListCashbackResponse{
		Rules:  rules,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

// normalizePagination нормализует параметры пагинации.
func (s *Service) normalizePagination(limit, offset int) (int, int) {
	if limit <= 0 {
		limit = defaultLimit
	} else if limit > maxLimit {
		limit = maxLimit
	}

	if offset < 0 {
		offset = 0
	}

	return limit, offset
}

// GetBestCashback получает правило с лучшим кэшбэком с fallback на "Все покупки".
func (s *Service) GetBestCashback(ctx context.Context, req *models.BestCashbackRequest) (*models.CashbackRule, error) {
	if err := validator.ValidateTextField("group_name", req.GroupName, true); err != nil {
		return nil, err
	}
	if err := validator.ValidateTextField("category", req.Category, true); err != nil {
		return nil, err
	}

	monthYear, err := validator.ValidateMonthYear(req.MonthYear)
	if err != nil {
		return nil, err
	}

	// Сначала ищем точное совпадение категории
	categoryRule, err := s.repo.GetBestCashback(ctx, req.GroupName, req.Category, monthYear)
	
	// Ищем кэшбэк на "Все покупки"
	allPurchasesRule, errAll := s.repo.GetBestCashback(ctx, req.GroupName, "Все покупки", monthYear)
	
	// Если нашли точную категорию
	if err == nil {
		// Если нашли "Все покупки" и он выгоднее
		if errAll == nil && allPurchasesRule.CashbackPercent > categoryRule.CashbackPercent {
			return allPurchasesRule, nil
		}
		return categoryRule, nil
	}
	
	// Если не нашли точную категорию, возвращаем "Все покупки" (если есть)
	if errAll == nil {
		return allPurchasesRule, nil
	}
	
	// Если ничего не нашли, возвращаем ошибку от первого запроса
	return nil, err
}

// --- Методы для работы с группами ---

// CreateGroup создаёт новую группу.
func (s *Service) CreateGroup(ctx context.Context, groupName, creatorID string) error {
	return s.repo.CreateGroup(ctx, groupName, creatorID)
}

// GetUserGroup получает группу пользователя.
func (s *Service) GetUserGroup(ctx context.Context, userID string) (string, error) {
	return s.repo.GetUserGroup(ctx, userID)
}

// SetUserGroup устанавливает группу пользователя.
func (s *Service) SetUserGroup(ctx context.Context, userID, groupName string) error {
	exists, err := s.repo.GroupExists(ctx, groupName)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("группа \"%s\": %w", groupName, ErrGroupNotExists)
	}

	return s.repo.SetUserGroup(ctx, userID, groupName)
}

// GroupExists проверяет существование группы.
func (s *Service) GroupExists(ctx context.Context, groupName string) (bool, error) {
	return s.repo.GroupExists(ctx, groupName)
}

// GetAllGroups возвращает список всех групп.
func (s *Service) GetAllGroups(ctx context.Context) ([]string, error) {
	return s.repo.GetAllGroups(ctx)
}

// GetGroupMembers возвращает участников группы.
func (s *Service) GetGroupMembers(ctx context.Context, groupName string) ([]string, error) {
	return s.repo.GetGroupMembers(ctx, groupName)
}

// GetCashbackByBank получает все кэшбэки по банку в группе.
func (s *Service) GetCashbackByBank(ctx context.Context, groupName, bankName string) ([]models.CashbackRule, error) {
	if err := validator.ValidateTextField("group_name", groupName, true); err != nil {
		return nil, err
	}
	if err := validator.ValidateTextField("bank_name", bankName, true); err != nil {
		return nil, err
	}

	// Используем текущую дату для фильтрации активных кэшбэков
	now := time.Now()
	return s.repo.GetCashbackByBank(ctx, groupName, bankName, now)
}

// GetActiveCategories возвращает список активных категорий в группе.
func (s *Service) GetActiveCategories(ctx context.Context, groupName string) ([]string, error) {
	if err := validator.ValidateTextField("group_name", groupName, true); err != nil {
		return nil, err
	}

	now := time.Now()
	return s.repo.GetActiveCategories(ctx, groupName, now)
}

// GetActiveBanks возвращает список активных банков в группе.
func (s *Service) GetActiveBanks(ctx context.Context, groupName string) ([]string, error) {
	if err := validator.ValidateTextField("group_name", groupName, true); err != nil {
		return nil, err
	}

	now := time.Now()
	return s.repo.GetActiveBanks(ctx, groupName, now)
}

// GetGroupUsers возвращает список пользователей группы.
func (s *Service) GetGroupUsers(ctx context.Context, groupName string) ([]models.UserInfo, error) {
	if err := validator.ValidateTextField("group_name", groupName, true); err != nil {
		return nil, err
	}

	return s.repo.GetGroupUsers(ctx, groupName)
}
