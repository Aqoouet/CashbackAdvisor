package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/rymax1e/open-cashback-advisor/internal/models"
)

// APIClient — клиент для взаимодействия с API сервиса.
type APIClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewAPIClient создаёт новый API клиент.
func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: HTTPClientTimeout,
		},
	}
}

// --- Приватные методы для HTTP запросов ---

// doRequest выполняет HTTP запрос и возвращает тело ответа.
func (c *APIClient) doRequest(req *http.Request) ([]byte, int, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка запроса: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	return body, resp.StatusCode, nil
}

// get выполняет GET запрос.
func (c *APIClient) get(endpoint string, params url.Values) ([]byte, int, error) {
	requestURL := c.baseURL + endpoint
	if len(params) > 0 {
		requestURL += "?" + params.Encode()
	}

	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	return c.doRequest(req)
}

// post выполняет POST запрос с JSON телом.
func (c *APIClient) post(endpoint string, payload interface{}) ([]byte, int, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка сериализации: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка создания запроса: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	return c.doRequest(req)
}

// put выполняет PUT запрос с JSON телом.
func (c *APIClient) put(endpoint string, payload interface{}) ([]byte, int, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка сериализации: %w", err)
	}

	req, err := http.NewRequest(http.MethodPut, c.baseURL+endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка создания запроса: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	return c.doRequest(req)
}

// delete выполняет DELETE запрос.
func (c *APIClient) delete(endpoint string) (int, error) {
	req, err := http.NewRequest(http.MethodDelete, c.baseURL+endpoint, nil)
	if err != nil {
		return 0, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	_, statusCode, err := c.doRequest(req)
	return statusCode, err
}

// parseResponse парсит JSON ответ в структуру.
func parseResponse[T any](body []byte, statusCode int, expectedStatus int) (*T, error) {
	if statusCode != expectedStatus {
		return nil, parseAPIError(body, statusCode)
	}

	var result T
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("ошибка десериализации: %w", err)
	}

	return &result, nil
}

// parseAPIError извлекает ошибку из ответа API.
func parseAPIError(body []byte, statusCode int) error {
	var errResp models.ErrorResponse
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != "" {
		return fmt.Errorf("%s", errResp.Error)
	}
	return fmt.Errorf("ошибка API: статус %d", statusCode)
}

// --- Методы для работы с кэшбэком ---

// Suggest вызывает эндпоинт /suggest для анализа данных.
func (c *APIClient) Suggest(req *models.SuggestRequest) (*models.SuggestResponse, error) {
	body, statusCode, err := c.post(EndpointCashbackSuggest, req)
	if err != nil {
		return nil, err
	}
	return parseResponse[models.SuggestResponse](body, statusCode, http.StatusOK)
}

// CreateCashback создаёт новое правило кэшбэка.
func (c *APIClient) CreateCashback(req *models.CreateCashbackRequest) (*models.CashbackRule, error) {
	body, statusCode, err := c.post(EndpointCashback, req)
	if err != nil {
		return nil, err
	}
	return parseResponse[models.CashbackRule](body, statusCode, http.StatusCreated)
}

// GetCashbackByID получает правило по ID.
func (c *APIClient) GetCashbackByID(id int64) (*models.CashbackRule, error) {
	endpoint := fmt.Sprintf("%s/%d", EndpointCashback, id)
	body, statusCode, err := c.get(endpoint, nil)
	if err != nil {
		return nil, err
	}
	return parseResponse[models.CashbackRule](body, statusCode, http.StatusOK)
}

// UpdateCashback обновляет правило кэшбэка.
func (c *APIClient) UpdateCashback(id int64, req *models.UpdateCashbackRequest) (*models.CashbackRule, error) {
	endpoint := fmt.Sprintf("%s/%d", EndpointCashback, id)
	body, statusCode, err := c.put(endpoint, req)
	if err != nil {
		return nil, err
	}
	return parseResponse[models.CashbackRule](body, statusCode, http.StatusOK)
}

// DeleteCashback удаляет правило по ID.
func (c *APIClient) DeleteCashback(id int64) error {
	endpoint := fmt.Sprintf("%s/%d", EndpointCashback, id)
	statusCode, err := c.delete(endpoint)
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK {
		return fmt.Errorf("ошибка удаления: статус %d", statusCode)
	}
	return nil
}

