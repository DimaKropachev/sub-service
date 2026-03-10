package models

import "time"

type UpdateSubscription struct {
	ID      int64
	Price   *int64
	EndDate *time.Time
}
