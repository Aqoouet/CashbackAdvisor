package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/rymax1e/open-cashback-advisor/internal/models"
	"github.com/rymax1e/open-cashback-advisor/internal/service"
)

// Handler представляет HTTP обработчики
type Handler struct {
	service *service.Service
}

// NewHandler создает новый обработчик
func NewHandler(service *service.Service) *Handler {
	return &Handler{service: service}
}

// respondJSON отправляет JSON ответ
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// respondError отправляет ответ с ошибкой
func respondError(w http.ResponseWriter, status int, message string, details ...string) {
	respondJSON(w, status, models.ErrorResponse{
		Error:   message,
		Details: details,
	})
}

// Suggest обрабатывает POST /api/v1/cashback/suggest
func (h *Handler) Suggest(w http.ResponseWriter, r *http.Request) {
	var req models.SuggestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Неверный формат запроса", err.Error())
		return
	}

	response, err := h.service.Suggest(r.Context(), &req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Внутренняя ошибка сервера", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, response)
}

// CreateCashback обрабатывает POST /api/v1/cashback
func (h *Handler) CreateCashback(w http.ResponseWriter, r *http.Request) {
	var req models.CreateCashbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Неверный формат запроса", err.Error())
		return
	}

	rule, err := h.service.CreateCashback(r.Context(), &req)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Ошибка создания правила", err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, rule)
}

// GetCashback обрабатывает GET /api/v1/cashback/{id}
func (h *Handler) GetCashback(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Неверный ID")
		return
	}

	rule, err := h.service.GetCashback(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "Правило не найдено", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, rule)
}

// UpdateCashback обрабатывает PUT /api/v1/cashback/{id}
func (h *Handler) UpdateCashback(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Неверный ID")
		return
	}

	var req models.UpdateCashbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Неверный формат запроса", err.Error())
		return
	}

	if err := h.service.UpdateCashback(r.Context(), id, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Ошибка обновления правила", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Правило успешно обновлено"})
}

// DeleteCashback обрабатывает DELETE /api/v1/cashback/{id}
func (h *Handler) DeleteCashback(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Неверный ID")
		return
	}

	if err := h.service.DeleteCashback(r.Context(), id); err != nil {
		respondError(w, http.StatusNotFound, "Правило не найдено", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Правило успешно удалено"})
}

// ListCashback обрабатывает GET /api/v1/cashback
func (h *Handler) ListCashback(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	userID := r.URL.Query().Get("user_id")       // Legacy
	groupName := r.URL.Query().Get("group_name") // New way

	limit := 20
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	req := &models.ListCashbackRequest{
		Limit:     limit,
		Offset:    offset,
		UserID:    userID,    // Legacy
		GroupName: groupName, // New way
	}

	response, err := h.service.ListCashback(r.Context(), req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Ошибка получения списка", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, response)
}

// GetBestCashback обрабатывает GET /api/v1/cashback/best
func (h *Handler) GetBestCashback(w http.ResponseWriter, r *http.Request) {
	groupName := r.URL.Query().Get("group_name")
	category := r.URL.Query().Get("category")
	monthYear := r.URL.Query().Get("month_year")

	if groupName == "" || category == "" || monthYear == "" {
		respondError(w, http.StatusBadRequest, "Параметры group_name, category и month_year обязательны")
		return
	}

	req := &models.BestCashbackRequest{
		GroupName: groupName,
		Category:  category,
		MonthYear: monthYear,
	}

	rule, err := h.service.GetBestCashback(r.Context(), req)
	if err != nil {
		respondError(w, http.StatusNotFound, "Правила не найдены", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, rule)
}

// Health обрабатывает GET /health
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

// RegisterRoutes регистрирует все маршруты
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1", func(r chi.Router) {
		// Cashback
		r.Route("/cashback", func(r chi.Router) {
			r.Post("/suggest", h.Suggest)
			r.Post("/", h.CreateCashback)
			r.Get("/", h.ListCashback)
			r.Get("/best", h.GetBestCashback)
			r.Get("/{id}", h.GetCashback)
			r.Put("/{id}", h.UpdateCashback)
			r.Delete("/{id}", h.DeleteCashback)
		})

		// Группы
		r.Route("/groups", func(r chi.Router) {
			r.Post("/", h.CreateGroup)
			r.Get("/", h.GetAllGroups)
			r.Get("/check", h.GetGroup)      // ?name=groupName
			r.Get("/members", h.GetGroupMembers) // ?name=groupName
		})

		// Пользователи и группы
		r.Route("/users/{userID}", func(r chi.Router) {
			r.Get("/group", h.GetUserGroup)
			r.Put("/group", h.SetUserGroup)
		})
	})

	r.Get("/health", h.Health)
}

// --- Обработчики для групп ---

// CreateGroup создаёт новую группу
func (h *Handler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	var req struct {
		GroupName string `json:"group_name"`
		CreatorID string `json:"creator_id"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Неверный формат запроса", err.Error())
		return
	}

	if req.GroupName == "" || req.CreatorID == "" {
		respondError(w, http.StatusBadRequest, "Укажите group_name и creator_id")
		return
	}

	err := h.service.CreateGroup(r.Context(), req.GroupName, req.CreatorID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Ошибка создания группы", err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, map[string]string{
		"message":    "Группа создана",
		"group_name": req.GroupName,
	})
}

// GetAllGroups возвращает список всех групп
func (h *Handler) GetAllGroups(w http.ResponseWriter, r *http.Request) {
	groups, err := h.service.GetAllGroups(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Ошибка получения групп", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"groups": groups,
	})
}

// GetGroup проверяет существование группы
func (h *Handler) GetGroup(w http.ResponseWriter, r *http.Request) {
	groupName := r.URL.Query().Get("name")
	if groupName == "" {
		respondError(w, http.StatusBadRequest, "Укажите параметр name")
		return
	}
	
	exists, err := h.service.GroupExists(r.Context(), groupName)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Ошибка проверки группы", err.Error())
		return
	}

	if !exists {
		respondError(w, http.StatusNotFound, "Группа не найдена")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"group_name": groupName,
		"exists":     true,
	})
}

// GetGroupMembers возвращает участников группы
func (h *Handler) GetGroupMembers(w http.ResponseWriter, r *http.Request) {
	groupName := r.URL.Query().Get("name")
	if groupName == "" {
		respondError(w, http.StatusBadRequest, "Укажите параметр name")
		return
	}
	
	members, err := h.service.GetGroupMembers(r.Context(), groupName)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Ошибка получения участников", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"members": members,
	})
}

// GetUserGroup получает группу пользователя
func (h *Handler) GetUserGroup(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	
	groupName, err := h.service.GetUserGroup(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Пользователь не в группе")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"group_name": groupName,
	})
}

// SetUserGroup устанавливает группу пользователя
func (h *Handler) SetUserGroup(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	
	var req struct {
		GroupName string `json:"group_name"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Неверный формат запроса", err.Error())
		return
	}

	if req.GroupName == "" {
		respondError(w, http.StatusBadRequest, "Укажите group_name")
		return
	}

	err := h.service.SetUserGroup(r.Context(), userID, req.GroupName)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Ошибка установки группы", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message":    "Группа установлена",
		"group_name": req.GroupName,
	})
}

