package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	// Правильное кодирование параметров URL
	params := url.Values{}
	params.Add("group_name", groupName)
	params.Add("category", category)
	params.Add("month_year", monthYear)
	
	requestURL := fmt.Sprintf("%s/api/v1/cashback/best?%s", c.baseURL, params.Encode())
	
	resp, err := c.httpClient.Get(requestURL)
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

// ListAllCategories получает список всех уникальных категорий
func (c *APIClient) ListAllCategories(groupName, monthYear string) ([]string, error) {
	params := url.Values{}
	params.Add("group_name", groupName)
	params.Add("month_year", monthYear)
	params.Add("limit", "1000") // Большой лимит для получения всех
	
	requestURL := fmt.Sprintf("%s/api/v1/cashback?%s", c.baseURL, params.Encode())
	
	resp, err := c.httpClient.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ошибка API: статус %d", resp.StatusCode)
	}

	var result models.ListCashbackResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("ошибка десериализации: %w", err)
	}

	// Собираем уникальные категории
	categories := make(map[string]bool)
	for _, rule := range result.Rules {
		categories[rule.Category] = true
	}
	
	var uniqueCategories []string
	for cat := range categories {
		uniqueCategories = append(uniqueCategories, cat)
	}
	
	return uniqueCategories, nil
}

// ListCashback получает список правил пользователя
func (c *APIClient) ListCashback(userID string, limit, offset int) (*models.ListCashbackResponse, error) {
	// Правильное кодирование параметров URL
	params := url.Values{}
	params.Add("user_id", userID)
	params.Add("limit", fmt.Sprintf("%d", limit))
	params.Add("offset", fmt.Sprintf("%d", offset))
	
	requestURL := fmt.Sprintf("%s/api/v1/cashback?%s", c.baseURL, params.Encode())
	
	resp, err := c.httpClient.Get(requestURL)
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

// GetCashbackByID получает правило по ID
func (c *APIClient) GetCashbackByID(id int64) (*models.CashbackRule, error) {
	requestURL := fmt.Sprintf("%s/api/v1/cashback/%d", c.baseURL, id)
	
	resp, err := c.httpClient.Get(requestURL)
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

// UpdateCashback обновляет правило кэшбэка
func (c *APIClient) UpdateCashback(id int64, req *models.UpdateCashbackRequest) (*models.CashbackRule, error) {
	requestURL := fmt.Sprintf("%s/api/v1/cashback/%d", c.baseURL, id)
	
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка сериализации: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPut, requestURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
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

// DeleteCashback удаляет правило по ID
func (c *APIClient) DeleteCashback(id int64) error {
	requestURL := fmt.Sprintf("%s/api/v1/cashback/%d", c.baseURL, id)
	
	req, err := http.NewRequest(http.MethodDelete, requestURL, nil)
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

// --- Методы для работы с группами ---

// GetUserGroup получает группу пользователя
func (c *APIClient) GetUserGroup(userID string) (string, error) {
	url := fmt.Sprintf("%s/api/v1/users/%s/group", c.baseURL, userID)
	
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("ошибка запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("пользователь не в группе")
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ошибка API: статус %d", resp.StatusCode)
	}

	var result struct {
		GroupName string `json:"group_name"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("ошибка декодирования: %w", err)
	}

	return result.GroupName, nil
}

// CreateGroup создаёт новую группу
func (c *APIClient) CreateGroup(groupName, creatorID string) error {
	url := fmt.Sprintf("%s/api/v1/groups", c.baseURL)
	
	reqData := map[string]string{
		"group_name": groupName,
		"creator_id": creatorID,
	}
	
	body, err := json.Marshal(reqData)
	if err != nil {
		return fmt.Errorf("ошибка сериализации: %w", err)
	}

	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("ошибка запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		var errResp models.ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil {
			return fmt.Errorf("%s", errResp.Error)
		}
		return fmt.Errorf("ошибка API: статус %d", resp.StatusCode)
	}

	return nil
}

// JoinGroup присоединяет пользователя к группе
func (c *APIClient) JoinGroup(userID, groupName string) error {
	url := fmt.Sprintf("%s/api/v1/users/%s/group", c.baseURL, userID)
	
	reqData := map[string]string{
		"group_name": groupName,
	}
	
	body, err := json.Marshal(reqData)
	if err != nil {
		return fmt.Errorf("ошибка сериализации: %w", err)
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("ошибка создания запроса: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

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

// GroupExists проверяет существование группы
func (c *APIClient) GroupExists(groupName string) bool {
	url := fmt.Sprintf("%s/api/v1/groups/%s", c.baseURL, groupName)
	
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// GetAllGroups возвращает список всех групп
func (c *APIClient) GetAllGroups() ([]string, error) {
	url := fmt.Sprintf("%s/api/v1/groups", c.baseURL)
	
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ошибка API: статус %d", resp.StatusCode)
	}

	var result struct {
		Groups []string `json:"groups"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования: %w", err)
	}

	return result.Groups, nil
}

// GetGroupMembers возвращает участников группы
func (c *APIClient) GetGroupMembers(groupName string) ([]string, error) {
	url := fmt.Sprintf("%s/api/v1/groups/%s/members", c.baseURL, groupName)
	
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ошибка API: статус %d", resp.StatusCode)
	}

	var result struct {
		Members []string `json:"members"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования: %w", err)
	}

	return result.Members, nil
}

