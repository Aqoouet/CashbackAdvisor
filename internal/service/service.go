package service

import (
	"context"
	"fmt"

	"github.com/rymax1e/open-cashback-advisor/internal/database"
	"github.com/rymax1e/open-cashback-advisor/internal/models"
	"github.com/rymax1e/open-cashback-advisor/internal/validator"
)

// Service представляет бизнес-логику приложения
type Service struct {
	repo *database.Repository
}

// NewService создает новый сервис
func NewService(repo *database.Repository) *Service {
	return &Service{repo: repo}
}

// Suggest анализирует данные и возвращает предложения
func (s *Service) Suggest(ctx context.Context, req *models.SuggestRequest) (*models.SuggestResponse, error) {
	// Валидация всех полей
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

	// Fuzzy-поиск по всем текстовым полям
	groupSuggestions, err := s.repo.FuzzySearchGroupName(ctx, req.GroupName, 0.6, 5)
	if err != nil {
		return nil, fmt.Errorf("ошибка fuzzy-поиска group_name: %w", err)
	}
	response.Suggestions.GroupName = groupSuggestions

	categorySuggestions, err := s.repo.FuzzySearchCategory(ctx, req.Category, 0.6, 5)
	if err != nil {
		return nil, fmt.Errorf("ошибка fuzzy-поиска category: %w", err)
	}
	response.Suggestions.Category = categorySuggestions

	bankSuggestions, err := s.repo.FuzzySearchBankName(ctx, req.BankName, 0.65, 5)
	if err != nil {
		return nil, fmt.Errorf("ошибка fuzzy-поиска bank_name: %w", err)
	}
	response.Suggestions.BankName = bankSuggestions

	userSuggestions, err := s.repo.FuzzySearchUserDisplayName(ctx, req.UserDisplayName, 0.7, 5)
	if err != nil {
		return nil, fmt.Errorf("ошибка fuzzy-поиска user_display_name: %w", err)
	}
	response.Suggestions.UserDisplayName = userSuggestions

	return response, nil
}

// CreateCashback создает новое правило кэшбэка
func (s *Service) CreateCashback(ctx context.Context, req *models.CreateCashbackRequest) (*models.CashbackRule, error) {
	// Валидация
	validationErrors := validator.ValidateCreateRequest(
		req.GroupName, req.Category, req.BankName, req.UserID,
		req.UserDisplayName, req.MonthYear, req.CashbackPercent, req.MaxAmount,
	)

	if len(validationErrors) > 0 {
		return nil, fmt.Errorf("ошибки валидации: %s", validationErrors.Error())
	}

	// Парсинг и округление
	monthYear, _ := validator.ValidateMonthYear(req.MonthYear)
	cashbackPercent := validator.RoundToTwoDecimals(req.CashbackPercent)
	maxAmount := validator.RoundToTwoDecimals(req.MaxAmount)

	// Создание правила
	rule := &models.CashbackRule{
		GroupName:       req.GroupName,
		Category:        req.Category,
		BankName:        req.BankName,
		UserID:          req.UserID,
		UserDisplayName: req.UserDisplayName,
		MonthYear:       monthYear,
		CashbackPercent: cashbackPercent,
		MaxAmount:       maxAmount,
	}

	if err := s.repo.Create(ctx, rule); err != nil {
		return nil, fmt.Errorf("не удалось создать правило: %w", err)
	}

	return rule, nil
}

// GetCashback получает правило по ID
func (s *Service) GetCashback(ctx context.Context, id int64) (*models.CashbackRule, error) {
	return s.repo.GetByID(ctx, id)
}

// UpdateCashback обновляет правило кэшбэка
func (s *Service) UpdateCashback(ctx context.Context, id int64, req *models.UpdateCashbackRequest) error {
	updates := make(map[string]interface{})

	if req.GroupName != "" {
		if err := validator.ValidateTextField("group_name", req.GroupName, true); err != nil {
			return err
		}
		updates["group_name"] = req.GroupName
	}

	if req.Category != "" {
		if err := validator.ValidateTextField("category", req.Category, true); err != nil {
			return err
		}
		updates["category"] = req.Category
	}

	if req.BankName != "" {
		if err := validator.ValidateTextField("bank_name", req.BankName, true); err != nil {
			return err
		}
		updates["bank_name"] = req.BankName
	}

	if req.MonthYear != "" {
		monthYear, err := validator.ValidateMonthYear(req.MonthYear)
		if err != nil {
			return err
		}
		updates["month_year"] = monthYear
	}

	if req.CashbackPercent != nil {
		if err := validator.ValidateCashbackPercent(*req.CashbackPercent); err != nil {
			return err
		}
		updates["cashback_percent"] = validator.RoundToTwoDecimals(*req.CashbackPercent)
	}

	if req.MaxAmount != nil {
		if err := validator.ValidateMaxAmount(*req.MaxAmount); err != nil {
			return err
		}
		updates["max_amount"] = validator.RoundToTwoDecimals(*req.MaxAmount)
	}

	return s.repo.Update(ctx, id, updates)
}

// DeleteCashback удаляет правило кэшбэка
func (s *Service) DeleteCashback(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

// ListCashback получает список правил с пагинацией
func (s *Service) ListCashback(ctx context.Context, req *models.ListCashbackRequest) (*models.ListCashbackResponse, error) {
	// Установка значений по умолчанию
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	rules, total, err := s.repo.List(ctx, req.Limit, req.Offset, req.UserID)
	if err != nil {
		return nil, err
	}

	return &models.ListCashbackResponse{
		Rules:  rules,
		Total:  total,
		Limit:  req.Limit,
		Offset: req.Offset,
	}, nil
}

// GetBestCashback получает правило с лучшим кэшбэком
func (s *Service) GetBestCashback(ctx context.Context, req *models.BestCashbackRequest) (*models.CashbackRule, error) {
	// Валидация
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

	return s.repo.GetBestCashback(ctx, req.GroupName, req.Category, monthYear)
}

