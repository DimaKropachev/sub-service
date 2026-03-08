package models

import (
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID        int64
	Service   string
	Price     int64
	UserID    uuid.UUID
	StartDate time.Time
	EndDate   *time.Time
}

