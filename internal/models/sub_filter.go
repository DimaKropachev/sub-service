package models

import (
	"time"

	"github.com/google/uuid"
)

type SubscriptionFilter struct {
	UserID  *uuid.UUID
	Service *string
	From    time.Time
	To      time.Time
}
