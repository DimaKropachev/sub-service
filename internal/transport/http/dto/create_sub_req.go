package dto

import (
	"fmt"
	"time"

	"github.com/DimaKropachev/sub-service/internal/models"
	"github.com/google/uuid"
)

type CreateSubscriptionRequest struct {
	Service   string  `json:"service_name" example:"Yandex Plus"`
	Price     int64   `json:"price" example:"399"`
	UserID    string  `json:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	StartDate string  `json:"start_date" example:"01-2026"`
	EndDate   *string `json:"end_date,omitempty" example:"07-2025"`
}

func (s *CreateSubscriptionRequest) ToDomain() (models.Subscription, error) {
	startDate, err := time.Parse("01-2006", s.StartDate)
	if err != nil {
		return models.Subscription{}, fmt.Errorf("invalid start_date format, expected MM-YYYY")
	}

	var endDate *time.Time
	if s.EndDate != nil {
		parsed, err := time.Parse("01-2006", *s.EndDate)
		if err != nil {
			return models.Subscription{}, fmt.Errorf("invalid end_date format, expected MM-YYYY")
		}
		endDate = &parsed
	}

	userID, err := uuid.Parse(s.UserID)
	if err != nil {
		return models.Subscription{}, fmt.Errorf("invalid UUID: %v", err)
	}

	return models.Subscription{
		Service:   s.Service,
		Price:     s.Price,
		UserID:    userID,
		StartDate: startDate,
		EndDate:   endDate,
	}, nil
}
