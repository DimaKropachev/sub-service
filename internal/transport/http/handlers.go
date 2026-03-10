package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/DimaKropachev/sub-service/internal/models"
	"github.com/DimaKropachev/sub-service/internal/repository"
	"github.com/DimaKropachev/sub-service/internal/transport/http/dto"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Service interface {
	AddSub(context.Context, models.Subscription) (int64, error)
	GetSub(context.Context, int64) (models.Subscription, error)
	UpdateSub(context.Context, models.UpdateSubscription) error
	DeleteSub(context.Context, int64) error
	GetListSub(context.Context) ([]models.Subscription, error)
	GetTotalCost(context.Context, models.SubscriptionFilter) (int64, error)
}

type Handler struct {
	s   Service
	log *zap.Logger
}

func NewHandler(s Service, log *zap.Logger) *Handler {
	return &Handler{
		s:   s,
		log: log,
	}
}

// @Summary Добавление новой подписки
// @Description Создает новую запись о подписке в базе данных на основе переданных данных
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param input body dto.CreateSubscriptionRequest true "Данные для создания подписки"
// @Success 201 {object} dto.AddSubscriptionResponse "ID созданной подписки"
// @Failure 400 {object} dto.ErrorResponse "Неверный формат JSON или некорректные данные"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router /subscriptions [post]
func (h *Handler) AddNewSubscription(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	ctx := r.Context()
	requestID := ctx.Value(RequestIDKey).(string)
	l := h.log.With(
		zap.String("x-request-id", requestID),
	)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	dtoSub := &dto.CreateSubscriptionRequest{}
	if err := decoder.Decode(dtoSub); err != nil {
		writeErr(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	sub, err := dtoSub.ToDomain()
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}

	id, err := h.s.AddSub(ctx, sub)
	if err != nil {
		l.Error("couldn't add subscription", zap.Error(err))
		writeErr(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err = json.NewEncoder(w).Encode(&dto.AddSubscriptionResponse{
		ID: id,
	}); err != nil {
		l.Error("couldn't send json response", zap.Error(err))
	}
}

// @Summary Получение подписки по ID
// @Description Возвращает информацию о подписке по ее идентификатору
// @Tags subscriptions
// @Produce json
// @Param id path int64 true "ID подписки" example(123)
// @Success 200 {object} dto.GetSubscriptionResponse "Информация о подписке"
// @Failure 400 {object} dto.ErrorResponse "Неверный формат ID"
// @Failure 404 {object} dto.ErrorResponse "Подписка не найдена"
// @Failure 404 {object} dto.ErrorResponse "Подписка не найдена"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router /subscriptions/{id} [get]
func (h *Handler) GetSubscriptionByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctx.Value(RequestIDKey).(string)
	l := h.log.With(
		zap.String("x-request-id", requestID),
	)

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	sub, err := h.s.GetSub(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrSubscriptionNoFound) {
			l.Info("subscription not found", zap.Int64("subscription_id", id))
			writeErr(w, http.StatusNotFound, err.Error())
			return
		}
		l.Error("couldn't get subscription by id", zap.Int64("subscription_id", id), zap.Error(err))
		writeErr(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	dtoResp := dto.FromDomain(sub)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err = json.NewEncoder(w).Encode(&dtoResp); err != nil {
		l.Error("couldn't send json response", zap.Error(err))
	}
}

// @Summary Обновление подписки
// @Description Обновляет данные существующей подписки
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path int64 true "ID подписки" example(123)
// @Param input body dto.UpdateSubscriptionRequest true "Данные для обновления подписки"
// @Success 204 "Подписка успешно обновлена"
// @Failure 400 {object} dto.ErrorResponse "Неверный формат JSON или некорректные данные"
// @Failure 404 {object} dto.ErrorResponse "Подписка не найдена"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router /subscriptions/{id} [patch]
func (h *Handler) UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	ctx := r.Context()
	requestID := ctx.Value(RequestIDKey).(string)
	l := h.log.With(
		zap.String("x-request-id", requestID),
	)

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	dtoUpdateSub := &dto.UpdateSubscriptionRequest{}
	if err := decoder.Decode(dtoUpdateSub); err != nil {
		writeErr(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	updateSub, err := dtoUpdateSub.ToDomain(id)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}

	if err = h.s.UpdateSub(ctx, updateSub); err != nil {
		if errors.Is(err, repository.ErrSubscriptionNoFound) {
			l.Info("subscription not found", zap.Int64("subscription_id", id))
			writeErr(w, http.StatusNotFound, err.Error())
			return
		}
		l.Error("couldn't update subscription", zap.Int64("subscription_id", id), zap.Error(err))
		writeErr(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Удаление подписки
// @Description Удаляет подписку по идентификатору
// @Tags subscriptions
// @Produce json
// @Param id path int64 true "ID подписки" example(123)
// @Success 204 "Подписка успешно удалена"
// @Failure 400 {object} dto.ErrorResponse "Неверный формат ID"
// @Failure 404 {object} dto.ErrorResponse "Подписка не найдена"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router /subscriptions/{id} [delete]
func (h *Handler) DeleteSubscriptionByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctx.Value(RequestIDKey).(string)
	l := h.log.With(
		zap.String("x-request-id", requestID),
	)

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	err = h.s.DeleteSub(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrSubscriptionNoFound) {
			l.Info("subscription not found", zap.Int64("subscription_id", id))
			writeErr(w, http.StatusNotFound, err.Error())
			return	
		}
		l.Error("couldn't delete subscription", zap.Int64("subscription_id", id), zap.Error(err))
		writeErr(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Получение списка подписок
// @Description Получает список всех подписок
// @Tags subscriptions
// @Produce json
// @Success 200 {array} dto.GetSubscriptionResponse "Список подписок"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router /subscriptions [get]
func (h *Handler) GetListSubscriptions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctx.Value(RequestIDKey).(string)
	l := h.log.With(
		zap.String("x-request-id", requestID),
	)

	subs, err := h.s.GetListSub(ctx)
	if err != nil {
		l.Error("couldn't get subscriptions", zap.Error(err))
		writeErr(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	dtoSubs := make([]dto.GetSubscriptionResponse, len(subs))
	for i, sub := range subs {
		dtoSubs[i] = dto.FromDomain(sub)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err = json.NewEncoder(w).Encode(dtoSubs); err != nil {
		l.Error("couldn't send json response", zap.Error(err))
	}
}

// @Summary Получение общей стоимости подписок
// @Description Возращает суммарную стоимость всех подписок
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "ID пользователя (UUID)" example(123e4567-e89b-12d3-a456-426614174000)
// @Param service query string false "Название сервиса" example(Netflix)
// @Param from_date query string true "Дата начала (MM-YYYY)" example(01-2026)
// @Param to_date query string true "Дата окончания (MM-YYYY)" example(07-2026)
// @Success 200 {object} dto.TotalCostResponse "Суммарная стоимость подписок"
// @Failure 400 {object} dto.ErrorResponse "Неверный формат JSON или некорректные данные"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router /subscriptions/cost [get]
func (h *Handler) GetTotalCostSubscriptions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctx.Value(RequestIDKey).(string)
	l := h.log.With(
		zap.String("x-request-id", requestID),
	)

	query := r.URL.Query()

	userIDStr := query.Get("user_id")
	s := query.Get("service")
	from := query.Get("from_date")
	to := query.Get("to_date")

	var (
		userID *uuid.UUID
		err    error
	)
	if userIDStr != "" {
		parsed, err := uuid.Parse(userIDStr)
		if err != nil {
			writeErr(w, http.StatusBadRequest, fmt.Sprintf("invalid UUID: %v", err))
		}
		userID = &parsed
	}

	var service *string
	if s != "" {
		service = &s
	}

	fromDate, err := time.Parse("01-2006", from)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Invalid from_date format, expected MM-YYYY")
		return
	}

	toDate, err := time.Parse("01-2006", to)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Invalid to_date formate, expected MM-YYYY")
		return
	}

	subFilter := models.SubscriptionFilter{
		UserID:  userID,
		Service: service,
		From:    fromDate,
		To:      toDate,
	}

	cost, err := h.s.GetTotalCost(ctx, subFilter)
	if err != nil {
		l.Error("couldn't get total cost subscription",
			zap.Time("from_date", fromDate),
			zap.Time("to_date", toDate),
			zap.Error(err),
		)
		writeErr(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err = json.NewEncoder(w).Encode(&dto.TotalCostResponse{
		TotalCost: cost,
	}); err != nil {
		l.Error("couldn't send json response", zap.Error(err))
	}
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(&dto.ErrorResponse{
		Err: msg,
	})
}
