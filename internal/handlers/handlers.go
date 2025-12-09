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
	userID := r.URL.Query().Get("user_id")

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
		Limit:  limit,
		Offset: offset,
		UserID: userID,
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
	r.Route("/api/v1/cashback", func(r chi.Router) {
		r.Post("/suggest", h.Suggest)
		r.Post("/", h.CreateCashback)
		r.Get("/", h.ListCashback)
		r.Get("/best", h.GetBestCashback)
		r.Get("/{id}", h.GetCashback)
		r.Put("/{id}", h.UpdateCashback)
		r.Delete("/{id}", h.DeleteCashback)
	})

	r.Get("/health", h.Health)
}

