package dto

import (
	"fmt"
	"time"

	"github.com/DimaKropachev/sub-service/internal/models"
)

type UpdateSubscriptionRequest struct {
	Price   *int64    `json:"price,omitempty" example:"399"`
	EndDate *string `json:"end_date,omitempty" example:"07-2026"`
}

func (s *UpdateSubscriptionRequest) ToDomain(id int64) (models.UpdateSubscription, error) {
	updateSub := models.UpdateSubscription{
		ID: id,
	}

	if s.Price != nil {
		updateSub.Price = s.Price
	}

	var (
		endDate time.Time
		err     error
	)
	if s.EndDate != nil {
		endDate, err = time.Parse("01-2006", *s.EndDate)
		if err != nil {
			return models.UpdateSubscription{}, fmt.Errorf("invalid start_date format, expected MM-YYYY")
		}
		updateSub.EndDate = &endDate
	}

	return updateSub, nil
}
