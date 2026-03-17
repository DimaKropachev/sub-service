package service

import (
	"context"
	"errors"

	"github.com/DimaKropachev/sub-service/internal/models"
	"github.com/DimaKropachev/sub-service/internal/repository"
	"github.com/DimaKropachev/sub-service/internal/transport/http"
	"go.uber.org/zap"
)

type Repository interface {
	AddSub(context.Context, models.Subscription) (int64, error)
	GetSub(context.Context, int64) (models.Subscription, error)
	UpdateSub(context.Context, models.UpdateSubscription) error
	DeleteSub(context.Context, int64) error
	GetListSub(context.Context, models.Pagination) ([]models.Subscription, error)
	GetTotalCost(context.Context, models.SubscriptionFilter) (int64, error)
}

type Service struct {
	repo Repository
	log  *zap.Logger
}

func New(repo Repository, log *zap.Logger) *Service {
	return &Service{
		repo: repo,
		log:  log,
	}
}

func (s *Service) AddSub(ctx context.Context, sub models.Subscription) (int64, error) {
	requestID := ctx.Value(http.RequestIDKey).(string)

	l := s.log.With(
		zap.String("method", "AddSub"),
		zap.String("x-request-id", requestID),
		zap.String("user_id", sub.UserID.String()),
	)

	id, err := s.repo.AddSub(ctx, sub)
	if err != nil {
		l.Error("add subscription failed", zap.Error(err))
		return 0, err
	}
	l.Debug("subscription added", zap.Int64("subscription_id", id))
	return id, nil
}

func (s *Service) GetSub(ctx context.Context, id int64) (models.Subscription, error) {
	requestID := ctx.Value(http.RequestIDKey).(string)

	l := s.log.With(
		zap.String("method", "GetSub"),
		zap.String("x-request-id", requestID),
		zap.Int64("subscription_id", id),
	)

	sub, err := s.repo.GetSub(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrSubscriptionNoFound) {
			l.Warn("subsription not found")
		} else {
			l.Error("failed get subscription", zap.Error(err))
		}
		return models.Subscription{}, err
	}
	l.Debug("subscription recieved")
	return sub, nil
}

func (s *Service) UpdateSub(ctx context.Context, sub models.UpdateSubscription) error {
	requestID := ctx.Value(http.RequestIDKey).(string)

	l := s.log.With(
		zap.String("method", "UpdateSub"),
		zap.String("x-request-id", requestID),
		zap.Int64("subscription_id", sub.ID),
	)

	if err := s.repo.UpdateSub(ctx, sub); err != nil {
		if errors.Is(err, repository.ErrSubscriptionNoFound) {
			l.Warn("subsription not found for update")
		} else {
			l.Error("failed update subscription", zap.Error(err))
		}
		return err
	}
	l.Debug("subscription updated")
	return nil
}

func (s *Service) DeleteSub(ctx context.Context, id int64) error {
	requestID := ctx.Value(http.RequestIDKey).(string)

	l := s.log.With(
		zap.String("method", "DeleteSub"),
		zap.String("x-request-id", requestID),
		zap.Int64("subscription_id", id),
	)

	if err := s.repo.DeleteSub(ctx, id); err != nil {
		if errors.Is(err, repository.ErrSubscriptionNoFound) {
			l.Warn("subsription not found for delete")
		} else {
			l.Error("failed delete subscription", zap.Error(err))
		}
		return err
	}
	l.Debug("subscription deleted")
	return nil
}

func (s *Service) GetListSub(ctx context.Context, pagination models.Pagination) ([]models.Subscription, error) {
	requestID := ctx.Value(http.RequestIDKey).(string)

	l := s.log.With(
		zap.String("method", "GetListSub"),
		zap.String("x-request-id", requestID),
	)

	subs, err := s.repo.GetListSub(ctx, pagination)
	if err != nil {
		l.Error("failed get subscriptions", zap.Error(err))
		return nil, err
	}
	l.Debug("subscriptions recieved")
	return subs, nil
}

func (s *Service) GetTotalCost(ctx context.Context, filter models.SubscriptionFilter) (int64, error) {
	requestID := ctx.Value(http.RequestIDKey).(string)

	l := s.log.With(
		zap.String("method", "GetTotalCost"),
		zap.String("x-request-id", requestID),
	)

	totalCost, err := s.repo.GetTotalCost(ctx, filter)
	if err != nil {
		l.Error("failed get total cost", zap.Error(err))
		return 0, err
	}
	l.Debug("total cost recieved")
	return totalCost, nil
}
