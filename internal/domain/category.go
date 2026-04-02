package domain

import (
	"time"
)

type Category struct {
	ID        int64     `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Type      string    `db:"type" json:"type"`
	Icon      *string   `db:"icon" json:"icon,omitempty"`
	ColorHex  *string   `db:"color_hex" json:"color_hex,omitempty"`
	AITags    []string  `db:"ai_tags" json:"ai_tags,omitempty"`
	IsSystem  bool      `db:"is_system" json:"is_system"`
	ParentID  *int64    `db:"parent_id" json:"parent_id,omitempty"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type UserCategory struct {
	ID         int64     `db:"id" json:"id"`
	UserID     string    `db:"user_id" json:"user_id"`
	CategoryID int64     `db:"category_id" json:"category_id"`
	CustomName *string   `db:"custom_name" json:"custom_name,omitempty"`
	CustomIcon *string   `db:"custom_icon" json:"custom_icon,omitempty"`
	IsHidden   bool      `db:"is_hidden" json:"is_hidden"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}
