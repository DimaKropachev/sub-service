package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/DimaKropachev/sub-service/internal/models"
	"github.com/Masterminds/squirrel"
)

var (
	ErrSubscriptionNoFound = errors.New("subscription not found")
)

type Repository struct {
	db      *sql.DB
	builder squirrel.StatementBuilderType
}

func New(db *sql.DB) *Repository {
	return &Repository{
		db:      db,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *Repository) AddSub(ctx context.Context, sub models.Subscription) (int64, error) {
	colums := []string{"service", "price", "user_id", "start_date"}
	values := []any{sub.Service, sub.Price, sub.UserID, sub.StartDate}
	if sub.EndDate != nil {
		colums = append(colums, "end_date")
		values = append(values, sub.EndDate)
	}

	qBuild := r.builder.Insert("subscriptions").Columns(colums...).Values(values...).Suffix("RETURNING id")
	q, args, err := qBuild.ToSql()
	if err != nil {
		return 0, fmt.Errorf("failed to build sql-query: %w", err)
	}

	var id int64
	err = r.db.QueryRowContext(ctx, q, args...).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create record: %w", err)
	}

	return id, nil
}

func (r *Repository) GetSub(ctx context.Context, id int64) (models.Subscription, error) {
	q, args, err := r.builder.Select("id", "service", "price", "user_id", "start_date", "end_date").From("subscriptions").Where(squirrel.Eq{"id": id}).ToSql()
	if err != nil {
		return models.Subscription{}, fmt.Errorf("failed to build sql-query: %w", err)
	}

	var endDate sql.NullTime
	sub := models.Subscription{}
	err = r.db.QueryRowContext(ctx, q, args...).Scan(&sub.ID, &sub.Service, &sub.Price, &sub.UserID, &sub.StartDate, &endDate)
	if err != nil {
		if errors.Is(err,sql.ErrNoRows) {
			return models.Subscription{}, ErrSubscriptionNoFound
		}
		return models.Subscription{}, fmt.Errorf("failed to get info of subscription: %w", err)
	}

	if endDate.Valid {
		sub.EndDate = &endDate.Time
	}

	return sub, nil
}

// update only price and end_date
func (r *Repository) UpdateSub(ctx context.Context, sub models.UpdateSubscription) error {
	qBuild := r.builder.Update("subscriptions")

	if sub.Price != nil {
		qBuild = qBuild.Set("price", sub.Price)
	}
	if sub.EndDate != nil {
		qBuild = qBuild.Set("end_date", sub.EndDate)
	}

	q, args, err := qBuild.Where(squirrel.Eq{"id": sub.ID}).ToSql()
	if err != nil {
		return fmt.Errorf("failed to build sql-query: %w", err)
	}

	res, err := r.db.ExecContext(ctx, q, args...)
	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	} else if n == 0 {
		return ErrSubscriptionNoFound
	}

	return nil
}

func (r *Repository) DeleteSub(ctx context.Context, id int64) error {
	q, args, err := r.builder.Delete("subscriptions").Where(squirrel.Eq{"id": id}).ToSql()
	if err != nil {
		return fmt.Errorf("failed to build sql-query: %w", err)
	}

	res, err := r.db.ExecContext(ctx, q, args...)
	if err != nil {
		return fmt.Errorf("failed to delete subscription")
	}

	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	} else if n == 0 {
		return ErrSubscriptionNoFound
	}

	return nil
}

func (r *Repository) GetListSub(ctx context.Context) ([]models.Subscription, error) {
	q, args, err := r.builder.Select("id", "service", "price", "user_id", "start_date", "end_date").From("subscriptions").ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build sql-query: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriptions: %w", err)
	}
	defer rows.Close()

	subs := []models.Subscription{}
	for rows.Next() {
		var endDate sql.NullTime
		sub := models.Subscription{}
		if err = rows.Scan(&sub.ID, &sub.Service, &sub.Price, &sub.UserID, &sub.StartDate, &endDate); err != nil {
			return nil, fmt.Errorf("failed to get subscriptions: %w", err)
		}

		if endDate.Valid {
			sub.EndDate = &endDate.Time
		}

		subs = append(subs, sub)
	}

	return subs, nil
}

func (r *Repository) GetTotalCost(ctx context.Context, filter models.SubscriptionFilter) (int64, error) {
	qBuild := r.builder.Select("COALESCE(SUM(price), 0)").From("subscriptions").Where(squirrel.Expr("start_date BETWEEN ? AND ?", filter.From, filter.To))

	if filter.UserID != nil {
		qBuild = qBuild.Where(squirrel.Eq{"user_id": *filter.UserID})
	}
	if filter.Service != nil {
		qBuild = qBuild.Where(squirrel.Eq{"service": *filter.Service})
	}

	q, args, err := qBuild.ToSql()
	if err != nil {
		return 0, fmt.Errorf("failed to build sql-query: %w", err)
	}

	var totalCost int64
	if err = r.db.QueryRowContext(ctx, q, args...).Scan(&totalCost); err != nil {
		return 0, fmt.Errorf("failed to get total cost of subcriptions: %w", err)
	}

	return totalCost, nil
}
