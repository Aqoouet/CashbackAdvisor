package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rymax1e/open-cashback-advisor/internal/models"
)

// APIClient клиент для взаимодействия с API сервиса
type APIClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewAPIClient создает новый API клиент
func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Suggest вызывает эндпоинт /suggest для анализа данных
func (c *APIClient) Suggest(req *models.SuggestRequest) (*models.SuggestResponse, error) {
	url := fmt.Sprintf("%s/api/v1/cashback/suggest", c.baseURL)
	
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка сериализации: %w", err)
	}

	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp models.ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil {
			return nil, fmt.Errorf("ошибка API: %s", errResp.Error)
		}
		return nil, fmt.Errorf("ошибка API: статус %d", resp.StatusCode)
	}

	var result models.SuggestResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("ошибка десериализации: %w", err)
	}

	return &result, nil
}

// CreateCashback создает новое правило кэшбэка
func (c *APIClient) CreateCashback(req *models.CreateCashbackRequest) (*models.CashbackRule, error) {
	url := fmt.Sprintf("%s/api/v1/cashback", c.baseURL)
	
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка сериализации: %w", err)
	}

	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		var errResp models.ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil {
			return nil, fmt.Errorf("ошибка API: %s", errResp.Error)
		}
		return nil, fmt.Errorf("ошибка API: статус %d", resp.StatusCode)
	}

	var result models.CashbackRule
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("ошибка десериализации: %w", err)
	}

	return &result, nil
}

// GetBestCashback получает лучший кэшбэк
func (c *APIClient) GetBestCashback(groupName, category, monthYear string) (*models.CashbackRule, error) {
	url := fmt.Sprintf("%s/api/v1/cashback/best?group_name=%s&category=%s&month_year=%s",
		c.baseURL, groupName, category, monthYear)
	
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp models.ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil {
			return nil, fmt.Errorf("%s", errResp.Error)
		}
		return nil, fmt.Errorf("ошибка API: статус %d", resp.StatusCode)
	}

	var result models.CashbackRule
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("ошибка десериализации: %w", err)
	}

	return &result, nil
}

// ListCashback получает список правил пользователя
func (c *APIClient) ListCashback(userID string, limit, offset int) (*models.ListCashbackResponse, error) {
	url := fmt.Sprintf("%s/api/v1/cashback?user_id=%s&limit=%d&offset=%d",
		c.baseURL, userID, limit, offset)
	
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp models.ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil {
			return nil, fmt.Errorf("%s", errResp.Error)
		}
		return nil, fmt.Errorf("ошибка API: статус %d", resp.StatusCode)
	}

	var result models.ListCashbackResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("ошибка десериализации: %w", err)
	}

	return &result, nil
}

// DeleteCashback удаляет правило по ID
func (c *APIClient) DeleteCashback(id int64) error {
	url := fmt.Sprintf("%s/api/v1/cashback/%d", c.baseURL, id)
	
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("ошибка создания запроса: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		var errResp models.ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil {
			return fmt.Errorf("%s", errResp.Error)
		}
		return fmt.Errorf("ошибка API: статус %d", resp.StatusCode)
	}

	return nil
}

