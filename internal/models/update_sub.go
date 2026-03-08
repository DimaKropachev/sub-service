package models

import "time"

type UpdateSubscription struct {
	ID      int64
	Price   *int
	EndDate *time.Time
}
