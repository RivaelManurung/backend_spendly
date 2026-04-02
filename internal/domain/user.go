package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                 uuid.UUID  `db:"id" json:"id"`
	Email              string     `db:"email" json:"email"`
	Name               *string    `db:"name" json:"name,omitempty"`
	AvatarURL          *string    `db:"avatar_url" json:"avatar_url,omitempty"`
	Status             string     `db:"status" json:"status"`
	CurrencyPreference string     `db:"currency_preference" json:"currency_preference"`
	CreatedAt          time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt          *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}