// ListCashback получает список правил группы.
func (c *APIClient) ListCashback(groupName string, limit, offset int) (*models.ListCashbackResponse, error) {
	params := url.Values{}
	params.Add("group_name", groupName)
	params.Add("limit", fmt.Sprintf("%d", limit))
	params.Add("offset", fmt.Sprintf("%d", offset))

	body, statusCode, err := c.get(EndpointCashback, params)
	if err != nil {
		return nil, err
	}
	return parseResponse[models.ListCashbackResponse](body, statusCode, http.StatusOK)
}

// GetBestCashback получает лучший кэшбэк.
func (c *APIClient) GetBestCashback(groupName, category, monthYear string) (*models.CashbackRule, error) {
	params := url.Values{}
	params.Add("group_name", groupName)
	params.Add("category", category)
	params.Add("month_year", monthYear)

	body, statusCode, err := c.get(EndpointCashbackBest, params)
	if err != nil {
		return nil, err
	}
	return parseResponse[models.CashbackRule](body, statusCode, http.StatusOK)
}

// ListAllCategories получает список всех уникальных категорий.
func (c *APIClient) ListAllCategories(groupName, monthYear string) ([]string, error) {
	params := url.Values{}
	params.Add("group_name", groupName)
	params.Add("month_year", monthYear)
	params.Add("limit", fmt.Sprintf("%d", MaxListLimit))

	body, statusCode, err := c.get(EndpointCashback, params)
	if err != nil {
		return nil, err
	}

	result, err := parseResponse[models.ListCashbackResponse](body, statusCode, http.StatusOK)
	if err != nil {
		return nil, err
	}

	// Собираем уникальные категории
	categorySet := make(map[string]struct{})
	for _, rule := range result.Rules {
		categorySet[rule.Category] = struct{}{}
	}

	categories := make([]string, 0, len(categorySet))
	for cat := range categorySet {
		categories = append(categories, cat)
	}

	return categories, nil
}

// --- Методы для работы с группами ---

// GetUserGroup получает группу пользователя.
func (c *APIClient) GetUserGroup(userID string) (string, error) {
	endpoint := fmt.Sprintf(EndpointUserGroup, userID)
	body, statusCode, err := c.get(endpoint, nil)
	if err != nil {
		return "", err
	}

	if statusCode == http.StatusNotFound {
		return "", ErrUserNotInGroup
	}
	if statusCode != http.StatusOK {
		return "", parseAPIError(body, statusCode)
	}

	var result struct {
		GroupName string `json:"group_name"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("ошибка декодирования: %w", err)
	}

	return result.GroupName, nil
}

// CreateGroup создаёт новую группу.
func (c *APIClient) CreateGroup(groupName, creatorID string) error {
	payload := map[string]string{
		"group_name": groupName,
		"creator_id": creatorID,
	}

	body, statusCode, err := c.post(EndpointGroups, payload)
	if err != nil {
		return err
	}

	if statusCode != http.StatusCreated && statusCode != http.StatusOK {
		return parseAPIError(body, statusCode)
	}

	return nil
}

// JoinGroup присоединяет пользователя к группе.
func (c *APIClient) JoinGroup(userID, groupName string) error {
	endpoint := fmt.Sprintf(EndpointUserGroup, userID)
	payload := map[string]string{
		"group_name": groupName,
	}

	body, statusCode, err := c.put(endpoint, payload)
	if err != nil {
		return err
	}

	if statusCode != http.StatusOK {
		return parseAPIError(body, statusCode)
	}

	return nil
}

// GroupExists проверяет существование группы.
func (c *APIClient) GroupExists(groupName string) bool {
	params := url.Values{}
	params.Add("name", groupName)

	_, statusCode, err := c.get(EndpointGroupsCheck, params)
	if err != nil {
		return false
	}

	return statusCode == http.StatusOK
}

// GetAllGroups возвращает список всех групп.
func (c *APIClient) GetAllGroups() ([]string, error) {
	body, statusCode, err := c.get(EndpointGroups, nil)
	if err != nil {
		return nil, err
	}

	if statusCode != http.StatusOK {
		return nil, parseAPIError(body, statusCode)
	}

	var result struct {
		Groups []string `json:"groups"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования: %w", err)
	}

	return result.Groups, nil
}

// GetGroupMembers возвращает участников группы.
func (c *APIClient) GetGroupMembers(groupName string) ([]string, error) {
	params := url.Values{}
	params.Add("name", groupName)

	body, statusCode, err := c.get(EndpointGroupsMembers, params)
	if err != nil {
		return nil, err
	}

	if statusCode != http.StatusOK {
		return nil, parseAPIError(body, statusCode)
	}

	var result struct {
		Members []string `json:"members"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования: %w", err)
	}

	return result.Members, nil
}
