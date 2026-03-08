package dto

import (
	"github.com/DimaKropachev/sub-service/internal/models"
)

type GetSubscriptionResponse struct {
	ID        int64   `json:"id" example:"123"`
	Service   string  `json:"service_name" example:"Yandex Plus"`
	Price     int64   `json:"price" example:"399"`
	UserID    string  `json:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	StartDate string  `json:"start_date" example:"01-2026"`
	EndDate   *string `json:"end_date,omitempty" example:"07-2026"`
}

func FromDomain(sub models.Subscription) GetSubscriptionResponse {
	dto := GetSubscriptionResponse{
		ID:        sub.ID,
		Service:   sub.Service,
		Price:     sub.Price,
		UserID:    sub.UserID.String(),
		StartDate: sub.StartDate.Format("01-2006"),
	}

	if sub.EndDate != nil {
		formatted := sub.EndDate.Format("01-2006")
		dto.EndDate = &formatted
	}

	return dto
}
